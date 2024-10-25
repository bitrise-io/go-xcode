package spaceship

import (
	"errors"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/mock"
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

	cmdFactory := new(mocks.RubyCommandFactory)
	cmdFactory.On("CreateBundleExec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd)

	c := &Client{
		cmdFactory:     cmdFactory,
		isNoSleepRetry: true,
	}
	out, err := c.runSpaceshipCommand("")
	require.NoError(t, err)
	require.Equal(t, "{}", out)

	cmd.AssertExpectations(t)
	cmdFactory.AssertExpectations(t)
}

func Test_runSpaceshipCommand_retries_only_temporarily_unavailable_error(t *testing.T) {
	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return("exit status 1", errors.New("exit status 1")).Once()

	cmdFactory := new(mocks.RubyCommandFactory)
	cmdFactory.On("CreateBundleExec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd)

	c := &Client{
		cmdFactory: cmdFactory,
	}
	out, err := c.runSpaceshipCommand("")
	require.EqualError(t, err, "spaceship command failed with output: exit status 1")
	require.Equal(t, "", out)

	cmd.AssertExpectations(t)
	cmdFactory.AssertExpectations(t)
}
