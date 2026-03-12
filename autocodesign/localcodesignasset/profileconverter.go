package localcodesignasset

import (
	"os"
	"path/filepath"

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
	// TODO: wire in as a dep on the struct
	pathModifier := pathutil.NewPathModifier()
	pathChecker := pathutil.NewPathChecker()

	absProvProfileDirPath, err := pathModifier.AbsPath(profileutil.ProvProfileSystemDirPath)
	if err != nil {
		return "", err
	}

	iosProvisioningProfileExt := ".mobileprovision"
	pth := filepath.Join(absProvProfileDirPath, uuid+iosProvisioningProfileExt)
	if exist, err := pathChecker.IsPathExists(pth); err != nil {
		return "", err
	} else if exist {
		return pth, nil
	}

	macOsProvisioningProfileExt := ".provisionprofile"
	pth = filepath.Join(absProvProfileDirPath, uuid+macOsProvisioningProfileExt)
	if exist, err := pathChecker.IsPathExists(pth); err != nil {
		return "", err
	} else if exist {
		return pth, nil
	}

	return "", nil
}
