package xcarchive

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
)

type iosBaseApplication struct {
	Path                    string
	InfoPlistPath           string
	ProvisioningProfilePath string
	EntitlementsPath        string
}

func newIosBaseApplication(path string) (iosBaseApplication, error) {
	infoPlistPath := filepath.Join(path, "Info.plist")
	if exist, err := pathutil.IsPathExists(infoPlistPath); err != nil {
		return iosBaseApplication{}, fmt.Errorf("failed to check if Info.plist exists at: %s, error: %s", infoPlistPath, err)
	} else if !exist {
		return iosBaseApplication{}, fmt.Errorf("Info.plist not exists at: %s", infoPlistPath)
	}

	provisioningProfilePath := filepath.Join(path, "embedded.mobileprovision")
	if exist, err := pathutil.IsPathExists(provisioningProfilePath); err != nil {
		return iosBaseApplication{}, fmt.Errorf("failed to check if profile exists at: %s, error: %s", provisioningProfilePath, err)
	} else if !exist {
		return iosBaseApplication{}, fmt.Errorf("profile not exists at: %s", provisioningProfilePath)
	}

	entitlementsPath := filepath.Join(path, "archived-expanded-entitlements.xcent")
	if exist, err := pathutil.IsPathExists(entitlementsPath); err != nil {
		return iosBaseApplication{}, fmt.Errorf("failed to check if entitlements exists at: %s, error: %s", entitlementsPath, err)
	} else if !exist {
		return iosBaseApplication{}, fmt.Errorf("entitlements not exists at: %s", entitlementsPath)
	}

	return iosBaseApplication{
		Path:                    path,
		InfoPlistPath:           infoPlistPath,
		ProvisioningProfilePath: provisioningProfilePath,
		EntitlementsPath:        entitlementsPath,
	}, nil
}

// IosExtension ...
type IosExtension struct {
	iosBaseApplication
}

// NewIosExtension ...
func NewIosExtension(path string) (IosExtension, error) {
	baseApp, err := newIosBaseApplication(path)
	if err != nil {
		return IosExtension{}, err
	}

	return IosExtension{
		baseApp,
	}, nil
}

// IosWatchApplication ...
type IosWatchApplication struct {
	iosBaseApplication
	Extensions []IosExtension
}

// NewIosWatchApplication ...
func NewIosWatchApplication(path string) (IosWatchApplication, error) {
	baseApp, err := newIosBaseApplication(path)
	if err != nil {
		return IosWatchApplication{}, err
	}

	extensions := []IosExtension{}
	pattern := filepath.Join(path, "PlugIns/*.appex")
	pths, err := filepath.Glob(pattern)
	if err != nil {
		return IosWatchApplication{}, fmt.Errorf("failed to search for watch application's extensions using pattern: %s, error: %s", pattern, err)
	}
	for _, pth := range pths {
		extension, err := NewIosExtension(pth)
		if err != nil {
			return IosWatchApplication{}, err
		}

		extensions = append(extensions, extension)
	}

	return IosWatchApplication{
		iosBaseApplication: baseApp,
		Extensions:         extensions,
	}, nil
}

// IosApplication ...
type IosApplication struct {
	iosBaseApplication
	WatchApplication *IosWatchApplication
	Extensions       []IosExtension
}

// NewIosApplication ...
func NewIosApplication(path string) (IosApplication, error) {
	baseApp, err := newIosBaseApplication(path)
	if err != nil {
		return IosApplication{}, err
	}

	var watchApp *IosWatchApplication
	{
		pattern := filepath.Join(path, "Watch/*.app")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return IosApplication{}, err
		}
		if len(pths) > 0 {
			watchPath := pths[0]
			app, err := NewIosWatchApplication(watchPath)
			if err != nil {
				return IosApplication{}, err
			}
			watchApp = &app
		}
	}

	extensions := []IosExtension{}
	{
		pattern := filepath.Join(path, "PlugIns/*.appex")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return IosApplication{}, fmt.Errorf("failed to search for watch application's extensions using pattern: %s, error: %s", pattern, err)
		}
		for _, pth := range pths {
			extension, err := NewIosExtension(pth)
			if err != nil {
				return IosApplication{}, err
			}

			extensions = append(extensions, extension)
		}
	}

	return IosApplication{
		iosBaseApplication: baseApp,
		WatchApplication:   watchApp,
		Extensions:         extensions,
	}, nil
}

// IosArchive ...
type IosArchive struct {
	Path          string
	InfoPlistPath string

	Application IosApplication
}

// NewIosArchive ...
func NewIosArchive(path string) (IosArchive, error) {
	infoPlistPath := filepath.Join(path, "Info.plist")
	if exist, err := pathutil.IsPathExists(infoPlistPath); err != nil {
		return IosArchive{}, fmt.Errorf("failed to check if Info.plist exists at: %s, error: %s", infoPlistPath, err)
	} else if !exist {
		return IosArchive{}, fmt.Errorf("Info.plist not exists at: %s", infoPlistPath)
	}

	pattern := filepath.Join(path, "Products/Applications/*.app")
	pths, err := filepath.Glob(pattern)
	if err != nil {
		return IosArchive{}, err
	}

	appPath := ""
	if len(pths) > 0 {
		appPath = pths[0]
	} else {
		return IosArchive{}, fmt.Errorf("failed to find main app, using pattern: %s", pattern)
	}

	app, err := NewIosApplication(appPath)
	if err != nil {
		return IosArchive{}, err
	}

	return IosArchive{
		Path:          path,
		InfoPlistPath: infoPlistPath,
		Application:   app,
	}, nil
}
