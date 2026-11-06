package xcodeproj

import (
	"fmt"
	"path"
	"regexp"

	"github.com/bitrise-io/go-utils/v2/sliceutil"
)

// appIconSetNameBuildSettingKey is read from the project file, not through BuildSettingsProvider
// like other build settings, so no xcodebuild call is needed per target and configuration. As a
// result, values set in .xcconfig files are not seen, and ${...} references are matched as globs.
const appIconSetNameBuildSettingKey = "ASSETCATALOG_COMPILER_APPICON_NAME"

const assetCatalogExtension = ".xcassets"

var buildSettingReferenceRegexp = regexp.MustCompile(`\$\{(.+)\}`)

// AppIconSetPaths implements XcodeProj.
func (p *xcodeProject) AppIconSetPaths() (map[string][]string, error) {
	targetToAppIcons := map[string][]string{}
	for _, target := range p.targets {
		// Only native targets build resources such as asset catalogs.
		if !target.isNativeTarget() {
			continue
		}

		appIconSetNames := appIconSetNames(target)
		if len(appIconSetNames) == 0 {
			continue
		}

		catalogs, err := p.file.assetCatalogPaths(target)
		if err != nil {
			return nil, err
		} else if len(catalogs) == 0 {
			continue
		}

		var appIcons []string
		for _, appIconSetName := range appIconSetNames {
			paths, err := p.globAppIconSets(catalogs, appIconSetName)
			if err != nil {
				return nil, err
			} else if len(paths) == 0 {
				return nil, fmt.Errorf("not found app icon set (%s) on paths: %s", appIconSetName, catalogs)
			}
			appIcons = append(appIcons, paths...)
		}
		targetToAppIcons[target.ID] = sliceutil.Unique(appIcons)
	}

	return targetToAppIcons, nil
}

func (p *xcodeProject) globAppIconSets(catalogPaths []string, appIconSetName string) ([]string, error) {
	pattern := buildSettingReferenceRegexp.ReplaceAllString(appIconSetName, "*")

	var icons []string
	for _, catalogPath := range catalogPaths {
		matches, err := p.pathProvider.Glob(path.Join(p.pathModifier.EscapeGlobPath(catalogPath), pattern+".appiconset"))
		if err != nil {
			return nil, err
		}
		icons = append(icons, matches...)
	}

	return icons, nil
}

// appIconSetNames returns nothing if any configuration lacks the setting.
func appIconSetNames(target Target) []string {
	var names []string
	for _, configuration := range target.BuildConfigurations {
		name, ok := configuration.BuildSetting(appIconSetNameBuildSettingKey)
		if !ok {
			return nil
		}
		names = append(names, name)
	}
	return names
}
