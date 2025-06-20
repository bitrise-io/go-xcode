package export

import (
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// MacCodeSignGroup ...
type MacCodeSignGroup struct {
	certificate          certificateutil.CertificateInfo
	installerCertificate *certificateutil.CertificateInfo
	bundleIDProfileMap   map[string]profileutil.ProvisioningProfileInfoModel
}

// Certificate ...
func (signGroup *MacCodeSignGroup) Certificate() certificateutil.CertificateInfo {
	return signGroup.certificate
}

// InstallerCertificate ...
func (signGroup *MacCodeSignGroup) InstallerCertificate() *certificateutil.CertificateInfo {
	return signGroup.installerCertificate
}

// BundleIDProfileMap ...
func (signGroup *MacCodeSignGroup) BundleIDProfileMap() map[string]profileutil.ProvisioningProfileInfoModel {
	return signGroup.bundleIDProfileMap
}

// NewMacGroup ...
func NewMacGroup(certificate certificateutil.CertificateInfo, installerCertificate *certificateutil.CertificateInfo, bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel) *MacCodeSignGroup {
	return &MacCodeSignGroup{
		certificate:          certificate,
		installerCertificate: installerCertificate,
		bundleIDProfileMap:   bundleIDProfileMap,
	}
}

// CreateMacCodeSignGroup ...
func CreateMacCodeSignGroup(selectableGroups []SelectableCodeSignGroup, installedInstallerCertificates []certificateutil.CertificateInfo, exportMethod exportoptions.Method) []MacCodeSignGroup {
	macosCodeSignGroups := []MacCodeSignGroup{}

	iosCodesignGroups := CreateIosCodeSignGroups(selectableGroups)

	for _, group := range iosCodesignGroups {
		if exportMethod.IsAppStore() {
			installerCertificates := []certificateutil.CertificateInfo{}

			for _, installerCertificate := range installedInstallerCertificates {
				if installerCertificate.TeamID == group.certificate.TeamID {
					installerCertificates = append(installerCertificates, installerCertificate)
				}
			}

			if len(installerCertificates) > 0 {
				installerCertificate := installerCertificates[0]
				macosCodeSignGroups = append(macosCodeSignGroups, MacCodeSignGroup{
					certificate:          group.certificate,
					installerCertificate: &installerCertificate,
					bundleIDProfileMap:   group.bundleIDProfileMap,
				})
			}
		} else {
			macosCodeSignGroups = append(macosCodeSignGroups, MacCodeSignGroup{
				certificate:        group.certificate,
				bundleIDProfileMap: group.bundleIDProfileMap,
			})
		}
	}

	return macosCodeSignGroups
}
