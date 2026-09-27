package buildsettings

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newReader fakes an xcodebuild run that writes stdout and stderr separately.
func newReader(t *testing.T, stdout, stderr string, runErr error) (Reader, *mocks.CommandFactory) {
	t.Helper()
	var opts *command.Opts
	cmd := new(mocks.Command)
	cmd.On("Run").Run(func(mock.Arguments) {
		_, _ = io.WriteString(opts.Stdout, stdout)
		_, _ = io.WriteString(opts.Stderr, stderr)
	}).Return(runErr)
	cmd.On("PrintableCommandArgs").Return("xcodebuild -showBuildSettings -json").Maybe()

	factory := new(mocks.CommandFactory)
	factory.On("Create", "xcodebuild", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		opts = args.Get(2).(*command.Opts)
	}).Return(cmd)

	return NewReader(factory, log.NewLogger()), factory
}

const warnings = `--- xcodebuild: WARNING: Using the first of multiple matching destinations:
{ platform:macOS, arch:arm64, id:0000, name:My Mac }`

func TestReader_Target(t *testing.T) {
	r, factory := newReader(t, `[{"action": "build", "target": "App", "buildSettings": {"SDKROOT": "iphoneos"}}]`, warnings, nil)

	settings, err := r.Target("App.xcodeproj", "App", "Release", "COMPILER_INDEX_STORE_ENABLE=NO")
	require.NoError(t, err)
	require.Equal(t, Settings{"SDKROOT": "iphoneos"}, settings, "stderr warnings stay out of the JSON")
	factory.AssertCalled(t, "Create", "xcodebuild", []string{"-project", "App.xcodeproj", "-target", "App", "-configuration", "Release", "-showBuildSettings", "-json", "COMPILER_INDEX_STORE_ENABLE=NO"}, mock.Anything)
}

// A scheme's targets are keyed by name, so the result does not depend on the order of the
// scheme's build entries.
func TestReader_SchemeTarget(t *testing.T) {
	out := readFixture(t, "scheme_with_test_target.json")
	r, factory := newReader(t, string(out), warnings, nil)

	settings, err := r.SchemeTarget("App.xcworkspace", "App", "ios-simple-objc", "Debug")
	require.NoError(t, err)
	require.Equal(t, "Bitrise.ios-simple-objc", settings["PRODUCT_BUNDLE_IDENTIFIER"])
	factory.AssertCalled(t, "Create", "xcodebuild", []string{"-workspace", "App.xcworkspace", "-scheme", "App", "-configuration", "Debug", "-showBuildSettings", "-json"}, mock.Anything)

	_, err = r.SchemeTarget("App.xcworkspace", "App", "Nope", "Debug")
	require.EqualError(t, err, "no build settings for target Nope in scheme App; it has settings for: ios-simple-objc, ios-simple-objcTests")
}

func TestReader_Scheme(t *testing.T) {
	r, _ := newReader(t, string(readFixture(t, "scheme_with_test_target.json")), "", nil)

	targets, err := r.Scheme("App.xcodeproj", "App", "")
	require.NoError(t, err)
	require.Len(t, targets, 2)
	require.Contains(t, targets, "ios-simple-objcTests")
}

func TestReader_rejectsPathsWithoutBuildSettings(t *testing.T) {
	r, factory := newReader(t, "[]", "", nil)

	_, err := r.Scheme("MyPackage/Package.swift", "MyPackage", "")
	require.ErrorContains(t, err, "needs an .xcodeproj or .xcworkspace, got MyPackage/Package.swift")
	_, err = r.SchemeTarget("MyPackage", "MyPackage", "MyPackage", "")
	require.ErrorContains(t, err, "needs an .xcodeproj or .xcworkspace")
	_, err = r.Target("App.xcworkspace", "App", "")
	require.ErrorContains(t, err, "reading a target's build settings needs an .xcodeproj, got App.xcworkspace")
	factory.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestReader_errors(t *testing.T) {
	r, _ := newReader(t, "", `xcodebuild: error: The project named "App" does not contain a target named "Nope".`, &exec.ExitError{})
	_, err := r.Target("App.xcodeproj", "Nope", "")
	require.ErrorContains(t, err, `does not contain a target named "Nope"`)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)

	notFound := errors.New("executable file not found in $PATH")
	r, _ = newReader(t, "", "", notFound)
	_, err = r.Target("App.xcodeproj", "App", "")
	require.ErrorIs(t, err, notFound)

	r, _ = newReader(t, "Build settings for action build and target App:", "", nil)
	_, err = r.Target("App.xcodeproj", "App", "")
	require.ErrorContains(t, err, "printed no build settings JSON")
}

// debugRecorder keeps what is logged at debug level.
type debugRecorder struct {
	log.Logger
	debug []string
}

func (l *debugRecorder) Debugf(format string, v ...interface{}) {
	l.debug = append(l.debug, fmt.Sprintf(format, v...))
}

// stderr carries xcodebuild's warnings: logged at debug level, kept out of the JSON.
func TestReader_logsStderrAtDebugLevel(t *testing.T) {
	_, factory := newReader(t, `[{"target": "App", "buildSettings": {}}]`, warnings, nil)
	logger := &debugRecorder{Logger: log.NewLogger()}

	_, err := NewReader(factory, logger).Target("App.xcodeproj", "App", "")
	require.NoError(t, err)
	require.Equal(t, []string{"xcodebuild stderr:\n" + warnings}, logger.debug)
}
