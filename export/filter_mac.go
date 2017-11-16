package export

import "github.com/bitrise-tools/go-xcode/certificateutil"

// FilterCodeSignGroupsForInstallerCertificate ...
func FilterCodeSignGroupsForInstallerCertificate(codeSignGroups []CodeSignGroup, installedInstallerCertificates []certificateutil.CertificateInfoModel) []CodeSignGroupMac {
	filteredGroups := []CodeSignGroupMac{}

	for _, group := range codeSignGroups {
		matchingGroup := true

		macGroup := CodeSignGroupMac{
			CodeSignGroup: group,
		}

		for _, installerCertificate := range installedInstallerCertificates {
			macGroup.InstallerCertificate = installerCertificate

			if macGroup.InstallerCertificate.TeamID != macGroup.Certificate.TeamID {
				matchingGroup = false
				break
			}
		}
		if matchingGroup {
			filteredGroups = append(filteredGroups, macGroup)
		}
	}
	return filteredGroups
}
