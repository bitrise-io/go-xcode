package export

import (
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// MacCodeSignGroup ...
type MacCodeSignGroup struct {
	certificate          certificateutil.CertificateInfoModel
	installerCertificate *certificateutil.CertificateInfoModel
	bundleIDProfileMap   map[string]profileutil.ProvisioningProfileInfoModel
}

// NewMacGroup ...
func NewMacGroup(certificate certificateutil.CertificateInfoModel, installerCertificate *certificateutil.CertificateInfoModel, bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel) *MacCodeSignGroup {
	return &MacCodeSignGroup{
		certificate:          certificate,
		installerCertificate: installerCertificate,
		bundleIDProfileMap:   bundleIDProfileMap,
	}
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
