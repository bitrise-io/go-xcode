package xcodebuild

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newProviderWithCommand(t *testing.T, out string, runErr error) (BuildSettingsProvider, *mocks.CommandFactory) {
	t.Helper()

	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return(out, runErr)
	cmd.On("PrintableCommandArgs").Return("xcodebuild ...").Maybe()

	factory := new(mocks.CommandFactory)
	factory.On("Create", toolName, mock.Anything, mock.Anything).Return(cmd)

	return NewShowBuildSettingsProvider(factory, log.NewLogger()), factory
}

func TestShowBuildSettingsProvider_TargetBuildSettings(t *testing.T) {
	provider, factory := newProviderWithCommand(t, "    PRODUCT_BUNDLE_IDENTIFIER = io.bitrise.App", nil)

	settings, err := provider.TargetBuildSettings("/p/App.xcodeproj", "App", "Release", "-destination", "generic/platform=iOS")
	require.NoError(t, err)

	assert.Equal(t, serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.App"}, settings)

	factory.AssertCalled(t, "Create", toolName, []string{
		"-project", "/p/App.xcodeproj",
		"-target", "App",
		"-configuration", "Release",
		"-showBuildSettings",
		"-destination", "generic/platform=iOS",
	}, (*command.Opts)(nil))
}

func TestShowBuildSettingsProvider_SchemeBuildSettings(t *testing.T) {
	provider, factory := newProviderWithCommand(t, "    SDKROOT = iphoneos18.0", nil)

	settings, err := provider.SchemeBuildSettings("/p/App.xcworkspace", "App", "Debug")
	require.NoError(t, err)

	assert.Equal(t, serialized.Object{"SDKROOT": "iphoneos18.0"}, settings)

	factory.AssertCalled(t, "Create", toolName, []string{
		"-workspace", "/p/App.xcworkspace",
		"-scheme", "App",
		"-configuration", "Debug",
		"-showBuildSettings",
	}, (*command.Opts)(nil))
}

func TestShowBuildSettingsProvider_exitStatusErrorReportsOutput(t *testing.T) {
	provider, _ := newProviderWithCommand(t, "xcodebuild: error: The project named \"App\" does not contain a target named \"Nope\"", &exec.ExitError{})

	_, err := provider.TargetBuildSettings("/p/App.xcodeproj", "Nope", "Debug")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain a target named")
	assert.Contains(t, err.Error(), "exit status")
	assert.Equal(t, 1, strings.Count(err.Error(), "xcodebuild ..."), "the command appears once")

	var exitErr *exec.ExitError
	assert.ErrorAs(t, err, &exitErr)
}

func TestShowBuildSettingsProvider_otherErrorIsWrapped(t *testing.T) {
	notFound := errors.New("executable file not found in $PATH")
	provider, _ := newProviderWithCommand(t, "", notFound)

	_, err := provider.TargetBuildSettings("/p/App.xcodeproj", "App", "Debug")

	require.Error(t, err)
	assert.ErrorIs(t, err, notFound)
}
