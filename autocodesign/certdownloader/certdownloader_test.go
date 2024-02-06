package certdownloader

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/stretchr/testify/assert"
)

func Test_downloader_GetCertificates_Local(t *testing.T) {
	certInfo := createTestCert(t)
	passphrase := ""

	certData, err := certInfo.EncodeToP12(passphrase)
	if err != nil {
		t.Errorf("init: failed to encode certificate: %s", err)
	}

	p12File, err := os.CreateTemp("", "*.p12")
	if err != nil {
		t.Errorf("init: failed to create temp test file: %s", err)
	}

	if _, err = p12File.Write(certData); err != nil {
		t.Errorf("init: failed to write test file: %s", err)
	}

	if err = p12File.Close(); err != nil {
		t.Errorf("init: failed to close file: %s", err)
	}

	p12path := "file://" + p12File.Name()

	d := downloader{
		certs: []CertificateAndPassphrase{{
			URL:        p12path,
			Passphrase: passphrase,
		}},
		client: http.DefaultClient,
	}
	got, err := d.GetCertificates()

	want := []certificateutil.CertificateInfoModel{
		certInfo,
	}

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_downloader_GetCertificates_Remote(t *testing.T) {
	certInfo := createTestCert(t)
	passphrase := ""

	certData, err := certInfo.EncodeToP12(passphrase)
	if err != nil {
		t.Errorf("init: failed to encode certificate: %s", err)
	}

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write(certData)
		if err != nil {
			t.Errorf("failed to write response: %s", err)
		}
	}))

	d := downloader{
		certs: []CertificateAndPassphrase{{
			URL:        storage.URL,
			Passphrase: passphrase,
		}},
		client: http.DefaultClient,
	}
	got, err := d.GetCertificates()

	want := []certificateutil.CertificateInfoModel{
		certInfo,
	}

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func createTestCert(t *testing.T) certificateutil.CertificateInfoModel {
	const (
		teamID     = "MYTEAMID"
		commonName = "Apple Developer: test"
		teamName   = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	)
	expiry := time.Now().AddDate(1, 0, 0)
	serial := int64(1234)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(serial, teamID, teamName, commonName, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate: %s", err)
	}

	certInfo := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. Serial: %s Team ID: %s Common name: %s", certInfo.Serial, certInfo.TeamID, certInfo.CommonName)

	return certInfo
}
