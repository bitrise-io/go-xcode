package certificate

import (
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"io"
	"strings"

	"github.com/bitrise-io/go-pkcs12"
)

/*
	Details.String()
	Certificate.CheckValidity()
	Certificate.EncodeToP12(passphrase string)
	NewCertificateInfo(certificate x509.Certificate, privateKey interface{})
	InstalledCodesigningCertificateInfos()
	InstalledInstallerCertificateInfos()
	CertificatesFromPKCS12Content(content []byte, password string)
	CertificatesFromPKCS12File(pkcs12Pth, password string)
	CertificateFromDERContent(content []byte)
	CeritifcateFromPemContent(content []byte)
	InstalledCodesigningCertificateNames()
	InstalledMacAppStoreCertificateNames()
	InstalledCodesigningCertificates()
	InstalledMacAppStoreCertificates()
	FilterCertificateInfoModelsByFilterFunc(certificates []CertificateInfoModel, filterFunc func(certificate CertificateInfoModel) bool)
	FilterValidCertificateInfos(certificateInfos []CertificateInfoModel)
*/

type Certificate struct {
	X509Certificate *x509.Certificate
	PrivateKey      interface{}
}

func NewCertificatesFromFile(reader io.Reader, password string) ([]Certificate, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	privateKeys, certificates, err := pkcs12.DecodeAll(data, password)
	if err != nil {
		return nil, err
	}

	if len(certificates) == 0 {
		return nil, fmt.Errorf("pkcs12: no certificate and private key pair found")
	}

	if len(certificates) != len(privateKeys) {
		return nil, fmt.Errorf("pkcs12: different number of certificates and private keys found")
	}

	var certs []Certificate
	for i, certificate := range certificates {
		privateKey := privateKeys[i]

		cert := newCertificate(certificate, privateKey)
		certs = append(certs, cert)
	}

	return certs, nil
}

func newCertificate(cert *x509.Certificate, key interface{}) Certificate {
	return Certificate{
		X509Certificate: cert,
		PrivateKey:      key,
	}
}

func (cert Certificate) Details() Details {
	return Details{
		CommonName:      cert.X509Certificate.Subject.CommonName,
		TeamName:        strings.Join(cert.X509Certificate.Subject.Organization, " "),
		TeamID:          strings.Join(cert.X509Certificate.Subject.OrganizationalUnit, " "),
		EndDate:         cert.X509Certificate.NotAfter,
		StartDate:       cert.X509Certificate.NotBefore,
		Serial:          cert.X509Certificate.SerialNumber.String(),
		SHA1Fingerprint: fmt.Sprintf("%x", sha1.Sum(cert.X509Certificate.Raw)),
	}
}
