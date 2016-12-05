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

		sdk, err := GetBuildConfigSDKRoot(pbxprojPth)
		require.NoError(t, err)
		require.Equal(t, "iphoneos", sdk)
	}

	t.Log("macos")
	{
		pbxprojPth, err := testMacOSPbxprojPth()
		require.NoError(t, err)

		sdk, err := GetBuildConfigSDKRoot(pbxprojPth)
		require.NoError(t, err)
		require.Equal(t, "macosx", sdk)
	}
}
