package xcarchive

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

type IosBaseApplication struct {
	Path                string
	InfoPlist           plistutil.PlistData
	Entitlements        plistutil.PlistData
	ProvisioningProfile profileutil.ProvisioningProfileInfoModel
}

func (app IosBaseApplication) BundleIdentifier() string {
	bundleID, _ := app.InfoPlist.GetString("CFBundleIdentifier")
	return bundleID
}

func NewIosBaseApplication(path string) (IosBaseApplication, error) {
	// TODO: wire in as a dep on the struct
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	pathModifier := pathutil.NewPathModifier()
	pathProvider := pathutil.NewPathProvider()
	pathChecker := pathutil.NewPathChecker()
	envRepo := env.NewRepository()
	cmdFactory := command.NewFactory(envRepo)
	profileReader := profileutil.NewProfileReader(logger, fileManager, pathModifier, pathProvider)

	var infoPlist plistutil.PlistData
	{
		infoPlistPath := filepath.Join(path, "Info.plist")
		if exist, err := pathChecker.IsPathExists(infoPlistPath); err != nil {
			return IosBaseApplication{}, fmt.Errorf("failed to check if Info.plist exists at: %s, error: %s", infoPlistPath, err)
		} else if !exist {
			return IosBaseApplication{}, fmt.Errorf("Info.plist not exists at: %s", infoPlistPath)
		}
		plist, err := newPlistDataFromFile(infoPlistPath)
		if err != nil {
			return IosBaseApplication{}, err
		}
		infoPlist = plist
	}

	var provisioningProfile profileutil.ProvisioningProfileInfoModel
	{
		provisioningProfilePath := filepath.Join(path, "embedded.mobileprovision")
		if exist, err := pathChecker.IsPathExists(provisioningProfilePath); err != nil {
			return IosBaseApplication{}, fmt.Errorf("failed to check if profile exists at: %s, error: %s", provisioningProfilePath, err)
		} else if !exist {
			return IosBaseApplication{}, fmt.Errorf("profile not exists at: %s", provisioningProfilePath)
		}

		profile, err := profileReader.ProvisioningProfileInfoFromFile(provisioningProfilePath)
		if err != nil {
			return IosBaseApplication{}, err
		}
		provisioningProfile = profile
	}

	executable := executableNameFromInfoPlist(infoPlist)
	entitlements, err := getEntitlements(cmdFactory, path, executable)
	if err != nil {
		return IosBaseApplication{}, err
	}

	return IosBaseApplication{
		Path:                path,
		InfoPlist:           infoPlist,
		Entitlements:        entitlements,
		ProvisioningProfile: provisioningProfile,
	}, nil
}

type IosExtension struct {
	IosBaseApplication
}

func NewIosExtension(path string) (IosExtension, error) {
	baseApp, err := NewIosBaseApplication(path)
	if err != nil {
		return IosExtension{}, err
	}

	return IosExtension{
		baseApp,
	}, nil
}

type IosWatchApplication struct {
	IosBaseApplication
	Extensions []IosExtension
}

type IosClipApplication struct {
	IosBaseApplication
}

func NewIosWatchApplication(path string) (IosWatchApplication, error) {

	baseApp, err := NewIosBaseApplication(path)
	if err != nil {
		return IosWatchApplication{}, err
	}

	extensions := []IosExtension{}
	pattern := filepath.Join(escapeGlobPath(path), "PlugIns/*.appex")
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
		IosBaseApplication: baseApp,
		Extensions:         extensions,
	}, nil
}

func NewIosClipApplication(path string) (IosClipApplication, error) {
	baseApp, err := NewIosBaseApplication(path)
	if err != nil {
		return IosClipApplication{}, err
	}

	return IosClipApplication{
		IosBaseApplication: baseApp,
	}, nil
}

type IosApplication struct {
	IosBaseApplication
	WatchApplication *IosWatchApplication
	ClipApplication  *IosClipApplication
	Extensions       []IosExtension
}

func NewIosApplication(path string) (IosApplication, error) {
	baseApp, err := NewIosBaseApplication(path)
	if err != nil {
		return IosApplication{}, err
	}

	var watchApp *IosWatchApplication
	{
		pattern := filepath.Join(escapeGlobPath(path), "Watch/*.app")
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

	var clipApp *IosClipApplication
	{
		pattern := filepath.Join(escapeGlobPath(path), "AppClips/*.app")
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return IosApplication{}, err
		}
		if len(pths) > 0 {
			clipPath := pths[0]
			app, err := NewIosClipApplication(clipPath)
			if err != nil {
				return IosApplication{}, err
			}
			clipApp = &app
		}
	}

	extensions := []IosExtension{}
	{
		pattern := filepath.Join(escapeGlobPath(path), "PlugIns/*.appex")
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
		IosBaseApplication: baseApp,
		WatchApplication:   watchApp,
		ClipApplication:    clipApp,
		Extensions:         extensions,
	}, nil
}

type IosArchive struct {
	Path        string
	InfoPlist   plistutil.PlistData
	Application IosApplication
}

func NewIosArchive(path string) (IosArchive, error) {
	pathChecker := pathutil.NewPathChecker()

	var infoPlist plistutil.PlistData
	{
		infoPlistPath := filepath.Join(path, "Info.plist")
		if exist, err := pathChecker.IsPathExists(infoPlistPath); err != nil {
			return IosArchive{}, fmt.Errorf("failed to check if Info.plist exists at: %s, error: %s", infoPlistPath, err)
		} else if !exist {
			return IosArchive{}, fmt.Errorf("Info.plist not exists at: %s", infoPlistPath)
		}
		plist, err := newPlistDataFromFile(infoPlistPath)
		if err != nil {
			return IosArchive{}, err
		}
		infoPlist = plist
	}

	var application IosApplication
	{
		appPath := ""
		if appRelativePathToProducts, found := applicationFromPlist(infoPlist); found {
			appPath = filepath.Join(path, "Products", appRelativePathToProducts)
		} else {
			var err error
			if appPath, err = applicationFromArchive(path); err != nil {
				return IosArchive{}, err
			}
		}
		if exist, err := pathChecker.IsPathExists(appPath); err != nil {
			return IosArchive{}, fmt.Errorf("failed to check if app exists, path: %s, error: %s", appPath, err)
		} else if !exist {
			return IosArchive{}, fmt.Errorf("application not found on path: %s, error: %s", appPath, err)
		}

		app, err := NewIosApplication(appPath)
		if err != nil {
			return IosArchive{}, err
		}
		application = app
	}

	return IosArchive{
		Path:        path,
		InfoPlist:   infoPlist,
		Application: application,
	}, nil
}

func applicationFromPlist(InfoPlist plistutil.PlistData) (string, bool) {
	if properties, found := InfoPlist.GetMapStringInterface("ApplicationProperties"); found {
		return properties.GetString("ApplicationPath")
	}
	return "", false
}

func applicationFromArchive(path string) (string, error) {
	pattern := filepath.Join(escapeGlobPath(path), "Products/Applications/*.app")
	pths, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(pths) == 0 {
		return "", fmt.Errorf("failed to find main app, using pattern: %s", pattern)
	}
	return pths[0], nil
}

func (archive IosArchive) IsXcodeManaged() bool {
	return archive.Application.ProvisioningProfile.IsXcodeManaged()
}

func (archive IosArchive) IsSigningManagedAutomatically() (bool, error) {
	return archive.IsXcodeManaged(), nil
}

func (archive IosArchive) SigningIdentity() string {
	if properties, found := archive.InfoPlist.GetMapStringInterface("ApplicationProperties"); found {
		identity, _ := properties.GetString("SigningIdentity")
		return identity
	}
	return ""
}

func (archive IosArchive) BundleIDEntitlementsMap() map[string]plistutil.PlistData {
	bundleIDEntitlementsMap := map[string]plistutil.PlistData{}

	bundleID := archive.Application.BundleIdentifier()
	bundleIDEntitlementsMap[bundleID] = archive.Application.Entitlements

	for _, plugin := range archive.Application.Extensions {
		bundleID := plugin.BundleIdentifier()
		bundleIDEntitlementsMap[bundleID] = plugin.Entitlements
	}

	if archive.Application.WatchApplication != nil {
		watchApplication := *archive.Application.WatchApplication

		bundleID := watchApplication.BundleIdentifier()
		bundleIDEntitlementsMap[bundleID] = watchApplication.Entitlements

		for _, plugin := range watchApplication.Extensions {
			bundleID := plugin.BundleIdentifier()
			bundleIDEntitlementsMap[bundleID] = plugin.Entitlements
		}
	}

	if archive.Application.ClipApplication != nil {
		clipApplication := *archive.Application.ClipApplication

		bundleID := clipApplication.BundleIdentifier()
		bundleIDEntitlementsMap[bundleID] = clipApplication.Entitlements
	}

	return bundleIDEntitlementsMap
}

func (archive IosArchive) BundleIDProfileInfoMap() map[string]profileutil.ProvisioningProfileInfoModel {
	bundleIDProfileMap := map[string]profileutil.ProvisioningProfileInfoModel{}

	bundleID := archive.Application.BundleIdentifier()
	bundleIDProfileMap[bundleID] = archive.Application.ProvisioningProfile

	for _, plugin := range archive.Application.Extensions {
		bundleID := plugin.BundleIdentifier()
		bundleIDProfileMap[bundleID] = plugin.ProvisioningProfile
	}

	if archive.Application.WatchApplication != nil {
		watchApplication := *archive.Application.WatchApplication

		bundleID := watchApplication.BundleIdentifier()
		bundleIDProfileMap[bundleID] = watchApplication.ProvisioningProfile

		for _, plugin := range watchApplication.Extensions {
			bundleID := plugin.BundleIdentifier()
			bundleIDProfileMap[bundleID] = plugin.ProvisioningProfile
		}
	}

	if archive.Application.ClipApplication != nil {
		clipApplication := *archive.Application.ClipApplication

		bundleID := clipApplication.BundleIdentifier()
		bundleIDProfileMap[bundleID] = clipApplication.ProvisioningProfile
	}

	return bundleIDProfileMap
}

func (archive IosArchive) FindDSYMs() ([]string, []string, error) {
	return findDSYMs(archive.Path)
}

func (archive IosArchive) TeamID() (string, error) {
	bundleIDProfileInfoMap := archive.BundleIDProfileInfoMap()
	for _, profileInfo := range bundleIDProfileInfoMap {
		return profileInfo.TeamID, nil
	}
	return "", errors.New("team id not found")
}

func (archive IosArchive) Platform() (autocodesign.Platform, error) {
	platformName := archive.Application.InfoPlist["DTPlatformName"]
	switch platformName {
	case "iphoneos":
		return autocodesign.IOS, nil
	case "appletvos":
		return autocodesign.TVOS, nil
	default:
		return "", fmt.Errorf("unsupported platform found: %s", platformName)
	}
}

func (archive IosArchive) GetAppLayout(_ bool) (autocodesign.AppLayout, error) {
	platform, err := archive.Platform()
	if err != nil {
		return autocodesign.AppLayout{}, err
	}

	bundleIDEntitlementsMap := archive.BundleIDEntitlementsMap()

	fmt.Printf("Reading %v code sign entitlements", len(bundleIDEntitlementsMap))

	entitlementsMap := map[string]autocodesign.Entitlements{}
	for bundleID, entitlements := range bundleIDEntitlementsMap {
		entitlementsMap[bundleID] = autocodesign.Entitlements(entitlements)
	}

	return autocodesign.AppLayout{
		Platform:                               platform,
		EntitlementsByArchivableTargetBundleID: entitlementsMap,
		UITestTargetBundleIDs:                  nil,
	}, nil
}
