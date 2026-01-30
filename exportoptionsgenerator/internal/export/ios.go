package export

import (
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/profileutil"
)

// IosCodeSignGroup ...
type IosCodeSignGroup struct {
	certificate        certificateutil.CertificateInfoModel
	bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel
}

// Certificate ...
func (signGroup *IosCodeSignGroup) Certificate() certificateutil.CertificateInfoModel {
	return signGroup.certificate
}

// InstallerCertificate ...
func (signGroup *IosCodeSignGroup) InstallerCertificate() *certificateutil.CertificateInfoModel {
	return nil
}

// BundleIDProfileMap ...
func (signGroup *IosCodeSignGroup) BundleIDProfileMap() map[string]profileutil.ProvisioningProfileInfoModel {
	return signGroup.bundleIDProfileMap
}

// NewIOSGroup ...
func NewIOSGroup(certificate certificateutil.CertificateInfoModel, bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel) *IosCodeSignGroup {
	return &IosCodeSignGroup{
		certificate:        certificate,
		bundleIDProfileMap: bundleIDProfileMap,
	}
}
