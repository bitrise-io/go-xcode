package xcarchive

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"github.com/bitrise-tools/go-xcode/profileutil"
	"github.com/bitrise-tools/go-xcode/utility"
)

// Application ...
type Application struct {
	Path                string
	InfoPlist           plistutil.PlistData
	Entitlements        plistutil.PlistData
	ProvisioningProfile profileutil.ProvisioningProfileInfoModel
	Plugins             []Application
	WatchApplication    *Application
	IsMacOS             bool
}

// BundleIdentifier ...
func (app Application) BundleIdentifier() string {
	bundleID, _ := app.InfoPlist.GetString("CFBundleIdentifier")
	return bundleID
}

// NewApplication ...
func NewApplication(applicationsDir string) (Application, error) {
	mainApplication := Application{}
	mainApplicationPth := ""
	applicationContentPth := ""
	{
		pattern := filepath.Join(applicationsDir, "*.app")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return Application{}, err
		}

		if len(pths) == 0 {
			return Application{}, fmt.Errorf("Failed to find main application using pattern: %s", pattern)
		} else if len(pths) > 1 {
			log.Warnf("Multiple main applications found")
			for _, pth := range pths {
				log.Warnf("- %s", pth)
			}

			mainApplicationPth = pths[0]
			log.Warnf("Using first: %s", mainApplicationPth)
		} else {
			mainApplicationPth = pths[0]
		}

		application := Application{
			Path: mainApplicationPth,
		}

		{
			applicationContentPth = filepath.Join(mainApplicationPth, "Contents")
			exists, err := pathutil.IsPathExists(applicationContentPth)
			if err != nil {
				return Application{}, err
			}
			if exists {
				application.IsMacOS = true
			} else {
				application.IsMacOS = false
				applicationContentPth = mainApplicationPth
			}
		}

		{
			infoPlistPth := filepath.Join(applicationContentPth, "Info.plist")
			infoPlistExist, err := pathutil.IsPathExists(infoPlistPth)
			if err != nil {
				return Application{}, err
			}
			if !infoPlistExist {
				if !infoPlistExist {
					return Application{}, fmt.Errorf("Info.plist does not exist at: (%s)", infoPlistPth)
				}
			}
			infoPlist, err := plistutil.NewPlistDataFromFile(infoPlistPth)
			if err != nil {
				return Application{}, err
			}
			application.InfoPlist = infoPlist
		}

		{
			profileName := "embedded.mobileprovision"

			if application.IsMacOS {
				profileName = "embedded.provisionprofile"
			}

			provisioningProfilePth := filepath.Join(applicationContentPth, profileName)
			exist, err := pathutil.IsPathExists(provisioningProfilePth)
			if err != nil {
				return Application{}, err
			} else if !exist && !application.IsMacOS {
				return Application{}, fmt.Errorf("%s does not exist at: %s", profileName, provisioningProfilePth)
			}
			if exist {
				profile, err := profileutil.NewProvisioningProfileInfoFromFile(provisioningProfilePth)
				if err != nil {
					return Application{}, err
				}
				application.ProvisioningProfile = profile
			}
		}

		{
			entitlementsBasePth := applicationContentPth
			if application.IsMacOS {
				entitlementsBasePth = filepath.Join(entitlementsBasePth, "Resources")
			}

			entitlementsPth := filepath.Join(entitlementsBasePth, "archived-expanded-entitlements.xcent")
			exist, err := pathutil.IsPathExists(entitlementsPth)
			if err != nil {
				return Application{}, err
			} else if exist {
				entitlements, err := plistutil.NewPlistDataFromFile(entitlementsPth)
				if err != nil {
					return Application{}, err
				}

				application.Entitlements = entitlements
			}
		}
		mainApplication = application
	}

	plugins := []Application{}
	{
		pattern := filepath.Join(applicationContentPth, "PlugIns/*.appex")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return Application{}, err
		}
		for _, pth := range pths {
			plugin, err := NewApplication(pth)
			if err != nil {
				return Application{}, err
			}

			plugins = append(plugins, plugin)
		}
		mainApplication.Plugins = plugins
	}

	var watchApplicationPtr *Application
	watchApplicationPth := ""
	{
		pattern := filepath.Join(applicationContentPth, "Watch/*.app")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return Application{}, err
		}

		if len(pths) > 1 {
			log.Warnf("Multiple watch applications found")
			for _, pth := range pths {
				log.Warnf("- %s", pth)
			}

			watchApplicationPth = pths[0]
			log.Warnf("Using first: %s", watchApplicationPth)
		} else if len(pths) == 1 {
			watchApplicationPth = pths[0]
		}

		if watchApplicationPth != "" {
			watchApplication, err := NewApplication(watchApplicationPth)
			if err != nil {
				return Application{}, err
			}

			watchApplicationPtr = &watchApplication
		}
	}

	watchPlugins := []Application{}
	{
		if watchApplicationPth != "" {
			pattern := filepath.Join(watchApplicationPth, "PlugIns/*.appex")
			pths, err := filepath.Glob(pattern)
			if err != nil {
				return Application{}, err
			}
			for _, pth := range pths {
				plugin, err := NewApplication(pth)
				if err != nil {
					return Application{}, err
				}

				watchPlugins = append(watchPlugins, plugin)
			}
			(*watchApplicationPtr).Plugins = watchPlugins
		}
	}

	return mainApplication, nil
}

// XCArchive ...
type XCArchive struct {
	Path        string
	Application Application
	InfoPlist   plistutil.PlistData
}

// IsXcodeManaged ...
func (archive XCArchive) IsXcodeManaged() bool {
	return archive.Application.ProvisioningProfile.IsXcodeManaged()
}

// SigningIdentity ...
func (archive XCArchive) SigningIdentity() string {
	properties, found := archive.InfoPlist.GetMapStringInterface("ApplicationProperties")
	if found {
		identity, _ := properties.GetString("SigningIdentity")
		return identity
	}
	return ""
}

// BundleIDEntitlementsMap ...
func (archive XCArchive) BundleIDEntitlementsMap() map[string]plistutil.PlistData {
	bundleIDEntitlementsMap := map[string]plistutil.PlistData{}

	bundleID := archive.Application.BundleIdentifier()
	bundleIDEntitlementsMap[bundleID] = archive.Application.Entitlements

	for _, plugin := range archive.Application.Plugins {
		bundleID := plugin.BundleIdentifier()
		bundleIDEntitlementsMap[bundleID] = plugin.Entitlements
	}

	if archive.Application.WatchApplication != nil {
		watchApplication := *archive.Application.WatchApplication

		bundleID := watchApplication.BundleIdentifier()
		bundleIDEntitlementsMap[bundleID] = watchApplication.Entitlements

		for _, plugin := range watchApplication.Plugins {
			bundleID := plugin.BundleIdentifier()
			bundleIDEntitlementsMap[bundleID] = plugin.Entitlements
		}
	}

	return bundleIDEntitlementsMap
}

// BundleIDProfileInfoMap ...
func (archive XCArchive) BundleIDProfileInfoMap() map[string]profileutil.ProvisioningProfileInfoModel {
	bundleIDProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}

	bundleID := archive.Application.BundleIdentifier()
	bundleIDProfileMap[bundleID] = archive.Application.ProvisioningProfile

	for _, plugin := range archive.Application.Plugins {
		bundleID := plugin.BundleIdentifier()
		bundleIDProfileMap[bundleID] = plugin.ProvisioningProfile
	}

	if archive.Application.WatchApplication != nil {
		watchApplication := *archive.Application.WatchApplication

		bundleID := watchApplication.BundleIdentifier()
		bundleIDProfileMap[bundleID] = watchApplication.ProvisioningProfile

		for _, plugin := range watchApplication.Plugins {
			bundleID := plugin.BundleIdentifier()
			bundleIDProfileMap[bundleID] = plugin.ProvisioningProfile
		}
	}

	return bundleIDProfileMap
}

// FindDSYMs ...
func (archive XCArchive) FindDSYMs() (string, []string, error) {
	dsymsDirPth := filepath.Join(archive.Path, "dSYMs")
	dsyms, err := utility.ListEntries(dsymsDirPth, utility.ExtensionFilter(".dsym", true))
	if err != nil {
		return "", []string{}, err
	}

	appDSYM := ""
	frameworkDSYMs := []string{}
	for _, dsym := range dsyms {
		if strings.HasSuffix(dsym, ".app.dSYM") {
			appDSYM = dsym
		} else {
			frameworkDSYMs = append(frameworkDSYMs, dsym)
		}
	}
	if appDSYM == "" && len(frameworkDSYMs) == 0 {
		return "", []string{}, fmt.Errorf("no dsym found")
	}

	return appDSYM, frameworkDSYMs, nil
}

// NewXCArchive ...
func NewXCArchive(xcarchivePth string) (XCArchive, error) {
	application := Application{}
	{
		applicationsDir := filepath.Join(xcarchivePth, "Products/Applications")
		exist, err := pathutil.IsDirExists(applicationsDir)
		if err != nil {
			return XCArchive{}, err
		} else if !exist {
			return XCArchive{}, fmt.Errorf("Applications dir does not exist at: %s", applicationsDir)
		}

		application, err = NewApplication(applicationsDir)
		if err != nil {
			return XCArchive{}, err
		}

	}

	infoPlist := plistutil.PlistData{}
	{
		infoPlistPth := filepath.Join(xcarchivePth, "Info.plist")
		exist, err := pathutil.IsPathExists(infoPlistPth)
		if err != nil {
			return XCArchive{}, err
		} else if !exist {
			return XCArchive{}, fmt.Errorf("Info.plist does not exist at: %s", infoPlistPth)
		}
		infoPlist, err = plistutil.NewPlistDataFromFile(infoPlistPth)
		if err != nil {
			return XCArchive{}, err
		}
	}

	return XCArchive{
		Path:        xcarchivePth,
		Application: application,
		InfoPlist:   infoPlist,
	}, nil
}
