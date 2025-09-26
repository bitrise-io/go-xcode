package localcodesignasset

import (
	"github.com/bitrise-io/go-utils/v2/fileutil"
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
	profileProvider := profileutil.NewProfileProvider(fileutil.NewFileManager(), pathutil.NewPathModifier(), pathutil.NewPathChecker())
	return profileProvider.InstalledProvisioningProfileInfos(profileutil.ProfileTypeIos)
}
