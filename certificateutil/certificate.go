package certificateutil

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/bitrise-io/go-pkcs12"
)

// CertificatesFromPKCS12Content returns an array of CertificateInfoModel.
// Used to parse a p12 file containing multiple codesign identities (exported from macOS Keychain).
func CertificatesFromPKCS12Content(content []byte, password string) ([]CertificateInfoModel, error) {
	privateKeys, certificates, err := pkcs12.DecodeAll(content, password)
	if err != nil {
		return nil, err
	}

	if len(certificates) != len(privateKeys) {
		return nil, fmt.Errorf("pkcs12: found %d certificates but %d private keys", len(certificates), len(privateKeys))
	}

	if len(certificates) == 0 {
		return nil, fmt.Errorf("pkcs12: no certificate and private key pair found")
	}

	infos := []CertificateInfoModel{}
	for i, certificate := range certificates {
		if certificate != nil {
			infos = append(infos, NewCertificateInfo(*certificate, privateKeys[i]))
		}
	}

	return infos, nil
}

// NewCertificateFromDERContent parses a certificate from DER-encoded content.
func NewCertificateFromDERContent(content []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(content)
}

// NewCertificateFromPemContent parses a certificate from PEM-encoded content.
func NewCertificateFromPemContent(content []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(content)
	if block == nil || len(block.Bytes) == 0 {
		return nil, fmt.Errorf("failed to parse certificate from: %s", string(content))
	}
	return NewCertificateFromDERContent(block.Bytes)
}
