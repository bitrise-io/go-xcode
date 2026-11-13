package metaparser

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/artifacts"
	"github.com/bitrise-io/go-xcode/v2/zip"
)

func (m *Parser) ParseIPAData(pth string) (*ArtifactMetadata, error) {
	appInfo, provisioningInfo, err := m.readIPADeploymentMeta(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment info for %s: %w", pth, err)
	}

	fileSize, err := m.fileManager.FileSizeInBytes(pth)
	if err != nil {
		m.logger.Warnf("Failed to get apk size, error: %s", err)
	}

	return &ArtifactMetadata{
		AppInfo:          appInfo,
		FileSizeBytes:    fileSize,
		ProvisioningInfo: provisioningInfo,
	}, nil
}

func (m *Parser) readIPADeploymentMeta(ipaPth string) (Info, ProvisionInfo, error) {
	reader, err := zip.NewDefaultReader(ipaPth, m.logger)
	if err != nil {
		return Info{}, ProvisionInfo{}, err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			m.logger.Warnf("%s", err)
		}
	}()

	ipaReader := artifacts.NewIPAReader(reader)
	infoPlist, err := ipaReader.AppInfoPlist()
	if err != nil {
		return Info{}, ProvisionInfo{}, fmt.Errorf("failed to unwrap Info.plist from ipa: %w", err)
	}

	appTitle, _ := infoPlist.GetString("CFBundleName")
	bundleID, _ := infoPlist.GetString("CFBundleIdentifier")
	version, _ := infoPlist.GetString("CFBundleShortVersionString")
	buildNumber, _ := infoPlist.GetString("CFBundleVersion")
	minOSVersion, _ := infoPlist.GetString("MinimumOSVersion")
	deviceFamilyList, _ := infoPlist.GetUInt64Array("UIDeviceFamily")

	appInfo := Info{
		AppTitle:         appTitle,
		BundleID:         bundleID,
		Version:          version,
		BuildNumber:      buildNumber,
		MinOSVersion:     minOSVersion,
		DeviceFamilyList: deviceFamilyList,
	}

	provisioningProfileInfo, err := ipaReader.ProvisioningProfileInfo()
	if err != nil {
		return Info{}, ProvisionInfo{}, fmt.Errorf("failed to read profile info from ipa: %w", err)
	}

	provisioningInfo := ProvisionInfo{
		CreationDate:         provisioningProfileInfo.CreationDate,
		ExpireDate:           provisioningProfileInfo.ExpirationDate,
		DeviceUDIDList:       provisioningProfileInfo.ProvisionedDevices,
		TeamName:             provisioningProfileInfo.TeamName,
		ProfileName:          provisioningProfileInfo.Name,
		ProvisionsAllDevices: provisioningProfileInfo.ProvisionsAllDevices,
		IPAExportMethod:      provisioningProfileInfo.ExportType,
	}

	return appInfo, provisioningInfo, nil
}
