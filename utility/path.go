// Package utility hosts cross-cutting helpers that did not fit a more specific
// package during the v1 → v2 migration.
package utility

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// FindFileInAppDir locates fileName inside appDir. If appDir/fileName exists,
// the path is returned directly; otherwise every `*.app` bundle under appDir
// is searched for an entry whose base name equals fileName, and the first
// match (joined back onto appDir) is returned.
func FindFileInAppDir(appDir, fileName string) (string, error) {
	directPth := filepath.Join(appDir, fileName)
	if exist, err := pathutil.NewPathChecker().IsPathExists(directPth); err != nil {
		return "", err
	} else if exist {
		return directPth, nil
	}

	appDirFS := os.DirFS(appDir)
	apps, err := pathutil.FilterFS(appDirFS, pathutil.ExtensionFilter(".app", true))
	if err != nil {
		return "", err
	}

	for _, app := range apps {
		matches, err := pathutil.FilterFS(os.DirFS(filepath.Join(appDir, app)), pathutil.BaseFilter(fileName, true))
		if err != nil {
			return "", err
		}
		if len(matches) > 0 {
			return filepath.Join(appDir, app, matches[0]), nil
		}
	}

	return "", fmt.Errorf("failed to find %s", fileName)
}
