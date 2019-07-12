package certificateutil

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFilterCertificateInfoModelsByFilterFunc(t *testing.T) {
	filterableCerts := []CertificateInfoModel{
		CertificateInfoModel{TeamID: "my-team-id"},
		CertificateInfoModel{TeamID: "find-this-team-id"},
		CertificateInfoModel{TeamID: "my--another-team-id"},
		CertificateInfoModel{TeamID: "test-team-id", CommonName: "test common name"},
		CertificateInfoModel{TeamID: "test-team-id2", CommonName: "find this common name"},
	}
	expectedCertsByTeamID := []CertificateInfoModel{
		CertificateInfoModel{TeamID: "find-this-team-id"},
	}

	foundCerts := FilterCertificateInfoModelsByFilterFunc(filterableCerts, func(cert CertificateInfoModel) bool { return cert.TeamID == "find-this-team-id" })
	require.Equal(t, expectedCertsByTeamID, foundCerts)

	expectedCertsByCommonNameExact := []CertificateInfoModel{
		CertificateInfoModel{TeamID: "test-team-id2", CommonName: "find this common name"},
	}

	foundCerts = FilterCertificateInfoModelsByFilterFunc(filterableCerts, func(cert CertificateInfoModel) bool { return cert.CommonName == "find this common name" })
	require.Equal(t, expectedCertsByCommonNameExact, foundCerts)

	expectedCertsByCommonNameMatch := []CertificateInfoModel{
		CertificateInfoModel{TeamID: "test-team-id", CommonName: "test common name"},
		CertificateInfoModel{TeamID: "test-team-id2", CommonName: "find this common name"},
	}

	foundCerts = FilterCertificateInfoModelsByFilterFunc(filterableCerts, func(cert CertificateInfoModel) bool { return strings.Contains(cert.CommonName, "common name") })
	require.Equal(t, expectedCertsByCommonNameMatch, foundCerts)
}

func TestFilterValidCertificateInfos(t *testing.T) {
	const serial = int64(1234)
	const teamID = "MYTEAMID"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	const commonName = "Apple Developer: test"
	validExpiry := time.Now().AddDate(1, 0, 0)
	earlierValidExpiry := time.Now().AddDate(0, 1, 0)
	invalidExpiry := time.Now().AddDate(-1, 0, 0)

	latestValidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, validExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	latestValidCertInfo := NewCertificateInfo(*latestValidCert, privateKey)
	t.Logf("Test certificate generated: %s", latestValidCertInfo)

	earlierValidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, earlierValidExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	earlierValidCertInfo := NewCertificateInfo(*latestValidCert, earlierValidCert)
	t.Logf("Test certificate generated: %s", earlierValidCertInfo)

	invalidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, invalidExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	invalidCertInfo := NewCertificateInfo(*invalidCert, privateKey)
	t.Logf("Test certificate generated: %s", invalidCertInfo)

	tests := []struct {
		name             string
		certificateInfos []CertificateInfoModel
		want             ValidCertificateInfo
	}{
		{
			name:             "one valid cert",
			certificateInfos: []CertificateInfoModel{latestValidCertInfo},
			want: ValidCertificateInfo{
				ValidCertificates:      []CertificateInfoModel{latestValidCertInfo},
				InvalidCertificates:    nil,
				DuplicatedCertificates: nil,
			},
		},
		{
			name:             "one valid, one invalid cert with same name",
			certificateInfos: []CertificateInfoModel{latestValidCertInfo, invalidCertInfo},
			want: ValidCertificateInfo{
				ValidCertificates:      []CertificateInfoModel{latestValidCertInfo},
				InvalidCertificates:    []CertificateInfoModel{invalidCertInfo},
				DuplicatedCertificates: nil,
			},
		},
		{
			name:             "2 valid, duplicated certs",
			certificateInfos: []CertificateInfoModel{latestValidCertInfo, earlierValidCertInfo, invalidCertInfo},
			want: ValidCertificateInfo{
				ValidCertificates:      []CertificateInfoModel{latestValidCertInfo},
				InvalidCertificates:    []CertificateInfoModel{invalidCertInfo},
				DuplicatedCertificates: []CertificateInfoModel{earlierValidCertInfo},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterValidCertificateInfos(tt.certificateInfos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterValidCertificateInfos() = %v, want %v", got, tt.want)
			}
		})
	}
}
