package exportoptionsgenerator

import "github.com/bitrise-io/go-xcode/v2/certificateutil"

// CodesignIdentityProvider can list certificate infos.
type CodesignIdentityProvider interface {
	ListCodesignIdentities() ([]certificateutil.CertificateInfo, error)
}

// LocalCodesignIdentityProvider ...
type LocalCodesignIdentityProvider struct{}

// ListCodesignIdentities ...
func (p LocalCodesignIdentityProvider) ListCodesignIdentities() ([]certificateutil.CertificateInfo, error) {
	certs, err := certificateutil.InstalledCodesigningCertificateInfos()
	if err != nil {
		return nil, err
	}
	certInfo := certificateutil.FilterValidCertificateInfos(certs)
	return append(certInfo.ValidCertificates, certInfo.DuplicatedCertificates...), nil
}
