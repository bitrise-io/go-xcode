package codesign

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/xcarchive"
)

// Archive ...
type Archive struct {
	archive xcarchive.IosArchive
}

// NewArchive ...
func NewArchive(archive xcarchive.IosArchive) Archive {
	return Archive{
		archive: archive,
	}
}

// IsSigningManagedAutomatically ...
func (a Archive) IsSigningManagedAutomatically() (bool, error) {
	return a.archive.IsXcodeManaged(), nil
}

// Platform ...
func (a Archive) Platform() (autocodesign.Platform, error) {
	platformName := a.archive.Application.InfoPlist["DTPlatformName"]
	switch platformName {
	case "iphoneos":
		return autocodesign.IOS, nil
	case "appletvos":
		return autocodesign.TVOS, nil
	default:
		return "", fmt.Errorf("unsupported platform found: %s", platformName)
	}
}

// GetAppLayout ...
func (a Archive) GetAppLayout(uiTestTargets bool) (autocodesign.AppLayout, error) {
	layout, err := a.archive.ReadCodesignParameters()
	if err != nil {
		return autocodesign.AppLayout{}, err
	}
	return *layout, nil
}

// ForceCodesignAssets ...
func (a Archive) ForceCodesignAssets(distribution autocodesign.DistributionType, codesignAssetsByDistributionType map[autocodesign.DistributionType]autocodesign.AppCodesignAssets) error {
	return nil
}
