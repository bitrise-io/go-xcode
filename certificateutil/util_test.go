package certificateutil

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/v2/timeutil"
	"github.com/stretchr/testify/require"
)

func TestFilterCertificateInfoModelsByFilterFunc(t *testing.T) {
	filterableCerts := []CertificateInfo{
		CertificateInfo{TeamID: "my-team-id"},
		CertificateInfo{TeamID: "find-this-team-id"},
		CertificateInfo{TeamID: "my--another-team-id"},
		CertificateInfo{TeamID: "test-team-id", CommonName: "test common name"},
		CertificateInfo{TeamID: "test-team-id2", CommonName: "find this common name"},
	}
	expectedCertsByTeamID := []CertificateInfo{
		CertificateInfo{TeamID: "find-this-team-id"},
	}

	foundCerts := FilterCertificateInfoModelsByFilterFunc(filterableCerts, func(cert CertificateInfo) bool { return cert.TeamID == "find-this-team-id" })
	require.Equal(t, expectedCertsByTeamID, foundCerts)

	expectedCertsByCommonNameExact := []CertificateInfo{
		CertificateInfo{TeamID: "test-team-id2", CommonName: "find this common name"},
	}

	foundCerts = FilterCertificateInfoModelsByFilterFunc(filterableCerts, func(cert CertificateInfo) bool { return cert.CommonName == "find this common name" })
	require.Equal(t, expectedCertsByCommonNameExact, foundCerts)

	expectedCertsByCommonNameMatch := []CertificateInfo{
		CertificateInfo{TeamID: "test-team-id", CommonName: "test common name"},
		CertificateInfo{TeamID: "test-team-id2", CommonName: "find this common name"},
	}

	foundCerts = FilterCertificateInfoModelsByFilterFunc(filterableCerts, func(cert CertificateInfo) bool { return strings.Contains(cert.CommonName, "common name") })
	require.Equal(t, expectedCertsByCommonNameMatch, foundCerts)
}

func TestFilterValidCertificateInfos(t *testing.T) {
	const serial = int64(1234)
	const teamID = "MYTEAMID"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	const commonName = "Apple Developer: test"
	notBefore := time.Now()
	validExpiry := notBefore.AddDate(1, 0, 0)
	earlierValidExpiry := notBefore.AddDate(0, 1, 0)
	invalidExpiry := notBefore.AddDate(-1, 0, 0)

	latestValidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, notBefore, validExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	latestValidCertInfo := NewCertificateInfo(*latestValidCert, privateKey)
	t.Logf("Test certificate generated: %s", latestValidCertInfo)

	earlierValidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, notBefore, earlierValidExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	earlierValidCertInfo := NewCertificateInfo(*earlierValidCert, privateKey)
	t.Logf("Test certificate generated: %s", earlierValidCertInfo)

	invalidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, notBefore, invalidExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	invalidCertInfo := NewCertificateInfo(*invalidCert, privateKey)
	t.Logf("Test certificate generated: %s", invalidCertInfo)

	tests := []struct {
		name             string
		certificateInfos []CertificateInfo
		want             ValidCertificateInfo
	}{
		{
			name:             "one valid cert",
			certificateInfos: []CertificateInfo{latestValidCertInfo},
			want: ValidCertificateInfo{
				ValidCertificates:      []CertificateInfo{latestValidCertInfo},
				InvalidCertificates:    nil,
				DuplicatedCertificates: nil,
			},
		},
		{
			name:             "one valid, one invalid cert with same name",
			certificateInfos: []CertificateInfo{latestValidCertInfo, invalidCertInfo},
			want: ValidCertificateInfo{
				ValidCertificates:      []CertificateInfo{latestValidCertInfo},
				InvalidCertificates:    []CertificateInfo{invalidCertInfo},
				DuplicatedCertificates: nil,
			},
		},
		{
			name:             "2 valid, duplicated certs",
			certificateInfos: []CertificateInfo{latestValidCertInfo, earlierValidCertInfo, invalidCertInfo},
			want: ValidCertificateInfo{
				ValidCertificates:      []CertificateInfo{latestValidCertInfo},
				InvalidCertificates:    []CertificateInfo{invalidCertInfo},
				DuplicatedCertificates: []CertificateInfo{earlierValidCertInfo},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterValidCertificateInfos(tt.certificateInfos, timeutil.NewDefaultTimeProvider()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterValidCertificateInfos() = %v, want %v", got, tt.want)
			}
		})
	}
}
