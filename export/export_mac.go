package export

import "github.com/bitrise-tools/go-xcode/certificateutil"

// CodeSignGroupMac ...
type CodeSignGroupMac struct {
	InstallerCertificate certificateutil.CertificateInfoModel
	CodeSignGroup
}
