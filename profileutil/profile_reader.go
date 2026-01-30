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

// ProfileReader ...
type ProfileReader struct {
	logger       log.Logger
	fileManager  fileutil.FileManager
	pathModifier pathutil.PathModifier
	pathProvider pathutil.PathProvider
	pathChecker  pathutil.PathChecker
}

// NewProfileReader ...
func NewProfileReader(logger log.Logger, fileManager fileutil.FileManager, pathModifier pathutil.PathModifier, pathProvider pathutil.PathProvider, pathChecker pathutil.PathChecker) ProfileReader {
	return ProfileReader{
		logger:       logger,
		fileManager:  fileManager,
		pathModifier: pathModifier,
		pathProvider: pathProvider,
		pathChecker:  pathChecker,
	}
}

// ProvisioningProfileInfoFromFile ...
func (reader ProfileReader) ProvisioningProfileInfoFromFile(pth string) (ProvisioningProfileInfoModel, error) {
	provisioningProfile, err := reader.provisioningProfileFromFile(pth)
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
	provisioningProfiles, err := reader.installedProvisioningProfiles(profileType)
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

func (reader ProfileReader) provisioningProfileFromFile(pth string) (*pkcs7.PKCS7, error) {
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
	return pkcs7.Parse(content)
}

func (reader ProfileReader) installedProvisioningProfiles(profileType ProfileType) ([]*pkcs7.PKCS7, error) {
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
		profile, err := reader.provisioningProfileFromFile(pth)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}
