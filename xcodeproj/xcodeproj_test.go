package xcodeproj

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBuildConfigSDKRoot(t *testing.T) {
	t.Log("ios")
	{
		pbxprojPth, err := testIOSPbxprojPth()
		require.NoError(t, err)

		sdks, err := GetBuildConfigSDKs(pbxprojPth)
		require.NoError(t, err)
		require.Equal(t, []string{"iphoneos"}, sdks)
	}

	t.Log("macos")
	{
		pbxprojPth, err := testMacOSPbxprojPth()
		require.NoError(t, err)

		sdks, err := GetBuildConfigSDKs(pbxprojPth)
		require.NoError(t, err)
		require.Equal(t, []string{"macosx"}, sdks)
	}
}
