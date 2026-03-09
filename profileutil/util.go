package profileutil

import (
	"path/filepath"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/fullsailor/pkcs7"
)

// ProfileType ...
type ProfileType string

// ProfileTypeIos ...
const ProfileTypeIos ProfileType = "ios"

// ProfileTypeMacOs ...
const ProfileTypeMacOs ProfileType = "osx"

// ProfileTypeTvOs ...
const ProfileTypeTvOs ProfileType = "tvos"

const (
	// ProvProfileSystemDirPath ...
	ProvProfileSystemDirPath = "~/Library/MobileDevice/Provisioning Profiles"
	// ProvProfileModernPath is used by Xcode 16 and later, but the old path is still supported.
	ProvProfileModernPath = "~/Library/Developer/Xcode/UserData/Provisioning Profiles"
)

// ProvisioningProfileFromContent ...
func ProvisioningProfileFromContent(content []byte) (*pkcs7.PKCS7, error) {
	return pkcs7.Parse(content)
}

// ProvisioningProfileFromFile ...
func ProvisioningProfileFromFile(pth string) (*pkcs7.PKCS7, error) {
	content, err := fileutil.ReadBytesFromFile(pth)
	if err != nil {
		return nil, err
	}
	return ProvisioningProfileFromContent(content)
}

func listProfiles(profileType ProfileType) ([]string, error) {
	ext := ".mobileprovision"
	if profileType == ProfileTypeMacOs {
		ext = ".provisionprofile"
	}

	absProvProfileDirPath, err := pathutil.AbsPath(ProvProfileSystemDirPath)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(pathutil.EscapeGlobPath(absProvProfileDirPath), "*"+ext)
	pths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	absProvProfileXcode16DirPath, err := pathutil.AbsPath(ProvProfileModernPath)
	if err != nil {
		return nil, err
	}

	pattern = filepath.Join(pathutil.EscapeGlobPath(absProvProfileXcode16DirPath), "*"+ext)
	newPaths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	pths = append(pths, newPaths...)
	return pths, nil
}

// InstalledProvisioningProfiles ...
func InstalledProvisioningProfiles(profileType ProfileType) ([]*pkcs7.PKCS7, error) {
	pths, err := listProfiles(profileType)
	if err != nil {
		return nil, err
	}

	profiles := []*pkcs7.PKCS7{}
	for _, pth := range pths {
		profile, err := ProvisioningProfileFromFile(pth)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// FindProvisioningProfile ...
func FindProvisioningProfile(uuid string) (*pkcs7.PKCS7, string, error) {
	{
		pths, err := listProfiles(ProfileTypeIos)
		if err != nil {
			return nil, "", err
		}

		profileName := uuid + ".mobileprovision"
		for _, pth := range pths {
			if filepath.Base(pth) == profileName {
				profile, err := ProvisioningProfileFromFile(pth)
				if err != nil {
					return nil, "", err
				}
				return profile, pth, nil
			}
		}
	}

	{
		pths, err := listProfiles(ProfileTypeMacOs)
		if err != nil {
			return nil, "", err
		}

		profileName := uuid + ".provisionprofile"
		for _, pth := range pths {
			if filepath.Base(pth) == profileName {
				profile, err := ProvisioningProfileFromFile(pth)
				if err != nil {
					return nil, "", err
				}
				return profile, pth, nil
			}
		}
	}

	return nil, "", nil
}
