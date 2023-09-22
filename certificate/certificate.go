package certificate

import (
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"io"
	"strings"
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

type Type string

const (
	AppleDevelopment  Type = "Apple Development"
	AppleDistribution Type = "Apple Distribution"

	iPhoneDeveloper    Type = "iPhone Developer"
	iPhoneDistribution Type = "iPhone Distribution"

	MacDeveloper                      Type = "Mac Developer"
	ThirdPartyMacDeveloperApplication Type = "3rd Party Mac Developer Application"
	ThirdPartyMacDeveloperInstaller   Type = "3rd Party Mac Developer Installer"
	DeveloperIDApplication            Type = "Developer ID Application"
	DeveloperIDInstaller              Type = "Developer ID Installer"
)

var knownSoftwareCertificateTypes = map[Type]bool{
	AppleDevelopment:                  true,
	AppleDistribution:                 true,
	iPhoneDeveloper:                   true,
	iPhoneDistribution:                true,
	MacDeveloper:                      true,
	ThirdPartyMacDeveloperApplication: true,
	ThirdPartyMacDeveloperInstaller:   true,
	DeveloperIDApplication:            true,
	DeveloperIDInstaller:              true,
}

type Platform string

const (
	IOS   Platform = "iOS"
	MacOS Platform = "macOS"
	All   Platform = "All"
)

// Details ...
type Details struct {
	CommonName      string
	TeamName        string
	TeamID          string
	EndDate         time.Time
	StartDate       time.Time
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

		cert := newCertificate(certificate, privateKey)
		certs = append(certs, cert)
	}

	return certs, nil
}

func newCertificate(cert *x509.Certificate, key interface{}) Certificate {
	return Certificate{
		Certificate: cert,
		PrivateKey:  key,
	}
}

func (cert Certificate) Details() Details {
	return Details{
		CommonName:      cert.Certificate.Subject.CommonName,
		TeamName:        strings.Join(cert.Certificate.Subject.Organization, " "),
		TeamID:          strings.Join(cert.Certificate.Subject.OrganizationalUnit, " "),
		EndDate:         cert.Certificate.NotAfter,
		StartDate:       cert.Certificate.NotBefore,
		Serial:          cert.Certificate.SerialNumber.String(),
		SHA1Fingerprint: fmt.Sprintf("%x", sha1.Sum(cert.Certificate.Raw)),
	}
}

func (d Details) Type() Type {
	split := strings.Split(d.CommonName, ":")
	if len(split) < 2 {
		// TODO: this shouldn't happen
		return ""
	}

	typeFromName := split[0]
	ok := knownSoftwareCertificateTypes[Type(typeFromName)]
	if !ok {
		// TODO: this should mean a Certificate for services (like Pass Type ID Certificate)
		return Type("")
	}

	return Type(typeFromName)
}

func (d Details) Platform() Platform {
	switch d.Type() {
	case AppleDevelopment, AppleDistribution:
		return All
	case iPhoneDeveloper, iPhoneDistribution:
		return IOS
	case MacDeveloper, ThirdPartyMacDeveloperApplication, ThirdPartyMacDeveloperInstaller, DeveloperIDApplication, DeveloperIDInstaller:
		return MacOS
	}

	// TODO: this should mean a Certificate for services (like Pass Type ID Certificate)
	return ""
}
