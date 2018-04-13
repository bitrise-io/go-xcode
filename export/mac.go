package export

import (
	"github.com/bitrise-tools/go-xcode/certificateutil"
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/profileutil"
)

// MacCodeSignGroup ...
type MacCodeSignGroup struct {
	certificate          certificateutil.CertificateInfoModel
	installerCertificate *certificateutil.CertificateInfoModel
	bundleIDProfileMap   map[string]profileutil.ProvisioningProfileInfoModel
}

// Certificate ...
func (signGroup *MacCodeSignGroup) Certificate() certificateutil.CertificateInfoModel {
	return signGroup.certificate
}

// InstallerCertificate ...
func (signGroup *MacCodeSignGroup) InstallerCertificate() *certificateutil.CertificateInfoModel {
	return signGroup.installerCertificate
}

// BundleIDProfileMap ...
func (signGroup *MacCodeSignGroup) BundleIDProfileMap() map[string]profileutil.ProvisioningProfileInfoModel {
	return signGroup.bundleIDProfileMap
}

// CreateMacCodeSignGroup ...
func CreateMacCodeSignGroup(selectableGroups []SelectableCodeSignGroup, installedInstallerCertificates []certificateutil.CertificateInfoModel, exportMethod exportoptions.Method) []MacCodeSignGroup {
	macosCodeSignGroups := []MacCodeSignGroup{}

	iosCodesignGroups := CreateIosCodeSignGroups(selectableGroups)

	for _, group := range iosCodesignGroups {
		if exportMethod == exportoptions.MethodAppStore {
			installerCertificates := []certificateutil.CertificateInfoModel{}

			for _, installerCertificate := range installedInstallerCertificates {
				if installerCertificate.TeamID == group.Certificate().TeamID {
					installerCertificates = append(installerCertificates, installerCertificate)
				}
			}

			if len(installerCertificates) > 0 {
				installerCert := installerCertificates[0]
				macosCodeSignGroups = append(macosCodeSignGroups, MacCodeSignGroup{
					certificate:          group.Certificate(),
					installerCertificate: &installerCert,
					bundleIDProfileMap:   group.BundleIDProfileMap(),
				})
			}
		} else {
			macosCodeSignGroups = append(macosCodeSignGroups, MacCodeSignGroup{
				certificate:        group.Certificate(),
				bundleIDProfileMap: group.BundleIDProfileMap(),
			})
		}
	}

	return macosCodeSignGroups
}
