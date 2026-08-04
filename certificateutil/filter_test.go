package certificateutil

import (
	"reflect"
	"testing"
	"time"
)

func TestGroupCertificatesByValidity(t *testing.T) {
	const serial = int64(1234)
	const teamID = "MYTEAMID"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	const commonName = "Apple Developer: test"
	validExpiry := time.Now().AddDate(1, 0, 0)
	earlierValidExpiry := time.Now().AddDate(0, 1, 0)
	invalidExpiry := time.Now().AddDate(-1, 0, 0)
	printer := printerAt(time.Now())

	latestValidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, validExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	latestValidCertInfo := NewCertificateInfo(*latestValidCert, privateKey)
	t.Logf("Test certificate generated: %s", printer.CertificateSummary(latestValidCertInfo))

	earlierValidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, earlierValidExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	earlierValidCertInfo := NewCertificateInfo(*earlierValidCert, privateKey)
	t.Logf("Test certificate generated: %s", printer.CertificateSummary(earlierValidCertInfo))

	invalidCert, privateKey, err := GenerateTestCertificate(serial, teamID, teamName, commonName, invalidExpiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	invalidCertInfo := NewCertificateInfo(*invalidCert, privateKey)
	t.Logf("Test certificate generated: %s", printer.CertificateSummary(invalidCertInfo))

	tests := []struct {
		name             string
		certificateInfos []CertificateInfoModel
		want             CertificateValidityGroups
	}{
		{
			name:             "one valid cert",
			certificateInfos: []CertificateInfoModel{latestValidCertInfo},
			want: CertificateValidityGroups{
				ValidCertificates:      []CertificateInfoModel{latestValidCertInfo},
				InvalidCertificates:    nil,
				DuplicatedCertificates: nil,
			},
		},
		{
			name:             "one valid, one invalid cert with same name",
			certificateInfos: []CertificateInfoModel{latestValidCertInfo, invalidCertInfo},
			want: CertificateValidityGroups{
				ValidCertificates:      []CertificateInfoModel{latestValidCertInfo},
				InvalidCertificates:    []CertificateInfoModel{invalidCertInfo},
				DuplicatedCertificates: nil,
			},
		},
		{
			name:             "2 valid, duplicated certs",
			certificateInfos: []CertificateInfoModel{latestValidCertInfo, earlierValidCertInfo, invalidCertInfo},
			want: CertificateValidityGroups{
				ValidCertificates:      []CertificateInfoModel{latestValidCertInfo},
				InvalidCertificates:    []CertificateInfoModel{invalidCertInfo},
				DuplicatedCertificates: []CertificateInfoModel{earlierValidCertInfo},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GroupCertificatesByValidity(tt.certificateInfos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupCertificatesByValidity() = %v, want %v", got, tt.want)
			}
		})
	}
}
