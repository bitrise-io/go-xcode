package metaparser

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/artifacts"
	"github.com/bitrise-io/go-xcode/v2/zip"
)

// ParseXCArchiveData ...
func (m *Parser) ParseXCArchiveData(pth string) (*ArtifactMetadata, error) {

	appInfo, scheme, err := m.readXCArchiveDeploymentMeta(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment info for %s: %w", pth, err)
	}

	fileSize, err := m.fileManager.FileSizeInBytes(pth)
	if err != nil {
		m.logger.Warnf("Failed to get apk size, error: %s", err)
	}

	return &ArtifactMetadata{
		AppInfo:       appInfo,
		FileSizeBytes: fileSize,
		Scheme:        scheme,
	}, nil
}

func (m *Parser) readXCArchiveDeploymentMeta(pth string) (Info, string, error) {
	reader, err := zip.NewDefaultReader(pth, m.logger)
	if err != nil {
		return Info{}, "", err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			m.logger.Warnf("%s", err)
		}
	}()

	xcarchiveReader := artifacts.NewXCArchiveReader(reader)
	isMacos := xcarchiveReader.IsMacOS()
	if isMacos {
		m.logger.Warnf("macOS archive deployment is not supported, skipping xcarchive")
		return Info{}, "", nil // MacOS project is not supported, so won't be deployed.
	}
	archiveInfoPlist, err := xcarchiveReader.InfoPlist()
	if err != nil {
		return Info{}, "", fmt.Errorf("failed to unwrap Info.plist from xcarchive: %w", err)
	}

	iosXCArchiveReader := artifacts.NewIOSXCArchiveReader(reader)
	appInfoPlist, err := iosXCArchiveReader.AppInfoPlist()
	if err != nil {
		return Info{}, "", fmt.Errorf("failed to unwrap application Info.plist from xcarchive: %w", err)
	}

	appTitle, _ := appInfoPlist.GetString("CFBundleName")
	bundleID, _ := appInfoPlist.GetString("CFBundleIdentifier")
	version, _ := appInfoPlist.GetString("CFBundleShortVersionString")
	buildNumber, _ := appInfoPlist.GetString("CFBundleVersion")
	minOSVersion, _ := appInfoPlist.GetString("MinimumOSVersion")
	deviceFamilyList, _ := appInfoPlist.GetUInt64Array("UIDeviceFamily")
	scheme, _ := archiveInfoPlist.GetString("SchemeName")

	appInfo := Info{
		AppTitle:         appTitle,
		BundleID:         bundleID,
		Version:          version,
		BuildNumber:      buildNumber,
		MinOSVersion:     minOSVersion,
		DeviceFamilyList: deviceFamilyList,
	}

	return appInfo, scheme, nil
}
