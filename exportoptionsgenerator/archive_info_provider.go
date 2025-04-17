package exportoptionsgenerator

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/xcarchive"
)

// ExportProduct ...
type ExportProduct string

const (
	// ExportProductApp ...
	ExportProductApp ExportProduct = "app"
	// ExportProductAppClip ...
	ExportProductAppClip ExportProduct = "app-clip"
)

// ReadArchiveExportInfo ...
func ReadArchiveExportInfo(archive xcarchive.IosArchive, exportedProduct ExportProduct) (ArchiveInfo, error) {
	productBundleID := ""
	appClipBundleID := ""
	if archive.Application.ClipApplication != nil {
		appClipBundleID = archive.Application.ClipApplication.BundleIdentifier()
	}

	switch exportedProduct {
	case ExportProductApp:
		productBundleID = archive.Application.BundleIdentifier()
	case ExportProductAppClip:
		if appClipBundleID == "" {
			return ArchiveInfo{}, fmt.Errorf("xcarchive does not contain an App Clip, cannot export an App Clip")
		}
		productBundleID = appClipBundleID
	default:
		panic("unknown export product")
	}

	return ArchiveInfo{
		ProductToDistributeBundleID: productBundleID,
		AppClipBundleID:             appClipBundleID,
		EntitlementsByBundleID:      archive.BundleIDEntitlementsMap(),
	}, nil
}
