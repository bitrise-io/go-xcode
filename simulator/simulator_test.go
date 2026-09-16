package simulator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination"
	mockcommand "github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testingMocks struct {
	commandFactory *mockcommand.CommandFactory
}

func Test_GivenSimulator_WhenResetLaunchServices_ThenPerformsAction(t *testing.T) {
	// Given
	xcodePath := "/some/path"
	manager, mocks := createSimulatorAndMocks()

	mocks.commandFactory.On("Create", "sw_vers", []string{"-productVersion"}, mock.Anything).Return(createCommand("11.6"))
	mocks.commandFactory.On("Create", "xcode-select", []string{"--print-path"}, mock.Anything).Return(createCommand(xcodePath))

	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	simulatorPath := filepath.Join(xcodePath, "Applications/Simulator.app")
	mocks.commandFactory.On("Create", lsregister, []string{"-f", simulatorPath}, mock.Anything).Return(createCommand(""))

	// When
	err := manager.ResetLaunchServices()

	// Then
	assert.NoError(t, err)
}

// makeFakeXcode lays out an Xcode.app-like tree and returns the developer directory path,
// the same value `xcode-select --print-path` would report.
func makeFakeXcode(t *testing.T, guiAppName string) (devDir string, guiAppPath string) {
	t.Helper()

	contents := filepath.Join(t.TempDir(), "Xcode.app", "Contents")
	devDir = filepath.Join(contents, "Developer")

	switch guiAppName {
	case "Simulator.app":
		guiAppPath = filepath.Join(devDir, "Applications", "Simulator.app")
	case "DeviceHub.app":
		guiAppPath = filepath.Join(contents, "Applications", "DeviceHub.app")
	default:
		guiAppPath = ""
	}

	require.NoError(t, os.MkdirAll(devDir, 0700))
	if guiAppPath != "" {
		require.NoError(t, os.MkdirAll(guiAppPath, 0700))
	}

	return devDir, guiAppPath
}

func Test_GivenSimulatorApp_WhenLaunchWithGUI_ThenOpensSimulatorAppWithoutBooting(t *testing.T) {
	// Given
	devDir, simulatorApp := makeFakeXcode(t, "Simulator.app")
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	openArgs := []string{simulatorApp, "--args", "-CurrentDeviceUDID", identifier}

	mocks.commandFactory.On("Create", "xcode-select", []string{"--print-path"}, mock.Anything).Return(createCommand(devDir))
	mocks.commandFactory.On("Create", "open", openArgs, mock.Anything).Return(createCommand(""))

	// When
	err := manager.LaunchWithGUI(identifier)

	// Then
	assert.NoError(t, err)
	mocks.commandFactory.AssertCalled(t, "Create", "open", openArgs, mock.Anything)
	// Simulator.app boots the device itself, so the step must not issue a separate boot.
	mocks.commandFactory.AssertNotCalled(t, "Create", "xcrun", mock.Anything, mock.Anything)
}

func Test_GivenXcode27_WhenLaunchWithGUI_ThenBootsAndOpensDeviceHub(t *testing.T) {
	// Given: Xcode 27 ships DeviceHub.app instead of Simulator.app.
	devDir, deviceHubApp := makeFakeXcode(t, "DeviceHub.app")
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	bootArgs := []string{"simctl", "boot", identifier}
	openArgs := []string{deviceHubApp, "--args", "-CurrentDeviceUDID", identifier}

	mocks.commandFactory.On("Create", "xcode-select", []string{"--print-path"}, mock.Anything).Return(createCommand(devDir))
	mocks.commandFactory.On("Create", "xcrun", bootArgs, mock.Anything).Return(createCommand(""))
	mocks.commandFactory.On("Create", "open", openArgs, mock.Anything).Return(createCommand(""))

	// When
	err := manager.LaunchWithGUI(identifier)

	// Then
	assert.NoError(t, err)
	// DeviceHub ignores -CurrentDeviceUDID, so the device has to be booted explicitly.
	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", bootArgs, mock.Anything)
	mocks.commandFactory.AssertCalled(t, "Create", "open", openArgs, mock.Anything)
}

func Test_GivenNoGUIApp_WhenLaunchWithGUI_ThenFailsNamingBothPaths(t *testing.T) {
	// Given
	devDir, _ := makeFakeXcode(t, "")
	manager, mocks := createSimulatorAndMocks()

	mocks.commandFactory.On("Create", "xcode-select", []string{"--print-path"}, mock.Anything).Return(createCommand(devDir))

	// When
	err := manager.LaunchWithGUI("test-identifier")

	// Then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Simulator.app")
	assert.Contains(t, err.Error(), "DeviceHub.app")
	mocks.commandFactory.AssertNotCalled(t, "Create", "open", mock.Anything, mock.Anything)
}

func Test_GivenSimulator_WhenBoot_ThenBootsTheRequestedSimulator(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	parameters := []string{"simctl", "boot", identifier}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.Boot(destination.Device{UDID: identifier})

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenBootRosetta_ThenBootsTheRequestedSimulator(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	parameters := []string{"simctl", "boot", identifier, "--arch=x86_64"}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.Boot(destination.Device{UDID: identifier, Arch: "x86_64"})

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenWaitForBootFinishedTimesOut_ThenFails(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	parameters := []string{"simctl", "launch", identifier, "com.apple.Preferences"}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createTimeoutCommand(time.Hour))

	// When
	err := manager.WaitForBootFinished(identifier, 3*time.Second)

	// Then
	assert.ErrorContains(t, err, "failed to boot Simulator in")

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenEnableVerboseLog_ThenEnablesIt(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	parameters := []string{"simctl", "logverbose", identifier, "enable"}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.EnableVerboseLog(identifier)

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenCollectDiagnostics_ThenCollectsIt(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	mocks.commandFactory.On("Create", "xcrun", mock.Anything, mock.Anything).Return(createCommand(""))

	// When
	diagnosticsOutDir, err := manager.CollectDiagnostics()

	// Then
	assert.NoError(t, err)

	parameters := []string{"simctl", "diagnose", "-b", "--no-archive", fmt.Sprintf("--output=%s", diagnosticsOutDir)}
	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenShutdown_ThenShutsItDown(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	parameters := []string{"simctl", "shutdown", identifier}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.Shutdown(identifier)

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenErase_ThenErases(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	const identifier = "test-identifier"
	parameters := []string{"simctl", "erase", identifier}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.Erase(identifier)

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

// Helpers

func createSimulatorAndMocks() (Manager, testingMocks) {
	commandFactory := new(mockcommand.CommandFactory)
	logger := log.NewLogger()
	manager := NewManager(logger, commandFactory)

	return manager, testingMocks{
		commandFactory: commandFactory,
	}
}

func createCommand(output string) *mockcommand.Command {
	command := new(mockcommand.Command)
	command.On("PrintableCommandArgs").Return("")
	command.On("Run").Return(nil)
	command.On("RunAndReturnExitCode").Return(0, nil)
	command.On("RunAndReturnTrimmedCombinedOutput").Return(output, nil)

	return command
}

func createTimeoutCommand(timeout time.Duration) *mockcommand.Command {
	command := new(mockcommand.Command)
	command.On("PrintableCommandArgs").Return("")
	command.On("Run").Return(func() error {
		time.Sleep(timeout)
		return nil
	})

	return command
}
