package utility

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetXcodeVersion(t *testing.T) {
	t.Log("GetXcodeVersion")
	{
		testModel, err := GetXcodeVersion()
		require.NoError(t, err, testModel.Version)
	}
}
