package xcarchive

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestNewMacosArchive(t *testing.T) {
	sampleArtifactsGitURI := "https://github.com/bitrise-samples/sample-artifacts.git"
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__artifacts__")
	require.NoError(t, err)

	cmd := command.New("git", "clone", sampleArtifactsGitURI, tmpDir)
	require.NoError(t, cmd.Run())

	macosArchivePth := filepath.Join(tmpDir, "archives/macos.xcarchive")
	archive, err := NewMacosArchive(macosArchivePth)
	require.NoError(t, err)
	require.Equal(t, macosArchivePth, archive.Path)
	require.Equal(t, filepath.Join(macosArchivePth, "Info.plist"), archive.InfoPlistPath)

	app := archive.Application
	appPth := filepath.Join(macosArchivePth, "Products/Applications/Test.app")
	require.Equal(t, appPth, app.Path)
	require.Equal(t, filepath.Join(appPth, "Contents/Info.plist"), app.InfoPlistPath)
	require.Equal(t, "", app.ProvisioningProfilePath)
	require.Equal(t, filepath.Join(appPth, "Contents/Resources/archived-expanded-entitlements.xcent"), app.EntitlementsPath)

	require.Equal(t, 1, len(app.Extensions))
	extension := app.Extensions[0]
	extensionPth := filepath.Join(appPth, "Contents/PlugIns/ActionExtension.appex")
	require.Equal(t, extensionPth, extension.Path)
	require.Equal(t, filepath.Join(extensionPth, "Contents/Info.plist"), extension.InfoPlistPath)
	require.Equal(t, "", extension.ProvisioningProfilePath)
	require.Equal(t, filepath.Join(extensionPth, "Contents/Resources/archived-expanded-entitlements.xcent"), extension.EntitlementsPath)
}
