package certificateutil

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/go-pkcs12"
	"github.com/bitrise-io/go-xcode/v2/timeutil"
)

type CertificateInfo struct {
	CommonName string
	TeamName   string
	TeamID     string
	EndDate    time.Time
	StartDate  time.Time

	Serial          string
	SHA1Fingerprint string

	Certificate x509.Certificate
	PrivateKey  interface{}
}

func NewCertificateInfo(certificate x509.Certificate, privateKey interface{}) CertificateInfo {
	fingerprint := sha1.Sum(certificate.Raw)
	fingerprintStr := fmt.Sprintf("%x", fingerprint)

	return CertificateInfo{
		CommonName:      certificate.Subject.CommonName,
		TeamName:        strings.Join(certificate.Subject.Organization, " "),
		TeamID:          strings.Join(certificate.Subject.OrganizationalUnit, " "),
		EndDate:         certificate.NotAfter,
		StartDate:       certificate.NotBefore,
		Serial:          certificate.SerialNumber.String(),
		SHA1Fingerprint: fingerprintStr,

		Certificate: certificate,
		PrivateKey:  privateKey,
	}
}

// CertificatesFromPKCS12Content returns an array of CertificateInfo
// Used to parse p12 file containing multiple codesign identities (exported from macOS Keychain)
func CertificatesFromPKCS12Content(content []byte, password string) ([]CertificateInfo, error) {
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

	infos := []CertificateInfo{}
	for i, certificate := range certificates {
		if certificate != nil {
			infos = append(infos, NewCertificateInfo(*certificate, privateKeys[i]))
		}
	}

	return infos, nil
}

func (info CertificateInfo) String() string {
	team := fmt.Sprintf("%s (%s)", info.TeamName, info.TeamID)
	certInfo := fmt.Sprintf("Serial: %s, Name: %s, Team: %s, Expiry: %s", info.Serial, info.CommonName, team, info.EndDate)

	//if timeProvider != nil {
	//	if err := info.CheckValidity(*timeProvider); err != nil {
	//		certInfo = certInfo + fmt.Sprintf(", error: %s", err)
	//	}
	//}

	return certInfo
}

func (info CertificateInfo) CheckValidity(timeProvider timeutil.TimeProvider) error {
	return CheckValidity(info.Certificate, timeProvider)
}

// EncodeToP12 encodes a CertificateInfo in pkcs12 (.p12) format.
func (info CertificateInfo) EncodeToP12(passphrase string) ([]byte, error) {
	return pkcs12.Encode(rand.Reader, info.PrivateKey, &info.Certificate, nil, passphrase)
}
