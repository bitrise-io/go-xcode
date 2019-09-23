package cache

import (
	"fmt"
	"path"

	"github.com/bitrise-io/go-steputils/cache"
)

// CollectPackagesCache marks the Swift Package Manager packages to be added the cache
// The directory cached is: $HOME/Library/Developer/Xcode/DerivedData/[PER_PROJECT_DERIVED_DATA]/SourcePackages
func CollectPackagesCache(projectPath string) error {
	projectDerivedData, err := xcodeProjectDerivedDataPath(projectPath)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	projectSwiftPMDir := path.Join(projectDerivedData, "SourcePackages")

	cache := cache.New()
	cache.IncludePath(projectSwiftPMDir)
	// Excluding manifest.db will result in a stable cache, as this file is modified in every build.
	cache.ExcludePath("!" + path.Join(projectSwiftPMDir, "manifest.db"))
	if err := cache.Commit(); err != nil {
		return fmt.Errorf("failed to commit cache, error: %s", err)
	}
	return nil
}
