package profileutil

import (
	"crypto/x509"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/fullsailor/pkcs7"
)

// ProvProfileSystemDirPaths ...
const (
	provProfileInMobileDevicesDirPath = "~/Library/MobileDevice/Provisioning Profiles"
	provProfileInXcodeUserDataDirPath = "~/Library/Developer/Xcode/UserData/Provisioning Profiles"
)

// ProfileType ...
type ProfileType string

// ProfileTypes ...
const (
	ProfileTypeIos   ProfileType = "ios"
	ProfileTypeMacOs ProfileType = "osx"
	ProfileTypeTvOs  ProfileType = "tvos"
)

const (
	iOSProfileExtension   string = ".mobileprovision"
	macOSProfileExtension string = ".provisionprofile"
)

type ProfileProvider struct {
	fileManager  fileutil.FileManager
	pathModifier pathutil.PathModifier
	pathChecker  pathutil.PathChecker
}

func NewProfileProvider(fileManager fileutil.FileManager, pathModifier pathutil.PathModifier, pathChecker pathutil.PathChecker) ProfileProvider {
	return ProfileProvider{fileManager: fileManager, pathModifier: pathModifier, pathChecker: pathChecker}
}

// ProvisioningProfileFromFile ...
func (p ProfileProvider) ProvisioningProfileFromFile(pth string) (*pkcs7.PKCS7, error) {
	content, err := p.fileManager.ReadFile(pth)
	if err != nil {
		return nil, err
	}
	return ProvisioningProfileFromContent(content)
}

// FindProvisioningProfile ...
func (p ProfileProvider) FindProvisioningProfile(uuid string, profileType ProfileType) (*pkcs7.PKCS7, string, error) {
	ext := extensionForProfileType(profileType)

	for _, provProfileDir := range []string{provProfileInMobileDevicesDirPath, provProfileInXcodeUserDataDirPath} {
		absProvProfileDirPath, err := p.pathModifier.AbsPath(provProfileDir)
		if err != nil {
			return nil, "", err
		}

		pth := filepath.Join(absProvProfileDirPath, uuid+ext)
		if exist, err := p.pathChecker.IsPathExists(pth); err != nil {
			return nil, "", err
		} else if exist {
			profile, err := p.ProvisioningProfileFromFile(pth)
			if err != nil {
				return nil, "", err
			}
			return profile, pth, nil
		}
	}

	return nil, "", nil
}

// InstalledProvisioningProfiles ...
func (p ProfileProvider) InstalledProvisioningProfiles(profileType ProfileType) ([]*pkcs7.PKCS7, error) {
	ext := extensionForProfileType(profileType)

	var profiles []*pkcs7.PKCS7
	for _, provProfileDir := range []string{provProfileInMobileDevicesDirPath, provProfileInXcodeUserDataDirPath} {
		absProvProfileDirPath, err := p.pathModifier.AbsPath(provProfileDir)
		if err != nil {
			return nil, err
		}

		pattern := filepath.Join(p.pathModifier.EscapeGlobPath(absProvProfileDirPath), "*"+ext)
		pths, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, pth := range pths {
			profile, err := p.ProvisioningProfileFromFile(pth)
			if err != nil {
				return nil, err
			}
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// InstalledProvisioningProfileInfos ...
func (p ProfileProvider) InstalledProvisioningProfileInfos(profileType ProfileType) ([]ProvisioningProfileInfoModel, error) {
	provisioningProfiles, err := p.InstalledProvisioningProfiles(profileType)
	if err != nil {
		return nil, err
	}

	infos := []ProvisioningProfileInfoModel{}
	for _, provisioningProfile := range provisioningProfiles {
		if provisioningProfile != nil {
			info, err := ProvisioningProfileInfoFromPKCS7(*provisioningProfile)
			if err != nil {
				return nil, err
			}
			infos = append(infos, info)
		}
	}
	return infos, nil
}

// ProvisioningProfileInfoFromFile ...
func (p ProfileProvider) ProvisioningProfileInfoFromFile(pth string) (ProvisioningProfileInfoModel, error) {
	provisioningProfile, err := p.ProvisioningProfileFromFile(pth)
	if err != nil {
		return ProvisioningProfileInfoModel{}, err
	}
	if provisioningProfile != nil {
		return ProvisioningProfileInfoFromPKCS7(*provisioningProfile)
	}
	return ProvisioningProfileInfoModel{}, errors.New("failed to parse provisioning profile infos")
}

// ProvisioningProfileFromContent ...
func ProvisioningProfileFromContent(content []byte) (*pkcs7.PKCS7, error) {
	return pkcs7.Parse(content)
}

// ProvisioningProfileInfoFromPKCS7 ...
func ProvisioningProfileInfoFromPKCS7(provisioningProfile pkcs7.PKCS7) (ProvisioningProfileInfoModel, error) {
	var data plistutil.PlistData
	if _, err := plist.Unmarshal(provisioningProfile.Content, &data); err != nil {
		return ProvisioningProfileInfoModel{}, err
	}

	platforms, _ := data.GetStringArray("Platform")
	if len(platforms) == 0 {
		return ProvisioningProfileInfoModel{}, fmt.Errorf("missing Platform array in profile")
	}

	platform := strings.ToLower(platforms[0])
	var profileType ProfileType

	switch platform {
	case string(ProfileTypeIos):
		profileType = ProfileTypeIos
	case string(ProfileTypeMacOs):
		profileType = ProfileTypeMacOs
	case string(ProfileTypeTvOs):
		profileType = ProfileTypeTvOs
	default:
		return ProvisioningProfileInfoModel{}, fmt.Errorf("unknown platform type: %s", platform)
	}

	profile := PlistData(data)
	info := ProvisioningProfileInfoModel{
		UUID:                 profile.GetUUID(),
		Name:                 profile.GetName(),
		TeamName:             profile.GetTeamName(),
		TeamID:               profile.GetTeamID(),
		BundleID:             profile.GetBundleIdentifier(),
		CreationDate:         profile.GetCreationDate(),
		ExpirationDate:       profile.GetExpirationDate(),
		ProvisionsAllDevices: profile.GetProvisionsAllDevices(),
		Type:                 profileType,
	}

	info.ExportType = profile.GetExportMethod()

	if devicesList := profile.GetProvisionedDevices(); devicesList != nil {
		info.ProvisionedDevices = devicesList
	}

	developerCertificates, found := data.GetByteArrayArray("DeveloperCertificates")
	if found {
		certificates := []*x509.Certificate{}
		for _, certificateBytes := range developerCertificates {
			certificate, err := certificateutil.CertificateFromDERContent(certificateBytes)
			if err == nil && certificate != nil {
				certificates = append(certificates, certificate)
			}
		}

		for _, certificate := range certificates {
			if certificate != nil {
				info.DeveloperCertificates = append(info.DeveloperCertificates, certificateutil.NewCertificateInfo(*certificate, nil))
			}
		}
	}

	info.Entitlements = profile.GetEntitlements()

	return info, nil
}

func extensionForProfileType(profileType ProfileType) string {
	ext := iOSProfileExtension
	if profileType == ProfileTypeMacOs {
		ext = macOSProfileExtension
	}
	return ext
}
