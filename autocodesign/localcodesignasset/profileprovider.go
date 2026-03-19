package localcodesignasset

import (
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// ProvisioningProfileProvider can list profile infos.
type ProvisioningProfileProvider interface {
	ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error)
}

type provisioningProfileProvider struct{}

// NewProvisioningProfileProvider ...
func NewProvisioningProfileProvider() ProvisioningProfileProvider {
	return provisioningProfileProvider{}
}

// ListProvisioningProfiles ...
func (p provisioningProfileProvider) ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error) {
	// TODO: wire deps on provisioningProfileProvider
	profileReader := profileutil.NewProfileReader(log.NewLogger(), fileutil.NewFileManager(), pathutil.NewPathModifier(), pathutil.NewPathProvider())
	return profileReader.InstalledProvisioningProfileInfos(profileutil.ProfileTypeIos)
}
