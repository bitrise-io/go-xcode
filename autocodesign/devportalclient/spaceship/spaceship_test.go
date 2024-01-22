package spaceship

import (
	"errors"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/require"
)

const spaceshipTemporaryUnavailableOutput = `Apple ID authentication failed: <html>
<head><title>503 Service Temporarily Unavailable</title></head>
<body>
<center><h1>503 Service Temporarily Unavailable</h1></center>
<hr><center>Apple</center>
</body>
</html>`

func Test_runSpaceshipCommand_retries_on_temporarily_unavailable_error(t *testing.T) {
	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return(spaceshipTemporaryUnavailableOutput, errors.New("exit status 1")).Once()
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return(spaceshipTemporaryUnavailableOutput, errors.New("exit status 1")).Once()
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return("{}", nil).Once()

	spaceshipCmd := spaceshipCommand{
		command:              cmd,
		printableCommandArgs: "mock",
	}

	out, err := runSpaceshipCommand(spaceshipCmd)
	require.NoError(t, err)
	require.Equal(t, "{}", out)

	cmd.AssertExpectations(t)
}

func Test_runSpaceshipCommand_retries_only_temporarily_unavailable_error(t *testing.T) {
	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return("exit status 1", errors.New("exit status 1")).Once()

	spaceshipCmd := spaceshipCommand{
		command:              cmd,
		printableCommandArgs: "mock",
	}

	out, err := runSpaceshipCommand(spaceshipCmd)
	require.EqualError(t, err, "spaceship command failed with output: exit status 1")
	require.Equal(t, "", out)

	cmd.AssertExpectations(t)
}
