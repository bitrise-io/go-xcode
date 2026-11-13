package exportoptionsgenerator

import (
	"github.com/bitrise-io/go-xcode/v2/xcarchive"
)

type ExportProduct string

const (
	ExportProductApp     ExportProduct = "app"
	ExportProductAppClip ExportProduct = "app-clip"
)

func ReadArchiveExportInfo(archive xcarchive.IosArchive) (ArchiveInfo, error) {
	appClipBundleID := ""
	if archive.Application.ClipApplication != nil {
		appClipBundleID = archive.Application.ClipApplication.BundleIdentifier()
	}

	return ArchiveInfo{
		AppBundleID:            archive.Application.BundleIdentifier(),
		AppClipBundleID:        appClipBundleID,
		EntitlementsByBundleID: archive.BundleIDEntitlementsMap(),
	}, nil
}
