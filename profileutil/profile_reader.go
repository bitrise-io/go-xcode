package profileutil

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/fullsailor/pkcs7"
)

// ProvProfileSystemDirPath ...
const ProvProfileSystemDirPath = "~/Library/MobileDevice/Provisioning Profiles"

type ProfileReader struct {
	logger       log.Logger
	fileManager  fileutil.FileManager
	pathModifier pathutil.PathModifier
	pathProvider pathutil.PathProvider
	pathChecker  pathutil.PathChecker
}

func NewProfileReader(logger log.Logger, fileManager fileutil.FileManager, pathModifier pathutil.PathModifier, pathProvider pathutil.PathProvider, pathChecker pathutil.PathChecker) ProfileReader {
	return ProfileReader{
		logger:       logger,
		fileManager:  fileManager,
		pathModifier: pathModifier,
		pathProvider: pathProvider,
		pathChecker:  pathChecker,
	}
}

// ProvisioningProfileFromFile ...
func (reader ProfileReader) ProvisioningProfileFromFile(pth string) (*pkcs7.PKCS7, error) {
	f, err := reader.fileManager.Open(pth)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			reader.logger.Warnf("Failed to close file %s, error: %s", pth, err)
		}
	}()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return ProvisioningProfileFromContent(content)
}

// InstalledProvisioningProfiles ...
func (reader ProfileReader) InstalledProvisioningProfiles(profileType ProfileType) ([]*pkcs7.PKCS7, error) {
	ext := ".mobileprovision"
	if profileType == ProfileTypeMacOs {
		ext = ".provisionprofile"
	}

	absProvProfileDirPath, err := reader.pathModifier.AbsPath(ProvProfileSystemDirPath)
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(reader.pathModifier.EscapeGlobPath(absProvProfileDirPath), "*"+ext)
	pths, err := reader.pathProvider.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var profiles []*pkcs7.PKCS7
	for _, pth := range pths {
		profile, err := reader.ProvisioningProfileFromFile(pth)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// FindProvisioningProfile ...
func (reader ProfileReader) FindProvisioningProfile(uuid string) (*pkcs7.PKCS7, string, error) {
	absProvProfileDirPath, err := reader.pathModifier.AbsPath(ProvProfileSystemDirPath)
	if err != nil {
		return nil, "", err
	}

	iosProvisioningProfileExt := ".mobileprovision"
	pth := filepath.Join(absProvProfileDirPath, uuid+iosProvisioningProfileExt)
	if exist, err := reader.pathChecker.IsPathExists(pth); err != nil {
		return nil, "", err
	} else if exist {
		profile, err := reader.ProvisioningProfileFromFile(pth)
		if err != nil {
			return nil, "", err
		}
		return profile, pth, nil
	}

	macOsProvisioningProfileExt := ".provisionprofile"
	pth = filepath.Join(absProvProfileDirPath, uuid+macOsProvisioningProfileExt)
	if exist, err := reader.pathChecker.IsPathExists(pth); err != nil {
		return nil, "", err
	} else if exist {
		profile, err := reader.ProvisioningProfileFromFile(pth)
		if err != nil {
			return nil, "", err
		}
		return profile, pth, nil
	}

	return nil, "", nil
}

// ProvisioningProfileInfoFromFile ...
func (reader ProfileReader) ProvisioningProfileInfoFromFile(pth string) (ProvisioningProfileInfoModel, error) {
	provisioningProfile, err := reader.ProvisioningProfileFromFile(pth)
	if err != nil {
		return ProvisioningProfileInfoModel{}, err
	}
	if provisioningProfile != nil {
		return NewProvisioningProfileInfo(*provisioningProfile)
	}
	return ProvisioningProfileInfoModel{}, errors.New("failed to parse provisioning profile infos")
}

// InstalledProvisioningProfileInfos ...
func (reader ProfileReader) InstalledProvisioningProfileInfos(profileType ProfileType) ([]ProvisioningProfileInfoModel, error) {
	provisioningProfiles, err := reader.InstalledProvisioningProfiles(profileType)
	if err != nil {
		return nil, err
	}

	var infos []ProvisioningProfileInfoModel
	for _, provisioningProfile := range provisioningProfiles {
		if provisioningProfile != nil {
			info, err := NewProvisioningProfileInfo(*provisioningProfile)
			if err != nil {
				return nil, err
			}
			infos = append(infos, info)
		}
	}
	return infos, nil
}

// ProvisioningProfileFromContent ...
func ProvisioningProfileFromContent(content []byte) (*pkcs7.PKCS7, error) {
	return pkcs7.Parse(content)
}
