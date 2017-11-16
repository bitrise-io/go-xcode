package export

import "github.com/bitrise-tools/go-xcode/certificateutil"

// FilterCodeSignGroupsForInstallerCertificate ...
func FilterCodeSignGroupsForInstallerCertificate(codeSignGroups []CodeSignGroup, installedInstallerCertificates []certificateutil.CertificateInfoModel) []CodeSignGroupMac {
	filteredGroups := []CodeSignGroupMac{}

	for _, group := range codeSignGroups {
		certs := []certificateutil.CertificateInfoModel{}
		for _, installerCertificate := range installedInstallerCertificates {
			if installerCertificate.TeamID == group.Certificate.TeamID {
				certs = append(certs, installerCertificate)
			}
		}

		if len(certs) > 0 {
			filteredGroups = append(filteredGroups, CodeSignGroupMac{
				CodeSignGroup:        group,
				InstallerCertificate: certs[0],
			})
		}
	}
	return filteredGroups
}
