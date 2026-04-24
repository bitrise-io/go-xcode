package localcodesignasset

import (
	"io"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// ProvisioningProfileConverter ...
type ProvisioningProfileConverter interface {
	ProfileInfoToProfile(info profileutil.ProvisioningProfileInfoModel) (autocodesign.Profile, error)
}

type provisioningProfileConverter struct {
	fileManager   fileutil.FileManager
	profileReader profileutil.ProfileReader
}

// NewProvisioningProfileConverter ...
func NewProvisioningProfileConverter() ProvisioningProfileConverter {
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	pathModifier := pathutil.NewPathModifier()
	pathProvider := pathutil.NewPathProvider()
	profileReader := profileutil.NewProfileReader(logger, fileManager, pathModifier, pathProvider)

	return provisioningProfileConverter{
		fileManager:   fileManager,
		profileReader: profileReader,
	}
}

// ProfileInfoToProfile ...
func (c provisioningProfileConverter) ProfileInfoToProfile(info profileutil.ProvisioningProfileInfoModel) (autocodesign.Profile, error) {
	pth, err := c.findProvisioningProfile(info.UUID)
	if err != nil {
		return nil, err
	}
	profile, err := c.fileManager.Open(pth)
	if err != nil {
		return nil, err
	}
	content, err := io.ReadAll(profile)
	if err != nil {
		return nil, err
	}

	return NewProfile(info, content), nil
}

func (c provisioningProfileConverter) findProvisioningProfile(uuid string) (string, error) {
	paths, err := c.profileReader.ListProfiles(profileutil.ProfileTypeIos, uuid)
	if err != nil {
		return "", err
	}
	macOSPaths, err := c.profileReader.ListProfiles(profileutil.ProfileTypeMacOs, uuid)
	if err != nil {
		return "", err
	}

	paths = append(paths, macOSPaths...)
	if len(paths) == 0 {
		// ToDo return error of not found, keeping the nil return values for backward compatibility for now
		return "", nil
	}

	_, err = c.profileReader.ProvisioningProfileInfoFromFile(paths[0])
	if err != nil {
		return "", err
	}
	return paths[0], nil
}
