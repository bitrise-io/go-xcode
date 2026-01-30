package certdownloader

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
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

	d := downloader{
		certs: []CertificateAndPassphrase{{
			URL:        "file:///fake/path/cert.p12",
			Passphrase: passphrase,
		}},
		logger:       log.NewLogger(),
		fileProvider: fakeFileProvider{content: certData},
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

	d := downloader{
		certs: []CertificateAndPassphrase{{
			URL:        "https://example.com/cert.p12",
			Passphrase: passphrase,
		}},
		logger:       log.NewLogger(),
		fileProvider: fakeFileProvider{content: certData},
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

type fakeFileProvider struct {
	content []byte
}

func (f fakeFileProvider) Contents(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.content)), nil
}

func (f fakeFileProvider) LocalPath(_ context.Context, _ string) (string, error) {
	return "", nil
}
