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
	require.Equal(t, 5, len(archive.InfoPlist))

	app := archive.Application
	require.Equal(t, 21, len(app.InfoPlist))
	require.Equal(t, 2, len(app.Entitlements))
	require.Nil(t, app.ProvisioningProfile)

	require.Equal(t, 1, len(app.Extensions))
	extension := app.Extensions[0]
	require.Equal(t, 22, len(extension.InfoPlist))
	require.Equal(t, 2, len(extension.Entitlements))
	require.Nil(t, extension.ProvisioningProfile)
}
