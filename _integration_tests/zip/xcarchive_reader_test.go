package zip

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/artifacts"
	"github.com/bitrise-io/go-xcode/v2/zip"
	"github.com/stretchr/testify/require"
)

func TestXCArchiveReader_DefaultZipReader_MacOSArchive(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	macOSXCArchivePath := filepath.Join(sampleArtifactsDir, "archives", "macos.xcarchive.zip")

	r, err := zip.NewStdlibRead(macOSXCArchivePath, log.NewLogger())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close())
	}()

	xcarchiveReader := artifacts.NewXCArchiveReader(r)
	plist, err := xcarchiveReader.InfoPlist()
	require.NoError(t, err)
	name, _ := plist.GetString("Name")
	require.Equal(t, "ActionExtension", name)

	require.NoError(t, err)
	require.Equal(t, true, xcarchiveReader.IsMacOS())
}

func TestXCArchiveReader_DittoZipReader_MacOSArchive(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	macOSXCArchivePath := filepath.Join(sampleArtifactsDir, "archives", "macos.xcarchive.zip")

	r := zip.NewDittoReader(macOSXCArchivePath, log.NewLogger())
	defer func() {
		require.NoError(t, r.Close())
	}()

	xcarchiveReader := artifacts.NewXCArchiveReader(r)
	plist, err := xcarchiveReader.InfoPlist()
	require.NoError(t, err)
	name, _ := plist.GetString("Name")
	require.Equal(t, "ActionExtension", name)

	require.NoError(t, err)
	require.Equal(t, true, xcarchiveReader.IsMacOS())
}

func TestXCArchiveReader_DefaultZipReader_IOSArchive(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	iosXCArchiveIPAPath := filepath.Join(sampleArtifactsDir, "archives", "ios.xcarchive.zip")

	r, err := zip.NewStdlibRead(iosXCArchiveIPAPath, log.NewLogger())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close())
	}()

	xcarchiveReader := artifacts.NewXCArchiveReader(r)
	plist, err := xcarchiveReader.InfoPlist()
	require.NoError(t, err)
	name, _ := plist.GetString("Name")
	require.Equal(t, "code-sign-test", name)

	require.NoError(t, err)
	require.Equal(t, false, xcarchiveReader.IsMacOS())

	iosXCArchiveReader := artifacts.NewIOSXCArchiveReader(r)
	appPlist, err := iosXCArchiveReader.AppInfoPlist()
	require.NoError(t, err)
	name, _ = appPlist.GetString("CFBundleIdentifier")
	require.Equal(t, "com.bitrise.code-sign-test", name)
}

func TestXCArchiveReader_DittoZipReader_IOSArchive(t *testing.T) {
	sampleArtifactsDir := _integration_tests.GetSampleArtifactsRepository(t)
	iosXCArchiveIPAPath := filepath.Join(sampleArtifactsDir, "archives", "ios.xcarchive.zip")

	r := zip.NewDittoReader(iosXCArchiveIPAPath, log.NewLogger())
	defer func() {
		require.NoError(t, r.Close())
	}()

	xcarchiveReader := artifacts.NewXCArchiveReader(r)
	plist, err := xcarchiveReader.InfoPlist()
	require.NoError(t, err)
	name, _ := plist.GetString("Name")
	require.Equal(t, "code-sign-test", name)

	require.NoError(t, err)
	require.Equal(t, false, xcarchiveReader.IsMacOS())

	iosXCArchiveReader := artifacts.NewIOSXCArchiveReader(r)
	appPlist, err := iosXCArchiveReader.AppInfoPlist()
	require.NoError(t, err)
	name, _ = appPlist.GetString("CFBundleIdentifier")
	require.Equal(t, "com.bitrise.code-sign-test", name)
}
