package xcodeproj

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/internal/xcproj"
)

const (
	assetCatalogExtension = ".xcassets"
	projectPathBase       = "<PROJECT>/"
)

var filePathUnescaper = strings.NewReplacer(`\\`, `\`, `\<`, `<`)

// AppIconSetPaths implements XcodeProj. It looks for asset catalogs in the targets' resources
// build phases and, on disk, in the targets' synchronized folders.
func (p *XCProj) AppIconSetPaths() (map[string][]string, error) {
	return p.appIconSetPaths(p.assetCatalogPaths)
}

func (p *XCProj) assetCatalogPaths(target Target) ([]string, error) {
	resources := xcproj.BuildPhaseReference{Target: target.Name, Kind: xcproj.BuildPhaseKindResources}
	projectDir := filepath.Dir(p.path)

	var catalogs, folders []string
	walkFiles(p.files, projectDir, projectDir, func(reference xcproj.Reference, pth string) {
		switch reference.Kind {
		case xcproj.ReferenceKindFolder:
			if slices.Contains(reference.Targets, target.Name) {
				folders = append(folders, pth)
			}
		case xcproj.ReferenceKindFile:
			if filepath.Ext(pth) == assetCatalogExtension && slices.Contains(reference.BuildPhases, resources) {
				catalogs = append(catalogs, pth)
			}
		}
	})

	for _, folder := range folders {
		folderCatalogs, err := p.findAssetCatalogs(folder)
		if err != nil {
			return nil, err
		}
		catalogs = append(catalogs, folderCatalogs...)
	}

	return catalogs, nil
}

func walkFiles(references []xcproj.Reference, groupDir, projectDir string, visit func(xcproj.Reference, string)) {
	for _, reference := range references {
		pth, ok := resolveXCProjFilePath(reference.Path, groupDir, projectDir)
		if !ok {
			continue
		}

		visit(reference, pth)
		if reference.Kind == xcproj.ReferenceKindGroup {
			walkFiles(reference.Children, pth, projectDir, visit)
		}
	}
}

// resolveXCProjFilePath reports false for a path whose base only the build machine can resolve,
// such as <PRODUCTS>/, <SDK>/, <USER:SETTING>/ or ~.
func resolveXCProjFilePath(pth, groupDir, projectDir string) (string, bool) {
	if relativePath, ok := strings.CutPrefix(pth, projectPathBase); ok {
		return filepath.Join(projectDir, relativePath), true
	}
	if strings.HasPrefix(pth, "<") || strings.HasPrefix(pth, "~") {
		return "", false
	}

	pth = filePathUnescaper.Replace(pth)
	if filepath.IsAbs(pth) {
		return filepath.Clean(pth), true
	}
	return filepath.Join(groupDir, pth), true
}

// findAssetCatalogs does not follow symbolic links.
func (p *XCProj) findAssetCatalogs(dir string) ([]string, error) {
	names, err := p.fileManager.ReadDirEntryNames(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", dir, err)
	}

	var catalogs []string
	for _, name := range names {
		pth := filepath.Join(dir, name)
		info, err := p.fileManager.Lstat(pth)
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %w", pth, err)
		}
		if !info.IsDir() {
			continue
		}

		if filepath.Ext(name) == assetCatalogExtension {
			catalogs = append(catalogs, pth)
			continue
		}

		nestedCatalogs, err := p.findAssetCatalogs(pth)
		if err != nil {
			return nil, err
		}
		catalogs = append(catalogs, nestedCatalogs...)
	}

	return catalogs, nil
}
