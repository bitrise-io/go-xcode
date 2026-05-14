package profileutil

import (
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/fullsailor/pkcs7"
	"github.com/stretchr/testify/require"
)

func TestNewProvisioningProfileInfo(t *testing.T) {
	profilePkcs7 := pkcs7.PKCS7{Content: []byte(iosDevelopmentProfileContent)}
	got, err := NewProvisioningProfileInfo(profilePkcs7)

	require.NoError(t, err)
	require.Equal(t, ProfileTypeIos, got.Type)
	require.Equal(t, "4b617a5f-e31e-4edc-9460-718a5abacd05", got.UUID)
	require.Equal(t, "Bitrise Test Development", got.Name)
	require.Equal(t, "Some Dude", got.TeamName)
	require.Equal(t, "9NS44DLTN7", got.TeamID)
	require.Equal(t, "*", got.BundleID)
	require.Equal(t, exportoptions.MethodDevelopment, got.ExportType)
	require.Equal(t, []string{"b13813075ad9b298cb9a9f28555c49573d8bc322"}, got.ProvisionedDevices)
	require.False(t, got.ProvisionsAllDevices)
	require.Empty(t, got.DeveloperCertificates)
	require.Equal(t, time.Date(2016, 9, 22, 11, 28, 46, 0, time.UTC), got.CreationDate)
	require.Equal(t, time.Date(2017, 9, 22, 11, 28, 46, 0, time.UTC), got.ExpirationDate)
	require.NotEmpty(t, got.Entitlements)
}

func TestNewProvisioningProfileInfoFromPKCS7Content(t *testing.T) {
	sd, err := pkcs7.NewSignedData([]byte(macosFullDevelopmentProfileContent))
	require.NoError(t, err)
	pkcs7Bytes, err := sd.Finish()
	require.NoError(t, err)

	got, err := NewProvisioningProfileInfoFromPKCS7Content(pkcs7Bytes)

	require.NoError(t, err)
	require.Equal(t, ProfileTypeMacOs, got.Type)
	require.Equal(t, "dea6a48c-d7d3-4624-9f6b-e0c3b3ce517d", got.UUID)
	require.Equal(t, "_profile_bug_type_catalyst", got.Name)
	require.Equal(t, "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG", got.TeamName)
	require.Equal(t, "72SA8V3WYL", got.TeamID)
	require.Equal(t, "io.bitrise.mobile.ios.QuickActionsTodayExtension", got.BundleID)
	require.Equal(t, exportoptions.MethodDevelopment, got.ExportType)
	require.Equal(t, []string{"BA0EC799-F254-5574-B335-E70B8A2FA5E7"}, got.ProvisionedDevices)
	require.False(t, got.ProvisionsAllDevices)
	require.Empty(t, got.DeveloperCertificates)
	require.Equal(t, time.Date(2022, 2, 28, 10, 35, 39, 0, time.UTC), got.CreationDate)
	require.Equal(t, time.Date(2023, 2, 28, 10, 35, 39, 0, time.UTC), got.ExpirationDate)
	require.NotEmpty(t, got.Entitlements)
}

// TestIsXcodeManaged covers full cases; this is a sanity check that the method delegates correctly.
func TestProvisioningProfileInfoModel_IsXcodeManaged(t *testing.T) {
	info := ProvisioningProfileInfoModel{Name: "XC iOS: com.example.app"}
	require.True(t, info.IsXcodeManaged())
}

func TestProvisioningProfileInfoModel_CheckValidity(t *testing.T) {
	expiration := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	info := ProvisioningProfileInfoModel{ExpirationDate: expiration}

	tests := []struct {
		name        string
		currentTime time.Time
		wantErr     bool
	}{
		{
			name:        "valid: before expiration",
			currentTime: expiration.Add(-time.Second),
		},
		{
			name:        "expired: exactly at expiration",
			currentTime: expiration,
			wantErr:     true,
		},
		{
			name:        "expired: after expiration",
			currentTime: expiration.Add(time.Second),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := info.CheckValidity(func() time.Time { return tt.currentTime })
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProvisioningProfileInfoModel_HasInstalledCertificate(t *testing.T) {
	tests := []struct {
		name                  string
		developerCertificates []certificateutil.CertificateInfoModel
		installedCertificates []certificateutil.CertificateInfoModel
		want                  bool
	}{
		{
			name:                  "no developer certificates",
			installedCertificates: []certificateutil.CertificateInfoModel{{Serial: "abc"}},
		},
		{
			name:                  "matching serial",
			developerCertificates: []certificateutil.CertificateInfoModel{{Serial: "abc"}},
			installedCertificates: []certificateutil.CertificateInfoModel{{Serial: "abc"}},
			want:                  true,
		},
		{
			name:                  "no matching serial",
			developerCertificates: []certificateutil.CertificateInfoModel{{Serial: "abc"}},
			installedCertificates: []certificateutil.CertificateInfoModel{{Serial: "xyz"}},
		},
		{
			name:                  "multiple certificates, one matches",
			developerCertificates: []certificateutil.CertificateInfoModel{{Serial: "abc"}, {Serial: "def"}},
			installedCertificates: []certificateutil.CertificateInfoModel{{Serial: "def"}},
			want:                  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ProvisioningProfileInfoModel{DeveloperCertificates: tt.developerCertificates}
			require.Equal(t, tt.want, info.HasInstalledCertificate(tt.installedCertificates))
		})
	}
}
