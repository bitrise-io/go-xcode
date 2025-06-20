package export

import (
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// CodeSignGroup ...
type CodeSignGroup interface {
	Certificate() certificateutil.CertificateInfoModel
	InstallerCertificate() *certificateutil.CertificateInfoModel
	BundleIDProfileMap() map[string]profileutil.ProvisioningProfileInfoModel
}
