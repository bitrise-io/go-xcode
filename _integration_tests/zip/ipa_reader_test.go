package zip

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/artifacts"
	"github.com/bitrise-io/go-xcode/v2/zip"
	"github.com/stretchr/testify/require"
)

func TestIPAReader_StdlibZipReader(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	watchTestIPAPath := filepath.Join(sampleArtifactsDir, "ipas", "watch-test.ipa")

	plist, profile := readIPAWithStdlibZipReader(t, watchTestIPAPath)
	bundleID, _ := plist.GetString("CFBundleIdentifier")
	require.Equal(t, "bitrise.watch-test", bundleID)
	require.Equal(t, "XC iOS: *", profile.Name)
}

func TestIPAReader_DittoZipReader(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	watchTestIPAPath := filepath.Join(sampleArtifactsDir, "ipas", "watch-test.ipa")

	plist, profile := readIPAWithDittoZipReader(t, watchTestIPAPath)
	bundleID, _ := plist.GetString("CFBundleIdentifier")
	require.Equal(t, "bitrise.watch-test", bundleID)
	require.Equal(t, "XC iOS: *", profile.Name)
}

func Benchmark_ZipReaders(b *testing.B) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(b)
	watchTestIPAPath := filepath.Join(sampleArtifactsDir, "ipas", "watch-test.ipa")

	for name, zipFunc := range map[string]readIPAFunc{
		"dittoReader": func() (plistutil.PlistData, *profileutil.ProvisioningProfileInfoModel) {
			return readIPAWithDittoZipReader(b, watchTestIPAPath)
		},
		"stdlibReader": func() (plistutil.PlistData, *profileutil.ProvisioningProfileInfoModel) {
			return readIPAWithStdlibZipReader(b, watchTestIPAPath)
		},
	} {
		b.Run(fmt.Sprintf("Benchmarking %s", name), func(b *testing.B) {
			_, _ = zipFunc()
		})
	}
}

type readIPAFunc func() (plistutil.PlistData, *profileutil.ProvisioningProfileInfoModel)

func readIPAWithStdlibZipReader(t require.TestingT, archivePth string) (plistutil.PlistData, *profileutil.ProvisioningProfileInfoModel) {
	r, err := zip.NewStdlibRead(archivePth, log.NewLogger())
	require.NoError(t, err)
	defer func() {
		err := r.Close()
		require.NoError(t, err)
	}()

	ipaReader := artifacts.NewIPAReader(r)
	plist, err := ipaReader.AppInfoPlist()
	require.NoError(t, err)

	profile, err := ipaReader.ProvisioningProfileInfo()
	require.NoError(t, err)

	return plist, profile
}

func readIPAWithDittoZipReader(t require.TestingT, archivePth string) (plistutil.PlistData, *profileutil.ProvisioningProfileInfoModel) {
	r := zip.NewDittoReader(archivePth, log.NewLogger())
	defer func() {
		err := r.Close()
		require.NoError(t, err)
	}()

	ipaReader := artifacts.NewIPAReader(r)
	plist, err := ipaReader.AppInfoPlist()
	require.NoError(t, err)

	profile, err := ipaReader.ProvisioningProfileInfo()
	require.NoError(t, err)

	return plist, profile
}
