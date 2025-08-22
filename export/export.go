package export

import (
	"fmt"
	"sort"

	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
	"github.com/ryanuber/go-glob"
)

// SelectableCodeSignGroup ...
type SelectableCodeSignGroup struct {
	Certificate         certificateutil.CertificateInfoModel
	BundleIDProfilesMap map[string][]profileutil.ProvisioningProfileInfoModel
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

			certificateProfiles, ok := serialProfilesMap[certificate.Serial]
			if !ok {
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

// String ...
func (group SelectableCodeSignGroup) String() string {
	result := fmt.Sprintf("Team: %s (%s)\n", group.Certificate.TeamName, group.Certificate.TeamID)
	result += fmt.Sprintf("Certificate: %s (%s)\n", group.Certificate.CommonName, group.Certificate.Serial)

	result += "Bundle ID - Profiles mapping:\n"
	for bundleID, profileInfos := range group.BundleIDProfilesMap {
		result += fmt.Sprintf("- %s:\n", bundleID)
		for _, profileInfo := range profileInfos {
			result += fmt.Sprintf("  - %s (%s)\n", profileInfo.Name, profileInfo.UUID)
		}
	}

	return result
}

func isCertificateInstalled(installedCertificates []certificateutil.CertificateInfoModel, certificate certificateutil.CertificateInfoModel) bool {
	for _, cert := range installedCertificates {
		if cert.Serial == certificate.Serial {
			return true
		}
	}
	return false
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
