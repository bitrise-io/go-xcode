package xcodecommand

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/stretchr/testify/require"
)

func TestCommand_ArgsReturnsACopy(t *testing.T) {
	cmd, err := Archive(ArchiveParams{ProjectPath: "App.xcodeproj"})
	require.NoError(t, err)

	args := cmd.Args()
	args[0] = "build"

	require.Equal(t, "archive", cmd.Args()[0])
}

type recordingFactory struct {
	name string
	args []string
	opts *command.Opts
}

func (f *recordingFactory) Create(name string, args []string, opts *command.Opts) command.Command {
	f.name, f.args, f.opts = name, args, opts
	return nil
}

func TestCommand_Create(t *testing.T) {
	cmd, err := Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Scheme: "App"})
	require.NoError(t, err)
	opts := &command.Opts{Dir: "/work"}
	factory := &recordingFactory{}

	cmd.Create(factory, opts)

	require.Equal(t, "xcodebuild", factory.name)
	require.Equal(t, cmd.Args(), factory.args)
	require.Same(t, opts, factory.opts)
}

func TestContainerOptions_trailingSlash(t *testing.T) {
	cmd, err := Build(BuildParams{ProjectPath: "App.xcworkspace/"})
	require.NoError(t, err)
	require.Equal(t, []string{"build", "-workspace", "App.xcworkspace/"}, cmd.Args())
}
