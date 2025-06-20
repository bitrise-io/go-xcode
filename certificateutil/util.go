package certificateutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sort"
	"time"
)

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

// CheckValidity ...
func CheckValidity(certificate x509.Certificate) error {
	timeNow := time.Now()
	if !timeNow.After(certificate.NotBefore) {
		return fmt.Errorf("Certificate is not yet valid - validity starts at: %s", certificate.NotBefore)
	}
	if !timeNow.Before(certificate.NotAfter) {
		return fmt.Errorf("Certificate is not valid anymore - validity ended at: %s", certificate.NotAfter)
	}
	return nil
}

// FilterCertificateInfoModelsByFilterFunc ...
func FilterCertificateInfoModelsByFilterFunc(certificates []CertificateInfoModel, filterFunc func(certificate CertificateInfoModel) bool) []CertificateInfoModel {
	filteredCertificates := []CertificateInfoModel{}

	for _, certificate := range certificates {
		if filterFunc(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		}
	}

	return filteredCertificates
}

// ValidCertificateInfo contains the certificate infos filtered as valid, invalid and duplicated common name certificates
type ValidCertificateInfo struct {
	ValidCertificates,
	InvalidCertificates,
	DuplicatedCertificates []CertificateInfoModel
}

// FilterValidCertificateInfos filters out invalid and duplicated common name certificaates
func FilterValidCertificateInfos(certificateInfos []CertificateInfoModel) ValidCertificateInfo {
	var invalidCertificates []CertificateInfoModel
	nameToCerts := map[string][]CertificateInfoModel{}
	for _, certificateInfo := range certificateInfos {
		if certificateInfo.CheckValidity() != nil {
			invalidCertificates = append(invalidCertificates, certificateInfo)
			continue
		}

		nameToCerts[certificateInfo.CommonName] = append(nameToCerts[certificateInfo.CommonName], certificateInfo)
	}

	var validCertificates, duplicatedCertificates []CertificateInfoModel
	for _, certs := range nameToCerts {
		if len(certs) == 0 {
			continue
		}

		sort.Slice(certs, func(i, j int) bool {
			return certs[i].EndDate.After(certs[j].EndDate)
		})
		validCertificates = append(validCertificates, certs[0])
		if len(certs) > 1 {
			duplicatedCertificates = append(duplicatedCertificates, certs[1:]...)
		}
	}

	return ValidCertificateInfo{
		ValidCertificates:      validCertificates,
		InvalidCertificates:    invalidCertificates,
		DuplicatedCertificates: duplicatedCertificates,
	}
}

// GenerateTestCertificate creates a certificate (signed by a self-signed CA cert) for test purposes
func GenerateTestCertificate(serial int64, teamID, teamName, commonName string, expiry time.Time) (*x509.Certificate, *rsa.PrivateKey, error) {
	CAtemplate := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"Pear Worldwide Developer Relations"},
			CommonName:   "Pear Worldwide Developer Relations CA",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	CAprivatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Self-signed certificate, parent is the template
	CAcertData, err := x509.CreateCertificate(rand.Reader, CAtemplate, CAtemplate, &CAprivatekey.PublicKey, CAprivatekey)
	if err != nil {
		return nil, nil, err
	}
	CAcert, err := x509.ParseCertificate(CAcertData)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SerialNumber:          big.NewInt(serial),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{teamName},
			OrganizationalUnit: []string{teamID},
			CommonName:         commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  expiry,
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		KeyUsage: x509.KeyUsageDigitalSignature,
	}

	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	certData, err := x509.CreateCertificate(rand.Reader, template, CAcert, &privatekey.PublicKey, CAprivatekey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, nil, err
	}

	return cert, privatekey, nil
}
