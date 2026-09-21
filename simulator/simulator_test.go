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

func Test_GivenXcode_WhenLaunchWithGUI_ThenOpensTheAvailableGUIApp(t *testing.T) {
	const udid = "test-udid"

	tests := []struct {
		name    string
		guiApp  string // relative to Xcode.app/Contents, empty if none is installed
		wantErr bool
	}{
		{name: "Xcode 26: Simulator.app", guiApp: "Developer/Applications/Simulator.app"},
		{name: "Xcode 27: DeviceHub.app", guiApp: "Applications/DeviceHub.app"},
		{name: "no GUI app", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents := filepath.Join(t.TempDir(), "Xcode.app", "Contents")
			devDir := filepath.Join(contents, "Developer")
			require.NoError(t, os.MkdirAll(devDir, 0700))

			appPath := ""
			if tt.guiApp != "" {
				appPath = filepath.Join(contents, tt.guiApp)
				require.NoError(t, os.MkdirAll(appPath, 0700))
			}
			openArgs := []string{appPath, "--args", "-CurrentDeviceUDID", udid}

			manager, mocks := createSimulatorAndMocks()
			mocks.commandFactory.On("Create", "xcode-select", []string{"--print-path"}, mock.Anything).Return(createCommand(devDir))
			mocks.commandFactory.On("Create", "open", openArgs, mock.Anything).Return(createCommand(""))

			err := manager.LaunchWithGUI(udid)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "Simulator.app")
				assert.Contains(t, err.Error(), "DeviceHub.app")
				mocks.commandFactory.AssertNotCalled(t, "Create", "open", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			mocks.commandFactory.AssertCalled(t, "Create", "open", openArgs, mock.Anything)
		})
	}
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
