package export

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/certificateutil"
	"github.com/bitrise-tools/go-xcode/profileutil"
	"github.com/ryanuber/go-glob"
)

// SelectableCodeSignGroup ..
type SelectableCodeSignGroup struct {
	Certificate         certificateutil.CertificateInfoModel
	BundleIDProfilesMap map[string][]profileutil.ProvisioningProfileInfoModel
}

func printableSelectableCodeSignGroup(group SelectableCodeSignGroup) string {
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

func isCertificateInstalled(installedCertificates []certificateutil.CertificateInfoModel, certificate certificateutil.CertificateInfoModel) bool {
	installedMap := map[string]bool{}
	for _, certificate := range installedCertificates {
		installedMap[certificate.Serial] = true
	}
	return installedMap[certificate.Serial]
}

// CreateSelectableCodeSignGroups ...
func CreateSelectableCodeSignGroups(certificates []certificateutil.CertificateInfoModel, profiles []profileutil.ProvisioningProfileInfoModel, bundleIDs []string) []SelectableCodeSignGroup {
	log.Debugf("\n")
	log.Debugf("Creating Codesign Groups...")

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

	if len(groups) == 0 {
		log.Debugf("Can not create any Codesign Groups for the bundle ids using the installed code sign files")
	}

	for _, group := range groups {
		log.Debugf(printableSelectableCodeSignGroup(group))
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
