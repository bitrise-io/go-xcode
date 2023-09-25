package certificate

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewCertificatesFromFile(t *testing.T) {
	tests := []struct {
		name         string
		pth          string
		wantType     Type
		wantPlatform Platform
	}{
		{
			name:         "Apple Development",
			pth:          "Apple_Development.json",
			wantType:     AppleDevelopment,
			wantPlatform: All,
		},
		{
			name:         "Apple Distribution",
			pth:          "Apple_Distribution.json",
			wantType:     AppleDistribution,
			wantPlatform: All,
		},
		{
			name:         "iOS App Development",
			pth:          "iOS_App_Development.json",
			wantType:     iPhoneDeveloper,
			wantPlatform: IOS,
		},
		{
			name:         "iOS Distribution",
			pth:          "iOS_Distribution.json",
			wantType:     iPhoneDistribution,
			wantPlatform: IOS,
		},
		{
			name:         "Mac Development",
			pth:          "Mac_Development.json",
			wantType:     MacDeveloper,
			wantPlatform: MacOS,
		},
		{
			name:         "Mac App Distribution",
			pth:          "Mac_App_Distribution.json",
			wantType:     ThirdPartyMacDeveloperApplication,
			wantPlatform: MacOS,
		},
		{
			name:         "Mac Installer Distribution",
			pth:          "Mac_Installer_Distribution.json",
			wantType:     ThirdPartyMacDeveloperInstaller,
			wantPlatform: MacOS,
		},
		{
			name:         "Developer ID Application",
			pth:          "Developer_ID_Application.json",
			wantType:     DeveloperIDApplication,
			wantPlatform: MacOS,
		},
		{
			name:         "Developer ID Installer",
			pth:          "Developer_ID_Installer.json",
			wantType:     DeveloperIDInstaller,
			wantPlatform: MacOS,
		},
		{
			name:         "Apple Push Notification service SSL Sandbox",
			pth:          "Apple_Push_Notification_service_SSL_Sandbox.json",
			wantType:     "",
			wantPlatform: "",
		},
		{
			name:         "Pass Type ID Certificate",
			pth:          "Pass_Type_ID_Certificate.json",
			wantType:     "",
			wantPlatform: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata/json", tt.pth))
			require.NoError(t, err)

			x509Cert := newCertFromJSON(t, f)
			cert := NewCertificate(x509Cert, nil)
			details := cert.Details()

			require.Equal(t, tt.wantType, details.Type())

			fmt.Println(details)
		})
	}
}

type TestIssuer struct {
	CommonName                       string
	Organization, OrganizationalUnit []string
}

type TestCertificate struct {
	Subject             TestIssuer
	NotBefore, NotAfter time.Time
	SerialNumber        *big.Int
	Raw                 []byte
}

func newCertFromJSON(t *testing.T, reader io.Reader) *x509.Certificate {
	b, err := io.ReadAll(reader)
	require.NoError(t, err)

	var testCertificate TestCertificate
	err = json.Unmarshal(b, &testCertificate)
	require.NoError(t, err)

	newCert := x509.Certificate{}
	newCert.Subject.CommonName = testCertificate.Subject.CommonName
	newCert.Subject.Organization = testCertificate.Subject.Organization
	newCert.Subject.OrganizationalUnit = testCertificate.Subject.OrganizationalUnit
	newCert.NotAfter = testCertificate.NotAfter
	newCert.NotBefore = testCertificate.NotBefore
	newCert.SerialNumber = testCertificate.SerialNumber
	newCert.Raw = testCertificate.Raw

	return &newCert
}

//func Test_convertCertToJSON(t *testing.T) {
//	certs := []string{
//		"testdata/Apple_Development.p12",
//		"testdata/Apple_Distribution.p12",
//		"...",
//	}
//
//	for certIdx, certPth := range certs {
//		rootDir := filepath.Dir(certPth)
//		f, err := os.Open(certPth)
//		require.NoError(t, err)
//
//		c, err := NewCertificatesFromFile(f, "")
//		require.NoError(t, err)
//		require.Equal(t, 1, len(c))
//
//		cert := c[0]
//		x509Cert := redactSensitiveInfo(t, certIdx, cert.Certificate)
//
//		b, err := json.MarshalIndent(x509Cert, "", "\t")
//		require.NoError(t, err)
//
//		fileName := strings.TrimSuffix(filepath.Base(certPth), filepath.Ext(certPth)) + ".json"
//		jsonPth := filepath.Join(rootDir, "json", fileName)
//
//		err = os.WriteFile(jsonPth, b, os.ModePerm)
//		require.NoError(t, err)
//	}
//}
//
//func redactSensitiveInfo(t *testing.T, id int, cert *x509.Certificate) *x509.Certificate {
//	split := strings.Split(cert.Subject.CommonName, ":")
//	require.True(t, len(split) > 1)
//	certType := split[0]
//
//	newCert := x509.Certificate{}
//	newCert.Subject.CommonName = certType + ": John Doe"
//	newCert.Subject.Organization = []string{"Dev Team"}
//	newCert.Subject.OrganizationalUnit = []string{"team_id_1"}
//	newCert.NotAfter = cert.NotAfter
//	newCert.NotBefore = cert.NotBefore
//	newCert.SerialNumber = big.NewInt(int64(id))
//	newCert.Raw = []byte("raw_data")
//
//	return &newCert
//}
