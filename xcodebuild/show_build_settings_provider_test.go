package xcodebuild

import (
	"errors"
	"io"
	"os"
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

// newProviderWithCommand fakes an xcodebuild run that writes stdout and stderr separately.
func newProviderWithCommand(t *testing.T, stdout, stderr string, runErr error) (BuildSettingsProvider, *mocks.CommandFactory) {
	t.Helper()

	var opts *command.Opts
	cmd := new(mocks.Command)
	cmd.On("Run").Run(func(mock.Arguments) {
		_, _ = io.WriteString(opts.Stdout, stdout)
		_, _ = io.WriteString(opts.Stderr, stderr)
	}).Return(runErr)
	cmd.On("PrintableCommandArgs").Return("xcodebuild ...").Maybe()

	factory := new(mocks.CommandFactory)
	factory.On("Create", toolName, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		opts = args.Get(2).(*command.Opts)
	}).Return(cmd)

	return NewShowBuildSettingsProvider(factory, log.NewLogger()), factory
}

const xcodebuildWarnings = `2026-09-27 10:00:00.000 xcodebuild[1:2] [MT] IDERunDestination: Supported platforms for the buildables in the current scheme is empty.
--- xcodebuild: WARNING: Using the first of multiple matching destinations:
{ platform:macOS, arch:arm64, id:0000, name:My Mac }`

func TestShowBuildSettingsProvider_TargetBuildSettings(t *testing.T) {
	provider, factory := newProviderWithCommand(t,
		`[{"action": "build", "target": "App", "buildSettings": {"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.App"}}]`, xcodebuildWarnings, nil)

	settings, err := provider.TargetBuildSettings("/p/App.xcodeproj", "App", "Release", "-destination", "generic/platform=iOS")
	require.NoError(t, err)

	assert.Equal(t, serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.App"}, settings, "stderr warnings do not reach the JSON")

	factory.AssertCalled(t, "Create", toolName, []string{
		"-project", "/p/App.xcodeproj",
		"-target", "App",
		"-configuration", "Release",
		"-showBuildSettings", "-json",
		"-destination", "generic/platform=iOS",
	}, mock.Anything)
}

func TestShowBuildSettingsProvider_SchemeBuildSettings(t *testing.T) {
	out, err := os.ReadFile("./testdata/showBuildSettingsSchemeWithTestTarget.json")
	require.NoError(t, err)
	provider, factory := newProviderWithCommand(t, string(out), xcodebuildWarnings, nil)

	settings, err := provider.SchemeBuildSettings("/p/App.xcworkspace", "App", "Debug")
	require.NoError(t, err)

	assert.Equal(t, "Bitrise.ios-simple-objc", settings["PRODUCT_BUNDLE_IDENTIFIER"], "the scheme's main target")
	assert.NotContains(t, settings, "BUNDLE_LOADER", "the test target's own settings are not mixed in")

	factory.AssertCalled(t, "Create", toolName, []string{
		"-workspace", "/p/App.xcworkspace",
		"-scheme", "App",
		"-configuration", "Debug",
		"-showBuildSettings", "-json",
	}, mock.Anything)
}

// An Xcode that prints text despite -json still works through the text parser.
func TestShowBuildSettingsProvider_textFallback(t *testing.T) {
	provider, _ := newProviderWithCommand(t, "Build settings for action build and target App:\n    SDKROOT = iphoneos18.0", "", nil)

	settings, err := provider.TargetBuildSettings("/p/App.xcodeproj", "App", "Release")
	require.NoError(t, err)
	assert.Equal(t, serialized.Object{"SDKROOT": "iphoneos18.0"}, settings)
}

func TestShowBuildSettingsProvider_exitStatusErrorReportsOutput(t *testing.T) {
	provider, _ := newProviderWithCommand(t, "", "xcodebuild: error: The project named \"App\" does not contain a target named \"Nope\"", &exec.ExitError{})

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
	provider, _ := newProviderWithCommand(t, "", "", notFound)

	_, err := provider.TargetBuildSettings("/p/App.xcodeproj", "App", "Debug")

	require.Error(t, err)
	assert.ErrorIs(t, err, notFound)
}
