package export

import (
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// IosCodeSignGroup ...
type IosCodeSignGroup struct {
	certificate        certificateutil.CertificateInfoModel
	bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel
}

// NewIOSGroup ...
func NewIOSGroup(certificate certificateutil.CertificateInfoModel, bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel) *IosCodeSignGroup {
	return &IosCodeSignGroup{
		certificate:        certificate,
		bundleIDProfileMap: bundleIDProfileMap,
	}
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
