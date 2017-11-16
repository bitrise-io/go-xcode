package certificateutil

import (
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"strings"
	"time"
)

// CertificateInfoModel ...
type CertificateInfoModel struct {
	CommonName string
	TeamName   string
	TeamID     string
	EndDate    time.Time
	StartDate  time.Time

	Serial          string
	SHA1Fingerprint string

	certificate x509.Certificate
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

// CheckValidity ...
func (info CertificateInfoModel) CheckValidity() error {
	return CheckValidity(info.certificate)
}

// NewCertificateInfo ...
func NewCertificateInfo(certificate x509.Certificate) CertificateInfoModel {
	fingerprint := sha1.Sum(certificate.Raw)
	fingerprintStr := fmt.Sprintf("%x", fingerprint)

	return CertificateInfoModel{
		CommonName:      certificate.Subject.CommonName,
		TeamName:        strings.Join(certificate.Subject.Organization, " "),
		TeamID:          strings.Join(certificate.Subject.OrganizationalUnit, " "),
		EndDate:         certificate.NotAfter,
		StartDate:       certificate.NotBefore,
		Serial:          certificate.SerialNumber.String(),
		SHA1Fingerprint: fingerprintStr,
		certificate:     certificate,
	}
}

// CertificateInfos ...
func CertificateInfos(certificates []*x509.Certificate) []CertificateInfoModel {
	infos := []CertificateInfoModel{}
	for _, certificate := range certificates {
		if certificate != nil {
			info := NewCertificateInfo(*certificate)
			infos = append(infos, info)
		}
	}

	return infos
}

// NewCertificateInfosFromPKCS12 ...
func NewCertificateInfosFromPKCS12(pkcs12Pth, password string) ([]CertificateInfoModel, error) {
	certificates, err := CertificatesFromPKCS12File(pkcs12Pth, password)
	if err != nil {
		return nil, err
	}
	return CertificateInfos(certificates), nil
}

// InstalledCodesigningCertificateInfos ...
func InstalledCodesigningCertificateInfos() ([]CertificateInfoModel, error) {
	certificates, err := InstalledCodesigningCertificates()
	if err != nil {
		return nil, err
	}
	return CertificateInfos(certificates), nil
}

// InstalledMacAppStoreCertificateInfos ...
func InstalledMacAppStoreCertificateInfos() ([]CertificateInfoModel, error) {
	certificates, err := InstalledMacAppStoreCertificates()
	if err != nil {
		return nil, err
	}
	return CertificateInfos(certificates), nil
}
