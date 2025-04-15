package exportoptionsgenerator

import (
	"github.com/bitrise-io/go-xcode/exportoptions"
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

// ArchiveTargetInfoProvider implements TargetInfoProvider.
type ArchiveTargetInfoProvider struct {
	archive       xcarchive.IosArchive
	exportProduct ExportProduct
}

func NewIosArchiveTargetInfoProvider(archive xcarchive.IosArchive, exportProduct ExportProduct) TargetInfoProvider {
	return ArchiveTargetInfoProvider{
		archive:       archive,
		exportProduct: exportProduct,
	}
}

func (a ArchiveTargetInfoProvider) applicationTargetsAndEntitlements(_ exportoptions.Method) (string, map[string]plistutil.PlistData, error) {
	productBundleID := ""

	switch a.exportProduct {
	case ExportProductApp:
		productBundleID = a.archive.Application.BundleIdentifier()
	case ExportProductAppClip:
		productBundleID = a.archive.Application.ClipApplication.BundleIdentifier()
	default:
		panic("unknown export product")
	}

	return productBundleID, a.archive.BundleIDEntitlementsMap(), nil
}
