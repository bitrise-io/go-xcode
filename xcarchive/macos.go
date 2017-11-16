package xcarchive

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
)

type macosBaseApplication struct {
	Path                    string
	InfoPlistPath           string
	ProvisioningProfilePath string
	EntitlementsPath        string
}

func newMacosBaseApplication(path string) (macosBaseApplication, error) {
	infoPlistPath := filepath.Join(path, "Contents/Info.plist")
	if exist, err := pathutil.IsPathExists(infoPlistPath); err != nil {
		return macosBaseApplication{}, fmt.Errorf("failed to check if Info.plist exists at: %s, error: %s", infoPlistPath, err)
	} else if !exist {
		return macosBaseApplication{}, fmt.Errorf("Info.plist not exists at: %s", infoPlistPath)
	}

	provisioningProfilePath := filepath.Join(path, "Contents/Resources/embedded.mobileprovision")
	if exist, err := pathutil.IsPathExists(provisioningProfilePath); err != nil {
		return macosBaseApplication{}, fmt.Errorf("failed to check if profile exists at: %s, error: %s", provisioningProfilePath, err)
	} else if !exist {
		provisioningProfilePath = ""
	}
	fmt.Printf("mac profile: %s\n", provisioningProfilePath)

	entitlementsPath := filepath.Join(path, "Contents/Resources/archived-expanded-entitlements.xcent")
	if exist, err := pathutil.IsPathExists(entitlementsPath); err != nil {
		return macosBaseApplication{}, fmt.Errorf("failed to check if entitlements exists at: %s, error: %s", entitlementsPath, err)
	} else if !exist {
		return macosBaseApplication{}, fmt.Errorf("entitlements not exists at: %s", entitlementsPath)
	}

	return macosBaseApplication{
		Path:                    path,
		InfoPlistPath:           infoPlistPath,
		ProvisioningProfilePath: provisioningProfilePath,
		EntitlementsPath:        entitlementsPath,
	}, nil
}

// MacosExtension ...
type MacosExtension struct {
	macosBaseApplication
}

// NewMacosExtension ...
func NewMacosExtension(path string) (MacosExtension, error) {
	baseApp, err := newMacosBaseApplication(path)
	if err != nil {
		return MacosExtension{}, err
	}

	return MacosExtension{
		baseApp,
	}, nil
}

// MacosApplication ...
type MacosApplication struct {
	macosBaseApplication
	Extensions []MacosExtension
}

// NewMacosApplication ...
func NewMacosApplication(path string) (MacosApplication, error) {
	baseApp, err := newMacosBaseApplication(path)
	if err != nil {
		return MacosApplication{}, err
	}

	extensions := []MacosExtension{}
	{
		pattern := filepath.Join(path, "Contents/PlugIns/*.appex")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return MacosApplication{}, fmt.Errorf("failed to search for watch application's extensions using pattern: %s, error: %s", pattern, err)
		}
		for _, pth := range pths {
			extension, err := NewMacosExtension(pth)
			if err != nil {
				return MacosApplication{}, err
			}

			extensions = append(extensions, extension)
		}
	}

	return MacosApplication{
		macosBaseApplication: baseApp,
		Extensions:           extensions,
	}, nil
}

// MacosArchive ...
type MacosArchive struct {
	Path          string
	InfoPlistPath string

	Application MacosApplication
}

// NewMacosArchive ...
func NewMacosArchive(path string) (MacosArchive, error) {
	infoPlistPath := filepath.Join(path, "Info.plist")
	if exist, err := pathutil.IsPathExists(infoPlistPath); err != nil {
		return MacosArchive{}, fmt.Errorf("failed to check if Info.plist exists at: %s, error: %s", infoPlistPath, err)
	} else if !exist {
		return MacosArchive{}, fmt.Errorf("Info.plist not exists at: %s", infoPlistPath)
	}

	pattern := filepath.Join(path, "Products/Applications/*.app")
	pths, err := filepath.Glob(pattern)
	if err != nil {
		return MacosArchive{}, err
	}

	appPath := ""
	if len(pths) > 0 {
		appPath = pths[0]
	} else {
		return MacosArchive{}, fmt.Errorf("failed to find main app, using pattern: %s", pattern)
	}

	app, err := NewMacosApplication(appPath)
	if err != nil {
		return MacosArchive{}, err
	}

	return MacosArchive{
		Path:          path,
		InfoPlistPath: infoPlistPath,
		Application:   app,
	}, nil
}
