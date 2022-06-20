package simulator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	mockcommand "github.com/bitrise-io/go-xcode/v2/simulator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func Test_GivenSimulator_WhenBoot_ThenBootsTheRequestedSimulator(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	identifier := "test-identifier"
	parameters := []string{"simctl", "boot", identifier}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.SimulatorBoot(identifier)

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenEnableVerboseLog_ThenEnablesIt(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	identifier := "test-identifier"
	parameters := []string{"simctl", "logverbose", identifier, "enable"}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.SimulatorEnableVerboseLog(identifier)

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenCollectDiagnostics_ThenCollectsIt(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	mocks.commandFactory.On("Create", "xcrun", mock.Anything, mock.Anything).Return(createCommand(""))

	// When
	diagnosticsOutDir, err := manager.SimulatorCollectDiagnostics()
	defer func() {
		// Do not forget to clen up the temp dir
		_ = os.RemoveAll(diagnosticsOutDir)
	}()

	// Then
	assert.NoError(t, err)

	parameters := []string{"simctl", "diagnose", "-b", "--no-archive", fmt.Sprintf("--output=%s", diagnosticsOutDir)}
	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenSimulator_WhenShutdown_ThenShutsItDown(t *testing.T) {
	// Given
	manager, mocks := createSimulatorAndMocks()

	identifier := "test-identifier"
	parameters := []string{"simctl", "shutdown", identifier}
	mocks.commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.SimulatorShutdown(identifier)

	// Then
	assert.NoError(t, err)

	mocks.commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

// Helpers

func createSimulatorAndMocks() (Manager, testingMocks) {
	commandFactory := new(mockcommand.CommandFactory)
	manager := NewManager(commandFactory)

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
