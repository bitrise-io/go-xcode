package cache

import (
	"fmt"
	"path"

	"github.com/bitrise-io/go-steputils/cache"
)

// CollectXcodeSwiftPackagesCache marks the Swift Package Manager packages to be added the cache
// The directory cached is: $HOME/Library/Developer/Xcode/DerivedData/[PER_PROJECT_DERIVED_DATA]/SourcePackages
func CollectXcodeSwiftPackagesCache(projectPath string) error {
	projectDerivedData, err := xcodeProjectDerivedDataPath(projectPath)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	projectSwiftPMDir := path.Join(projectDerivedData, "SourcePackages")

	cache := cache.New()
	cache.IncludePath(projectSwiftPMDir)
	if err := cache.Commit(); err != nil {
		return fmt.Errorf("failed to commit cache, error: %s", err)
	}
	return nil
}
