package xcodecommand

import (
	"testing"

	gocommand "github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	mockcommand "github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

type testingMocks struct {
	command  *mockcommand.Command
	xcpretty *mockXcprettyManager
}

func Test_GivenNotInstalled_WhenInstall_ThenInstallsIt(t *testing.T) {
	// Given
	installer, version, mocks := createInstallerAndMocks(t, false)

	// When
	installedVersion, err := installer.CheckInstall()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, version, installedVersion)
	mocks.xcpretty.AssertCalled(t, "isDepInstalled")
	mocks.xcpretty.AssertCalled(t, "installDep")
	mocks.xcpretty.AssertCalled(t, "depVersion")
	mocks.command.AssertCalled(t, "Run")
}

func Test_GivenInstalled_WhenInstall_OnlyReturnsVersion(t *testing.T) {
	// Given
	installer, version, mocks := createInstallerAndMocks(t, true)

	// When
	installedVersion, err := installer.CheckInstall()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, version, installedVersion)
	mocks.xcpretty.AssertCalled(t, "isDepInstalled")
	mocks.xcpretty.AssertNotCalled(t, "installDep")
	mocks.xcpretty.AssertCalled(t, "depVersion")
	mocks.command.AssertNotCalled(t, "Run")
}

func createInstallerAndMocks(t *testing.T, installed bool) (Runner, *version.Version, testingMocks) {
	command := new(mockcommand.Command)
	command.On("Run").Return(nil)

	version, _ := version.NewVersion("1.0.0")

	mockxcpretty := newMockXcprettyManager(t)
	mockxcpretty.On("isDepInstalled").Return(installed, nil)
	if !installed {
		mockxcpretty.On("installDep").Return([]gocommand.Command{command}, nil)
	}
	mockxcpretty.On("depVersion").Return(version, nil)

	installer := &XcprettyCommandRunner{
		logger:   log.NewLogger(),
		xcpretty: mockxcpretty,
	}

	return installer, version, testingMocks{
		command:  command,
		xcpretty: mockxcpretty,
	}
}
