package certificate

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bitrise-io/go-pkcs12"
)

type Certificate struct {
	X509Certificate *x509.Certificate
	PrivateKey      interface{}
}

func NewCertificate(cert *x509.Certificate, key interface{}) Certificate {
	return Certificate{
		X509Certificate: cert,
		PrivateKey:      key,
	}
}

func NewCertificatesFromPKCS12File(reader io.Reader, password string) ([]Certificate, error) {
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

		cert := NewCertificate(certificate, privateKey)
		certs = append(certs, cert)
	}

	return certs, nil
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

func (cert Certificate) CheckValidity() error {
	timeNow := time.Now()
	if !timeNow.After(cert.X509Certificate.NotBefore) {
		return fmt.Errorf("validity starts at: %s", cert.X509Certificate.NotBefore)
	}
	if !timeNow.Before(cert.X509Certificate.NotAfter) {
		return fmt.Errorf("validity ended at: %s", cert.X509Certificate.NotAfter)
	}
	return nil
}

func (cert Certificate) EncodeToP12(passphrase string) ([]byte, error) {
	return pkcs12.Encode(rand.Reader, cert.PrivateKey, cert.X509Certificate, nil, passphrase)
}
