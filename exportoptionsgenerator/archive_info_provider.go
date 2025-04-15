package exportoptionsgenerator

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/plistutil"
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

// ArchiveInfoProvider implements InfoProvider.
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

// ExportableBundleIDToEntitlements ...
func (a ArchiveInfoProvider) ExportableBundleIDToEntitlements() (string, map[DistributedProduct]plistutil.PlistData, error) {
	productBundleID := ""
	appClipBundleID := ""

	switch a.exportProduct {
	case ExportProductApp:
		productBundleID = a.archive.Application.BundleIdentifier()
	case ExportProductAppClip:
		if a.archive.Application.ClipApplication == nil {
			return "", nil, fmt.Errorf("xcarchive does not contain an App Clip, cannot export an App Clip")
		}

		appClipBundleID = a.archive.Application.ClipApplication.BundleIdentifier()
		productBundleID = appClipBundleID
	default:
		panic("unknown export product")
	}

	bundleIDToEntitlements := map[DistributedProduct]plistutil.PlistData{}
	for bundleID, entitlements := range a.archive.BundleIDEntitlementsMap() {
		if appClipBundleID != "" && bundleID == appClipBundleID {
			bundleIDToEntitlements[DistributedProduct{BundleID: bundleID, IsAppClip: true}] = entitlements
		} else {
			bundleIDToEntitlements[DistributedProduct{BundleID: bundleID, IsAppClip: false}] = entitlements
		}
	}

	return productBundleID, bundleIDToEntitlements, nil
}
