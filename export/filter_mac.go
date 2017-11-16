package export

import "github.com/bitrise-tools/go-xcode/certificateutil"

// FilterCodeSignGroupsForInstallerCertificate ...
func FilterCodeSignGroupsForInstallerCertificate(codeSignGroups []CodeSignGroup, installedInstallerCertificates []certificateutil.CertificateInfoModel) []CodeSignGroupMac {
	filteredGroups := []CodeSignGroupMac{}

	for _, group := range codeSignGroups {
		macGroup := CodeSignGroupMac{
			CodeSignGroup: group,
		}

		for _, installerCertificate := range installedInstallerCertificates {
			macGroup.InstallerCertificate = installerCertificate

			if macGroup.InstallerCertificate.TeamID == macGroup.Certificate.TeamID {
				filteredGroups = append(filteredGroups, macGroup)
				break
			}
		}
	}
	return filteredGroups
}
