package localcodesignasset

import (
	"os"

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
}

// NewProvisioningProfileConverter ...
func NewProvisioningProfileConverter() ProvisioningProfileConverter {
	return provisioningProfileConverter{}
}

// ProfileInfoToProfile ...
func (c provisioningProfileConverter) ProfileInfoToProfile(info profileutil.ProvisioningProfileInfoModel) (autocodesign.Profile, error) {
	pth, err := findProvisioningProfile(info.UUID)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}

	return NewProfile(info, content), nil
}

func findProvisioningProfile(uuid string) (string, error) {
	// TODO: wire deps on ProvisioningProfileConverter
	profileReader := profileutil.NewProfileReader(log.NewLogger(), fileutil.NewFileManager(), pathutil.NewPathModifier(), pathutil.NewPathProvider())
	paths, err := profileReader.ListProfiles(profileutil.ProfileTypeIos, uuid)
	if err != nil {
		return "", err
	}
	macOSPaths, err := profileReader.ListProfiles(profileutil.ProfileTypeMacOs, uuid)
	if err != nil {
		return "", err
	}

	paths = append(paths, macOSPaths...)
	if len(paths) == 0 {
		// ToDo return error of not found, keeping the nil return values for backward compatibility for now
		return "", nil
	}

	_, err = profileReader.ProvisioningProfileInfoFromFile(paths[0])
	if err != nil {
		return "", err
	}
	return paths[0], nil
}
