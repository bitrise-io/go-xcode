package certificateutil

import (
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCertificatesFromPKCS12Content_roundTrip(t *testing.T) {
	const (
		commonName = "Apple Development: Jane Doe"
		teamID     = "TEAMID"
		teamName   = "Acme Inc"
		passphrase = "secret"
	)
	cert, privateKey, err := GenerateTestCertificate(1, teamID, teamName, commonName, time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)

	p12, err := NewCertificateInfo(*cert, privateKey).EncodeToP12(passphrase)
	require.NoError(t, err)

	infos, err := CertificatesFromPKCS12Content(p12, passphrase)
	require.NoError(t, err)
	require.Len(t, infos, 1)
	require.Equal(t, commonName, infos[0].CommonName)
	require.Equal(t, teamID, infos[0].TeamID)
	require.Equal(t, teamName, infos[0].TeamName)
}

func TestCertificatesFromPKCS12Content_wrongPassword(t *testing.T) {
	cert, privateKey, err := GenerateTestCertificate(1, "TEAM", "Team", "Name", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	p12, err := NewCertificateInfo(*cert, privateKey).EncodeToP12("correct")
	require.NoError(t, err)

	_, err = CertificatesFromPKCS12Content(p12, "wrong")
	require.Error(t, err)
}

func TestNewCertificateFromPemContent(t *testing.T) {
	cert, _, err := GenerateTestCertificate(1, "TEAM", "Team", "Name", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	pemContent := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

	parsed, err := NewCertificateFromPemContent(pemContent)
	require.NoError(t, err)
	require.Equal(t, cert.SerialNumber, parsed.SerialNumber)

	_, err = NewCertificateFromPemContent([]byte("not a pem"))
	require.Error(t, err)

	// A caller can pass anything to this exported function. A malformed private key reaches the
	// same branch, so the error must describe the content rather than quote it.
	malformedKey := []byte("-----BEGIN PRIVATE KEY-----\nMIIsecretkeymaterial\n-----")
	_, err = NewCertificateFromPemContent(malformedKey)
	require.Error(t, err)
	// Pins the branch: without this, NotContains would also pass if pem.Decode had succeeded and
	// x509.ParseCertificate produced the error instead.
	require.Contains(t, err.Error(), "no PEM block found")
	require.NotContains(t, err.Error(), "secretkeymaterial")
}

func TestNewCertificateFromDERContent(t *testing.T) {
	cert, _, err := GenerateTestCertificate(1, "TEAM", "Team", "Name", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)

	parsed, err := NewCertificateFromDERContent(cert.Raw)
	require.NoError(t, err)
	require.Equal(t, cert.SerialNumber, parsed.SerialNumber)

	_, err = NewCertificateFromDERContent([]byte("garbage"))
	require.Error(t, err)
}
