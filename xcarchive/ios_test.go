package xcarchive

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestNewIosArchive(t *testing.T) {
	sampleArtifactsGitURI := "https://github.com/bitrise-samples/sample-artifacts.git"
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__artifacts__")
	require.NoError(t, err)

	cmd := command.New("git", "clone", sampleArtifactsGitURI, tmpDir)
	require.NoError(t, cmd.Run())

	iosArchivePth := filepath.Join(tmpDir, "archives/ios.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)
	require.Equal(t, iosArchivePth, archive.Path)
	require.Equal(t, filepath.Join(iosArchivePth, "Info.plist"), archive.InfoPlistPath)

	app := archive.Application
	appPth := filepath.Join(iosArchivePth, "Products/Applications/code-sign-test.app")
	require.Equal(t, appPth, app.Path)
	require.Equal(t, filepath.Join(appPth, "Info.plist"), app.InfoPlistPath)
	require.Equal(t, filepath.Join(appPth, "embedded.mobileprovision"), app.ProvisioningProfilePath)
	require.Equal(t, filepath.Join(appPth, "archived-expanded-entitlements.xcent"), app.EntitlementsPath)

	require.Equal(t, 1, len(app.Extensions))
	extension := app.Extensions[0]
	extensionPth := filepath.Join(appPth, "PlugIns/share-extension.appex")
	require.Equal(t, extensionPth, extension.Path)
	require.Equal(t, filepath.Join(extensionPth, "Info.plist"), extension.InfoPlistPath)
	require.Equal(t, filepath.Join(extensionPth, "embedded.mobileprovision"), extension.ProvisioningProfilePath)
	require.Equal(t, filepath.Join(extensionPth, "archived-expanded-entitlements.xcent"), extension.EntitlementsPath)

	require.NotNil(t, app.WatchApplication)
	watchApp := *app.WatchApplication
	watchAppPth := filepath.Join(appPth, "Watch/watchkit-app.app")
	require.Equal(t, watchAppPth, watchApp.Path)
	require.Equal(t, filepath.Join(watchAppPth, "Info.plist"), watchApp.InfoPlistPath)
	require.Equal(t, filepath.Join(watchAppPth, "embedded.mobileprovision"), watchApp.ProvisioningProfilePath)
	require.Equal(t, filepath.Join(watchAppPth, "archived-expanded-entitlements.xcent"), watchApp.EntitlementsPath)

	require.Equal(t, 1, len(watchApp.Extensions))
	watchExtension := watchApp.Extensions[0]
	watchExtensionPth := filepath.Join(watchAppPth, "PlugIns/watchkit-app Extension.appex")
	require.Equal(t, watchExtensionPth, watchExtension.Path)
	require.Equal(t, filepath.Join(watchExtensionPth, "Info.plist"), watchExtension.InfoPlistPath)
	require.Equal(t, filepath.Join(watchExtensionPth, "embedded.mobileprovision"), watchExtension.ProvisioningProfilePath)
	require.Equal(t, filepath.Join(watchExtensionPth, "archived-expanded-entitlements.xcent"), watchExtension.EntitlementsPath)
}
