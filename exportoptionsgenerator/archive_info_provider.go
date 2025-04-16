package exportoptionsgenerator

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/xcarchive"
)

// ExportProduct ...
type ExportProduct string

const (
	// ExportProductApp ...
	ExportProductApp ExportProduct = "app"
	// ExportProductAppClip ...
	ExportProductAppClip ExportProduct = "app-clip"
)

// ArchiveInfoProvider fetches bundleID and entitlements from an xcarchive.
type ArchiveInfoProvider struct {
	archive       xcarchive.IosArchive
	exportProduct ExportProduct
}

// NewIosArchiveInfoProvider ...
func NewIosArchiveInfoProvider(archive xcarchive.IosArchive, exportProduct ExportProduct) InfoProvider {
	return ArchiveInfoProvider{
		archive:       archive,
		exportProduct: exportProduct,
	}
}

// Read ...
func (a ArchiveInfoProvider) Read() (ArchiveInfo, error) {
	productBundleID := ""
	appClipBundleID := ""
	if a.archive.Application.ClipApplication != nil {
		appClipBundleID = a.archive.Application.ClipApplication.BundleIdentifier()
	}

	switch a.exportProduct {
	case ExportProductApp:
		productBundleID = a.archive.Application.BundleIdentifier()
	case ExportProductAppClip:
		if appClipBundleID == "" {
			return ArchiveInfo{}, fmt.Errorf("xcarchive does not contain an App Clip, cannot export an App Clip")
		}
		productBundleID = appClipBundleID
	default:
		panic("unknown export product")
	}

	return ArchiveInfo{
		MainBundleID:           productBundleID,
		AppClipBundleID:        appClipBundleID,
		EntitlementsByBundleID: a.archive.BundleIDEntitlementsMap(),
	}, nil
}
