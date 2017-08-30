package xcodeproj

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func cloneSampleProject(t *testing.T, url, projectPth string) string {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__codesignproperties__")
	require.NoError(t, err)

	cmd := command.New("git", "clone", url, tmpDir)
	require.NoError(t, cmd.Run())

	return filepath.Join(tmpDir, projectPth)
}

func TestTargetCodeSignMapping(t *testing.T) {
	{
		projectPth := cloneSampleProject(t, "https://github.com/bitrise-samples/sample-apps-ios-multi-target.git", "code-sign-test.xcodeproj")

		mapping, err := TargetCodeSignMapping(projectPth)
		require.NoError(t, err)
		require.Equal(t, 4, len(mapping))

		{
			properties, ok := mapping["watchkit-app Extension"]
			require.True(t, ok)
			require.Equal(t, "com.bitrise.code-sign-test.watchkitapp.watchkitextension", properties.BundleIdentifier)
			require.Equal(t, "Automatic", properties.ProvisioningStyle)
			require.Equal(t, "iPhone Developer", properties.CodeSignIdentity)
			require.Equal(t, "", properties.ProvisioningProfile)
			require.Equal(t, "", properties.ProvisioningProfileSpecifier)
		}

		{
			properties, ok := mapping["code-sign-test"]
			require.True(t, ok)
			require.Equal(t, "com.bitrise.code-sign-test", properties.BundleIdentifier)
			require.Equal(t, "Automatic", properties.ProvisioningStyle)
			require.Equal(t, "iPhone Developer", properties.CodeSignIdentity)
			require.Equal(t, "", properties.ProvisioningProfile)
			require.Equal(t, "", properties.ProvisioningProfileSpecifier)
		}

		{
			properties, ok := mapping["share-extension"]
			require.True(t, ok)
			require.Equal(t, "com.bitrise.code-sign-test.share-extension", properties.BundleIdentifier)
			require.Equal(t, "Automatic", properties.ProvisioningStyle)
			require.Equal(t, "iPhone Developer", properties.CodeSignIdentity)
			require.Equal(t, "", properties.ProvisioningProfile)
			require.Equal(t, "", properties.ProvisioningProfileSpecifier)
		}

		{
			properties, ok := mapping["watchkit-app"]
			require.True(t, ok)
			require.Equal(t, "com.bitrise.code-sign-test.watchkitapp", properties.BundleIdentifier)
			require.Equal(t, "Automatic", properties.ProvisioningStyle)
			require.Equal(t, "iPhone Developer", properties.CodeSignIdentity)
			require.Equal(t, "", properties.ProvisioningProfile)
			require.Equal(t, "", properties.ProvisioningProfileSpecifier)
		}
	}

	{
		projectPth := cloneSampleProject(t, "https://github.com/bitrise-samples/sample-apps-ios-simple-objc.git", "ios-simple-objc/ios-simple-objc.xcodeproj")

		mapping, err := TargetCodeSignMapping(projectPth)
		require.NoError(t, err)
		require.Equal(t, 1, len(mapping))

		{
			properties, ok := mapping["ios-simple-objc"]
			require.True(t, ok)
			require.Equal(t, "Bitrise.ios-simple-objc", properties.BundleIdentifier)
			require.Equal(t, "Manual", properties.ProvisioningStyle)
			require.Equal(t, "iPhone Developer", properties.CodeSignIdentity)
			require.Equal(t, "", properties.ProvisioningProfile)
			require.Equal(t, "BitriseBot-Wildcard", properties.ProvisioningProfileSpecifier)
		}
	}
}
