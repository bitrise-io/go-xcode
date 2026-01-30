package export

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/ryanuber/go-glob"
)

// SelectableCodeSignGroup ...
type SelectableCodeSignGroup struct {
	Certificate         certificateutil.CertificateInfoModel
	BundleIDProfilesMap map[string][]profileutil.ProvisioningProfileInfoModel
}

// String ...
func (group SelectableCodeSignGroup) String() string {
	printable := map[string]interface{}{}
	printable["team"] = fmt.Sprintf("%s (%s)", group.Certificate.TeamName, group.Certificate.TeamID)
	printable["certificate"] = fmt.Sprintf("%s (%s)", group.Certificate.CommonName, group.Certificate.Serial)

	bundleIDProfiles := map[string][]string{}
	for bundleID, profileInfos := range group.BundleIDProfilesMap {
		printableProfiles := []string{}
		for _, profileInfo := range profileInfos {
			printableProfiles = append(printableProfiles, fmt.Sprintf("%s (%s)", profileInfo.Name, profileInfo.UUID))
		}
		bundleIDProfiles[bundleID] = printableProfiles
	}
	printable["bundle_id_profiles"] = bundleIDProfiles

	data, err := json.MarshalIndent(printable, "", "\t")
	if err != nil {
		log.Errorf("Failed to marshal: %v, error: %s", printable, err)
		return ""
	}

	return string(data)
}

func containsCertificate(installedCertificates []certificateutil.CertificateInfoModel, certificate certificateutil.CertificateInfoModel) bool {
	for _, cert := range installedCertificates {
		if cert.Serial == certificate.Serial {
			return true
		}
	}
	return false
}

// CreateSelectableCodeSignGroups ...
func CreateSelectableCodeSignGroups(installedCertificates []certificateutil.CertificateInfoModel, profiles []profileutil.ProvisioningProfileInfoModel, bundleIDs []string) []SelectableCodeSignGroup {
	groups := []SelectableCodeSignGroup{}

	serialToProfiles := map[string][]profileutil.ProvisioningProfileInfoModel{}
	serialToCertificate := map[string]certificateutil.CertificateInfoModel{}
	for _, profile := range profiles {
		for _, certificate := range profile.DeveloperCertificates {
			if !containsCertificate(installedCertificates, certificate) {
				continue
			}

			certificateProfiles, ok := serialToProfiles[certificate.Serial]
			if !ok {
				certificateProfiles = []profileutil.ProvisioningProfileInfoModel{}
			}
			certificateProfiles = append(certificateProfiles, profile)
			serialToProfiles[certificate.Serial] = certificateProfiles
			serialToCertificate[certificate.Serial] = certificate
		}
	}

	for serial, profiles := range serialToProfiles {
		certificate := serialToCertificate[serial]

		bundleIDToProfiles := map[string][]profileutil.ProvisioningProfileInfoModel{}
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
				bundleIDToProfiles[bundleID] = matchingProfiles
			}
		}

		if len(bundleIDToProfiles) == len(bundleIDs) {
			group := SelectableCodeSignGroup{
				Certificate:         certificate,
				BundleIDProfilesMap: bundleIDToProfiles,
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
