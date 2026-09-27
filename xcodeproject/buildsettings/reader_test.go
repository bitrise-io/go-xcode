package buildsettings

import (
	"errors"
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
func newReader(stdout, stderr string, runErr error) (Reader, *mocks.CommandFactory) {
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

func TestReader_Read(t *testing.T) {
	tests := []struct {
		name  string
		query Query
		args  []string
	}{
		{
			name:  "target",
			query: Query{ProjectPath: "App.xcodeproj", Target: "App", Configuration: "Release", AdditionalOptions: []string{"COMPILER_INDEX_STORE_ENABLE=NO"}},
			args:  []string{"-project", "App.xcodeproj", "-target", "App", "-configuration", "Release", "-showBuildSettings", "-json", "COMPILER_INDEX_STORE_ENABLE=NO"},
		},
		{
			name:  "scheme in a workspace",
			query: Query{ProjectPath: "App.xcworkspace", Scheme: "App"},
			args:  []string{"-workspace", "App.xcworkspace", "-scheme", "App", "-showBuildSettings", "-json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, factory := newReader(`[{"action": "build", "target": "App", "buildSettings": {"SDKROOT": "iphoneos"}}]`, warnings, nil)

			list, err := r.Read(tt.query)
			require.NoError(t, err)
			require.Equal(t, List{{Target: "App", Action: "build", Values: map[string]string{"SDKROOT": "iphoneos"}}}, list, "stderr warnings stay out of the JSON")
			factory.AssertCalled(t, "Create", "xcodebuild", tt.args, mock.Anything)
		})
	}
}

func TestReader_Read_failedRunReportsOutput(t *testing.T) {
	r, _ := newReader("", `xcodebuild: error: The project named "App" does not contain a target named "Nope".`, &exec.ExitError{})

	_, err := r.Read(Query{ProjectPath: "App.xcodeproj", Target: "Nope"})
	require.ErrorContains(t, err, `does not contain a target named "Nope"`)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
}

func TestReader_Read_errors(t *testing.T) {
	notFound := errors.New("executable file not found in $PATH")
	r, _ := newReader("", "", notFound)
	_, err := r.Read(Query{ProjectPath: "App.xcodeproj", Target: "App"})
	require.ErrorIs(t, err, notFound)

	r, _ = newReader("Build settings for action build and target App:", "", nil)
	_, err = r.Read(Query{ProjectPath: "App.xcodeproj", Target: "App"})
	require.ErrorContains(t, err, "printed no build settings JSON")
}
