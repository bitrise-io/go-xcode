package codesigngroup

import (
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

type Ios struct {
	certificate        certificateutil.CertificateInfoModel
	bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel
}

func NewIOSGroup(certificate certificateutil.CertificateInfoModel, bundleIDProfileMap map[string]profileutil.ProvisioningProfileInfoModel) *Ios {
	return &Ios{
		certificate:        certificate,
		bundleIDProfileMap: bundleIDProfileMap,
	}
}

func (signGroup *Ios) Certificate() certificateutil.CertificateInfoModel {
	return signGroup.certificate
}

func (signGroup *Ios) InstallerCertificate() *certificateutil.CertificateInfoModel {
	return nil
}

func (signGroup *Ios) BundleIDProfileMap() map[string]profileutil.ProvisioningProfileInfoModel {
	return signGroup.bundleIDProfileMap
}
