package export

import (
	"sort"

	"github.com/bitrise-tools/go-xcode/certificateutil"
	"github.com/bitrise-tools/go-xcode/profileutil"
	"github.com/ryanuber/go-glob"
)

// SelectableCodeSignGroup ..
type SelectableCodeSignGroup struct {
	Certificate         certificateutil.CertificateInfoModel
	BundleIDProfilesMap map[string][]profileutil.ProvisioningProfileInfoModel
}

func isCertificateInstalled(installedCertificates []certificateutil.CertificateInfoModel, certificate certificateutil.CertificateInfoModel) bool {
	installedMap := map[string]bool{}
	for _, certificate := range installedCertificates {
		installedMap[certificate.Serial] = true
	}
	return installedMap[certificate.Serial]
}

// CreateSelectableCodeSignGroups ...
func CreateSelectableCodeSignGroups(certificates []certificateutil.CertificateInfoModel, profiles []profileutil.ProvisioningProfileInfoModel, bundleIDs []string) []SelectableCodeSignGroup {
	groups := []SelectableCodeSignGroup{}

	serialProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
	serialCertificateMap := map[string]certificateutil.CertificateInfoModel{}
	for _, profile := range profiles {
		for _, certificate := range profile.DeveloperCertificates {
			if !isCertificateInstalled(certificates, certificate) {
				continue
			}

			certificateProfiles := serialProfilesMap[certificate.Serial]
			if certificateProfiles == nil {
				certificateProfiles = []profileutil.ProvisioningProfileInfoModel{}
			}
			certificateProfiles = append(certificateProfiles, profile)
			serialProfilesMap[certificate.Serial] = certificateProfiles
			serialCertificateMap[certificate.Serial] = certificate
		}
	}

	for serial, profiles := range serialProfilesMap {
		certificate := serialCertificateMap[serial]

		bundleIDProfilesMap := map[string][]profileutil.ProvisioningProfileInfoModel{}
		for _, bundleID := range bundleIDs {

			matchingProfiles := []profileutil.ProvisioningProfileInfoModel{}
			for _, profile := range profiles {
				if !glob.Glob(profile.BundleID, bundleID) {
					continue
				}

				matchingProfiles = append(matchingProfiles, profile)
			}

			if len(matchingProfiles) > 0 {
				sort.Sort(ByBundleIDLength(matchingProfiles))
				bundleIDProfilesMap[bundleID] = matchingProfiles
			}
		}

		if len(bundleIDProfilesMap) == len(bundleIDs) {
			group := SelectableCodeSignGroup{
				Certificate:         certificate,
				BundleIDProfilesMap: bundleIDProfilesMap,
			}
			groups = append(groups, group)
		}
	}

	return groups
}

// ByBundleIDLength ...
type ByBundleIDLength []profileutil.ProvisioningProfileInfoModel

// Len ..
func (s ByBundleIDLength) Len() int {
	return len(s)
}

// Swap ...
func (s ByBundleIDLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less ...
func (s ByBundleIDLength) Less(i, j int) bool {
	return len(s[i].BundleID) > len(s[j].BundleID)
}
