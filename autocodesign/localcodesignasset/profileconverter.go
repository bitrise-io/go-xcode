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
	// TODO: wire in as a dep on the struct
	profileReader := profileutil.NewProfileReader(log.NewLogger(), fileutil.NewFileManager(), pathutil.NewPathModifier(), pathutil.NewPathProvider(), pathutil.NewPathChecker())
	_, pth, err := profileReader.FindProvisioningProfile(info.UUID)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(pth)
	if err != nil {
		return nil, err
	}

	return NewProfile(info, content), nil
}
