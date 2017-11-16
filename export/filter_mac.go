package export

import "github.com/bitrise-tools/go-xcode/certificateutil"

// FilterCodeSignGroupsForInstallerCertificate ...
func FilterCodeSignGroupsForInstallerCertificate(codeSignGroups []CodeSignGroup, installedInstallerCertificates []certificateutil.CertificateInfoModel) []CodeSignGroupMac {
	filteredGroups := []CodeSignGroupMac{}

	for _, group := range codeSignGroups {
		for _, installerCertificate := range installedInstallerCertificates {
			matchedMacGroups := []CodeSignGroupMac{}
			if installerCertificate.TeamID == group.Certificate.TeamID {
				matchedMacGroups = append(matchedMacGroups, CodeSignGroupMac{
					CodeSignGroup:        group,
					InstallerCertificate: installerCertificate,
				})
			}
			if len(matchedMacGroups) > 0 {
				filteredGroups = append(filteredGroups, matchedMacGroups[0])
			}
		}
	}
	return filteredGroups
}
