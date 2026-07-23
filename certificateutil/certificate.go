package certificateutil

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/bitrise-io/go-pkcs12"
	"github.com/bitrise-io/go-utils/fileutil"
)

// CertificatesFromPKCS12Content returns an array of CertificateInfoModel
// Used to parse p12 file containing multiple codesign identities (exported from macOS Keychain)
func CertificatesFromPKCS12Content(content []byte, password string) ([]CertificateInfoModel, error) {
	privateKeys, certificates, err := pkcs12.DecodeAll(content, password)
	if err != nil {
		return nil, err
	}

	if len(certificates) != len(privateKeys) {
		return nil, errors.New("pkcs12: different number of certificates and private keys found")
	}

	if len(certificates) == 0 {
		return nil, errors.New("pkcs12: no certificate and private key pair found")
	}

	infos := []CertificateInfoModel{}
	for i, certificate := range certificates {
		if certificate != nil {
			infos = append(infos, NewCertificateInfo(*certificate, privateKeys[i]))
		}
	}

	return infos, nil
}

// CertificatesFromPKCS12File ...
func CertificatesFromPKCS12File(pkcs12Pth, password string) ([]CertificateInfoModel, error) {
	content, err := fileutil.ReadBytesFromFile(pkcs12Pth)
	if err != nil {
		return nil, err
	}

	return CertificatesFromPKCS12Content(content, password)
}

// CertificateFromDERContent ...
func CertificateFromDERContent(content []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(content)
}

// CeritifcateFromPemContent ...
func CeritifcateFromPemContent(content []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(content)
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		return nil, fmt.Errorf("failed to parse profile from: %s", string(content))
	}
	return CertificateFromDERContent(block.Bytes)
}
