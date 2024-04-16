package zip

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/zip"
	"github.com/bitrise-io/go-xcode/v2/ziputil"
	"github.com/stretchr/testify/require"
)

func TestIPAReader_DefaultZipReader(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	watchTestIPAPath := filepath.Join(sampleArtifactsDir, "ipas", "watch-test.ipa")

	r, err := ziputil.NewDefaultRead(watchTestIPAPath, log.NewLogger())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close())
	}()

	ipaReader := zip.NewIPAReader(r)
	plist, err := ipaReader.AppInfoPlist()
	require.NoError(t, err)
	bundleID, _ := plist.GetString("CFBundleIdentifier")
	require.Equal(t, "bitrise.watch-test", bundleID)

	profile, err := ipaReader.ProvisioningProfileInfo()
	require.NoError(t, err)
	require.Equal(t, "XC iOS: *", profile.Name)
}

func TestIPAReader_DittoZipReader(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	watchTestIPAPath := filepath.Join(sampleArtifactsDir, "ipas", "watch-test.ipa")

	r, err := ziputil.NewDittoReader(watchTestIPAPath, log.NewLogger())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close())
	}()

	ipaReader := zip.NewIPAReader(r)
	plist, err := ipaReader.AppInfoPlist()
	require.NoError(t, err)
	bundleID, _ := plist.GetString("CFBundleIdentifier")
	require.Equal(t, "bitrise.watch-test", bundleID)

	profile, err := ipaReader.ProvisioningProfileInfo()
	require.NoError(t, err)
	require.Equal(t, "XC iOS: *", profile.Name)
}
