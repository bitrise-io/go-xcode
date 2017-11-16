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
	require.Equal(t, 5, len(archive.InfoPlist))

	app := archive.Application
	require.Equal(t, 26, len(app.InfoPlist))
	require.Equal(t, 2, len(app.Entitlements))
	require.Equal(t, "*", app.ProvisioningProfile.BundleID)

	require.Equal(t, 1, len(app.Extensions))
	extension := app.Extensions[0]
	require.Equal(t, 23, len(extension.InfoPlist))
	require.Equal(t, 2, len(extension.Entitlements))
	require.Equal(t, "*", extension.ProvisioningProfile.BundleID)

	require.NotNil(t, app.WatchApplication)
	watchApp := *app.WatchApplication
	require.Equal(t, 24, len(watchApp.InfoPlist))
	require.Equal(t, 2, len(watchApp.Entitlements))
	require.Equal(t, "*", watchApp.ProvisioningProfile.BundleID)

	require.Equal(t, 1, len(watchApp.Extensions))
	watchExtension := watchApp.Extensions[0]
	require.Equal(t, 23, len(watchExtension.InfoPlist))
	require.Equal(t, 2, len(watchExtension.Entitlements))
	require.Equal(t, "*", watchExtension.ProvisioningProfile.BundleID)
}
