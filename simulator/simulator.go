package simulator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
)

// Manager provides methods for issuing Simulator commands
type Manager interface {
	LaunchWithGUI(simulatorID string) error

	ResetLaunchServices() error
	Boot(id string) error
	WaitForBootFinished(id string, timeout time.Duration) error
	EnableVerboseLog(id string) error
	CollectDiagnostics() (string, error)
	Shutdown(id string) error
	Erase(id string) error
}

type manager struct {
	commandFactory command.Factory
	logger         log.Logger
}

// NewManager ...
func NewManager(logger log.Logger, commandFactory command.Factory) Manager {
	return manager{
		logger:         logger,
		commandFactory: commandFactory,
	}
}

func (m manager) getSimulatorAppAbsolutePath() (string, error) {
	cmd := m.commandFactory.Create("xcode-select", []string{"--print-path"}, nil)
	xcodeDevDirPath, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Xcode Developer Directory - most likely Xcode.app is not installed: %w", err)
	}

	return filepath.Join(xcodeDevDirPath, "Applications", "Simulator.app"), nil
}

// LaunchWithGUI can be used to run in non-headless mode (with the Simulator visible).
func (m manager) LaunchWithGUI(simulatorID string) error {
	simulatorAppFullPath, err := m.getSimulatorAppAbsolutePath()
	if err != nil {
		return err
	}

	openCmd := m.commandFactory.Create("open", []string{simulatorAppFullPath, "--args", "-CurrentDeviceUDID", simulatorID}, nil)

	m.logger.Printf("$ %s", openCmd.PrintableCommandArgs())
	outStr, err := openCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start simulators (%s), error: %s, output: %s", simulatorID, err, outStr)
	}

	return nil
}

// ResetLaunchServices resets launch services database to avoid Big Sur's sporadic failure to find the Simulator App
// The following error is printed when this happens: "kLSNoExecutableErr: The executable is missing"
// Details:
// - https://stackoverflow.com/questions/2182040/the-application-cannot-be-opened-because-its-executable-is-missing/16546673#16546673
// - https://ss64.com/osx/lsregister.html
func (m manager) ResetLaunchServices() error {
	cmd := m.commandFactory.Create("sw_vers", []string{"-productVersion"}, nil)

	macOSVersion, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return err
	}

	if strings.HasPrefix(macOSVersion, "11.") { // It's Big Sur
		simulatorAppPath, err := m.getSimulatorAppAbsolutePath()
		if err != nil {
			return err
		}

		cmdString := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
		cmd = m.commandFactory.Create(cmdString, []string{"-f", simulatorAppPath}, nil)

		m.logger.Infof("Applying launch services reset workaround before booting simulator")
		_, err = cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			return err
		}
	}

	return nil
}

// Boot boots Simulator in headless mode
func (m manager) Boot(id string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "boot", id}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	m.logger.Donef("$ %s", cmd.PrintableCommandArgs())
	exitCode, err := cmd.RunAndReturnExitCode()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			if exitCode == 149 { // Simulator already booted
				return nil
			}
			m.logger.Warnf("Failed to boot Simulator, command exited with code %d", exitCode)
			return nil
		}
		return fmt.Errorf("failed to boot Simulator, command execution failed: %v", err)
	}

	return nil
}

// WaitForBootFinished waits until simulator finsihed boot
func (m manager) WaitForBootFinished(id string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer func() {
		timer.Stop()
	}()

	launchDoneCh := make(chan error, 1)
	doWait := func() {
		waitCmd := m.commandFactory.Create("xcrun", []string{"simctl", "launch", id, "com.apple.Preferences"}, &command.Opts{
			Stdout: os.Stderr,
			Stderr: os.Stderr,
		})

		m.logger.Println()
		m.logger.TDonef("$ %s", waitCmd.PrintableCommandArgs())
		launchDoneCh <- waitCmd.Run()
	}

	go doWait()

	for {
		select {
		case err := <-launchDoneCh:
			{
				if err != nil {
					return fmt.Errorf("failed to wait for simulator boot: %w", err)
				}
				return nil // launch succeeded
			}
		case <-timer.C:
			return fmt.Errorf("failed to boot Simulator in %s", timeout)
		}
	}
}

// EnableVerboseLog enables verbose log. Simulator needs to be booted.
func (m manager) EnableVerboseLog(id string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "logverbose", id, "enable"}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	m.logger.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			m.logger.Warnf("Failed to enable Simulator verbose logging, command exited with code %d", err)
			return nil
		}

		return fmt.Errorf("failed to enable Simulator verbose logging, command execution failed: %v", err)
	}

	return nil
}

// CollectDiagnostics collects Simulator diagnostics
func (m manager) CollectDiagnostics() (string, error) {
	diagnosticsName, err := m.diagnosticsName()
	if err != nil {
		return "", err
	}
	diagnosticsOutDir, err := ioutil.TempDir("", diagnosticsName)
	if err != nil {
		return "", fmt.Errorf("failed to collect Simulator diagnostics, could not create temporary directory: %v", err)
	}

	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "diagnose", "-b", "--no-archive", fmt.Sprintf("--output=%s", diagnosticsOutDir)}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  bytes.NewReader([]byte("\n")),
	})

	m.logger.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", fmt.Errorf("failed to collect Simulator diagnostics: %v", err)

		}
		return "", fmt.Errorf("failed to collect Simulator diagnostics, command execution failed: %v", err)
	}

	return diagnosticsOutDir, nil
}

// Shutdown shuts down the Simulator
func (m manager) Shutdown(id string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "shutdown", id}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	m.logger.Donef("$ %s", cmd.PrintableCommandArgs())
	exitCode, err := cmd.RunAndReturnExitCode()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			if exitCode == 149 { // Simulator already shut down
				return nil
			}
			m.logger.Warnf("Failed to shutdown Simulator, command exited with code %d", exitCode)
			return nil
		}
		return fmt.Errorf("failed to shutdown Simulator, command execution failed: %v", err)
	}

	return nil
}

// Erase erases Simulator content
func (m manager) Erase(id string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "erase", id}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	m.logger.Donef("$ %s", cmd.PrintableCommandArgs())
	exitCode, err := cmd.RunAndReturnExitCode()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			m.logger.Warnf("Failed to erase Simulator, command exited with code %d", exitCode)
			return nil
		}
		return fmt.Errorf("failed to erase Simulator, command execution failed: %v", err)
	}

	return nil
}

func (m manager) diagnosticsName() (string, error) {
	timestamp, err := time.Now().MarshalText()
	if err != nil {
		return "", fmt.Errorf("failed to marshal timestamp: %w", err)
	}

	return fmt.Sprintf("simctl_diagnose_%s.zip", strings.ReplaceAll(string(timestamp), ":", "-")), nil
}
