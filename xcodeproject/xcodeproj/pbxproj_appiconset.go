package xcodeproj

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// assetCatalogPaths looks for asset catalogs in the target's resources build phase.
//
// TODO: catalogs in the target's synchronized folders (PBXFileSystemSynchronizedRootGroup) are not
// found.
func (p *pbxprojFile) assetCatalogPaths(target Target) ([]string, error) {
	objects, ok := p.rawProj.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to find asset catalogs of target %s: project has no objects", target.Name)
	}

	catalogs, err := assetCatalogs(p.buildPhaseIDs[target.ID], objects)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(catalogs))
	for _, catalog := range catalogs {
		resolvedPath, err := resolveObjectAbsolutePath(catalog.id, p.projectID, p.path, objects)
		if err != nil {
			return nil, err
		} else if resolvedPath == "" {
			return nil, fmt.Errorf("could not resolve path")
		}
		paths = append(paths, resolvedPath)
	}

	return paths, nil
}

func assetCatalogs(buildPhaseIDs []string, objects serialized.Object) ([]fileReference, error) {
	buildPhase, err := findResourcesBuildPhase(buildPhaseIDs, objects)
	if err != nil {
		return nil, fmt.Errorf("getting resource build phases failed, error: %w", err)
	}

	return filterAssetCatalogs(buildPhase, objects)
}

func findResourcesBuildPhase(buildPhaseIDs []string, objects serialized.Object) (resourcesBuildPhase, error) {
	for _, id := range buildPhaseIDs {
		rawBuildPhase, ok := objects.Object(id)
		if !ok {
			return resourcesBuildPhase{}, fmt.Errorf("build phase %s not found", id)
		}
		if isResourcesBuildPhase(rawBuildPhase) {
			buildPhase, err := parseResourcesBuildPhase(id, objects)
			if err != nil {
				return resourcesBuildPhase{}, fmt.Errorf("failed to parse ResourcesBuildPhase, error: %w", err)
			}
			return buildPhase, nil
		}
	}
	return resourcesBuildPhase{}, fmt.Errorf("resource build phase not found")
}

func filterAssetCatalogs(buildPhase resourcesBuildPhase, objects serialized.Object) ([]fileReference, error) {
	var catalogs []fileReference
	for _, fileID := range buildPhase.files {
		file, err := parseBuildFile(fileID, objects)
		if err != nil {
			// A build file without a file reference, such as "(null) in Resources", is skipped.
			continue
		}

		rawElement, ok := objects.Object(file.fileRef)
		if !ok {
			return nil, fmt.Errorf("object %s not found", file.fileRef)
		}
		isReference, err := isFileReference(rawElement)
		if err != nil {
			return nil, err
		} else if !isReference {
			// PBXVariantGroup
			continue
		}

		reference, err := parseFileReference(file.fileRef, objects)
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(reference.path, assetCatalogExtension) {
			catalogs = append(catalogs, reference)
		}
	}
	return catalogs, nil
}
