package certificate

import (
	"crypto/x509"
	"fmt"
	"io"
	"time"

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

// Details ...
type Details struct {
	CommonName string
	TeamName   string
	TeamID     string
	EndDate    time.Time
	StartDate  time.Time

	Serial          string
	SHA1Fingerprint string
}

type Certificate struct {
	Certificate *x509.Certificate
	PrivateKey  interface{}
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

		certs = append(certs, Certificate{
			Certificate: certificate,
			PrivateKey:  privateKey,
		})
	}

	return certs, nil
}
