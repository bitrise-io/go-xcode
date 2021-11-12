package localcodesignasset

import "github.com/bitrise-io/go-xcode/profileutil"

// ProvisioningProfileProvider can list profile infos.
type ProvisioningProfileProvider interface {
	ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error)
}

// LocalProvisioningProfileProvider ...
type LocalProvisioningProfileProvider struct{}

// ListProvisioningProfiles ...
func (p LocalProvisioningProfileProvider) ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error) {
	return profileutil.InstalledProvisioningProfileInfos(profileutil.ProfileTypeIos)
}
