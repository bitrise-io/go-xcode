package certificateutil

import (
	v1certificate "github.com/bitrise-io/go-xcode/certificateutil"
)

func V2Certificate(cert CertificateInfoModel) v1certificate.CertificateInfoModel {
	return v1certificate.CertificateInfoModel(cert)
}

func V1Certificate(cert v1certificate.CertificateInfoModel) CertificateInfoModel {
	return CertificateInfoModel(cert)
}
