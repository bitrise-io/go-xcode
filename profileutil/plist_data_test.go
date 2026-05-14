package profileutil

import (
	"testing"

	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/stretchr/testify/require"
)

func TestPlistData_IOSProfile(t *testing.T) {
	t.Log("development profile specifies development export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(iosDevelopmentProfileContent)
		require.NoError(t, err)
		profileType, err := PlistData(profile).GetProfileType()
		require.NoError(t, err)
		require.Equal(t, ProfileTypeIos, profileType)
		require.Equal(t, "4b617a5f-e31e-4edc-9460-718a5abacd05", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test Development", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.*", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "*", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodDevelopment, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "Some Dude", PlistData(profile).GetTeamName())
		require.Equal(t, "2016-09-22T11:28:46Z", PlistData(profile).GetCreationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, "2017-09-22T11:28:46Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string{"b13813075ad9b298cb9a9f28555c49573d8bc322"}, PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
		require.Empty(t, PlistData(profile).GetDeveloperCertificateInfo())
		require.Equal(t, false, PlistData(profile).GetProvisionsAllDevices())
		require.NotEmpty(t, PlistData(profile).GetEntitlements())
	}

	t.Log("app store profile specifies app-store export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(iosAppStoreProfileContent)
		require.NoError(t, err)
		profileType, err := PlistData(profile).GetProfileType()
		require.NoError(t, err)
		require.Equal(t, ProfileTypeIos, profileType)
		require.Equal(t, "a60668dd-191a-4770-8b1e-b453b87aa60b", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test App Store", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.*", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "*", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodAppStore, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "Some Dude", PlistData(profile).GetTeamName())
		require.Equal(t, "2016-09-22T11:29:12Z", PlistData(profile).GetCreationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, "2017-09-21T13:20:06Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string(nil), PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
		require.Empty(t, PlistData(profile).GetDeveloperCertificateInfo())
		require.Equal(t, false, PlistData(profile).GetProvisionsAllDevices())
		require.NotEmpty(t, PlistData(profile).GetEntitlements())
	}

	t.Log("ad hoc profile specifies ad-hoc export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(iosAdHocProfileContent)
		require.NoError(t, err)
		profileType, err := PlistData(profile).GetProfileType()
		require.NoError(t, err)
		require.Equal(t, ProfileTypeIos, profileType)
		require.Equal(t, "26668300-5743-46a1-8e00-7023e2e35c7d", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test Ad Hoc", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.*", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "*", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodAdHoc, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "Some Dude", PlistData(profile).GetTeamName())
		require.Equal(t, "2016-09-22T11:29:38Z", PlistData(profile).GetCreationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, "2017-09-21T13:20:06Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string{"b13813075ad9b298cb9a9f28555c49573d8bc322"}, PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
		require.Empty(t, PlistData(profile).GetDeveloperCertificateInfo())
		require.Equal(t, false, PlistData(profile).GetProvisionsAllDevices())
		require.NotEmpty(t, PlistData(profile).GetEntitlements())
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := plistutil.NewPlistDataFromContent(iosEnterpriseProfileContent)
		require.NoError(t, err)
		profileType, err := PlistData(profile).GetProfileType()
		require.NoError(t, err)
		require.Equal(t, ProfileTypeIos, profileType)
		require.Equal(t, "8d6caa15-ac49-48f9-9bd3-ce9244add6a0", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test Enterprise", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.com.Bitrise.Test", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "com.Bitrise.Test", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodEnterprise, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "Bitrise", PlistData(profile).GetTeamName())
		require.Equal(t, "2015-10-05T13:32:46Z", PlistData(profile).GetCreationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, "2016-10-04T13:32:46Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string(nil), PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
		require.Empty(t, PlistData(profile).GetDeveloperCertificateInfo())
		require.Equal(t, true, PlistData(profile).GetProvisionsAllDevices())
		require.NotEmpty(t, PlistData(profile).GetEntitlements())
	}
}

func TestPlistData_TVOSProfile(t *testing.T) {
	t.Log("it creates model from tvOS appstore profile content")
	{
		profile, err := plistutil.NewPlistDataFromContent(tvosAppStoreProfileContent)
		require.NoError(t, err)
		profileType, err := PlistData(profile).GetProfileType()
		require.NoError(t, err)
		require.Equal(t, ProfileTypeTvOs, profileType)
		require.Equal(t, "dec523d5-624b-44bd-8d16-6d1d69c63276", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise app-store - (bdh.NPO-Live.bitrise.sample)", PlistData(profile).GetName())
		require.Equal(t, "72SA8V3WYL.bdh.NPO-Live.bitrise.sample", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "bdh.NPO-Live.bitrise.sample", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodAppStore, PlistData(profile).GetExportMethod())
		require.Equal(t, "72SA8V3WYL", PlistData(profile).GetTeamID())
		require.Equal(t, "Bitrise", PlistData(profile).GetTeamName())
		require.Equal(t, "2018-10-24T11:22:30Z", PlistData(profile).GetCreationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, "2019-04-16T08:42:18Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string(nil), PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
		require.Empty(t, PlistData(profile).GetDeveloperCertificateInfo())
		require.Equal(t, false, PlistData(profile).GetProvisionsAllDevices())
		require.NotEmpty(t, PlistData(profile).GetEntitlements())
	}
}

func TestPlistData_GetProfileTypeErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "missing Platform key",
			content: noPlatformProfileContent,
		},
		{
			name:    "unknown platform",
			content: unknownPlatformProfileContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := plistutil.NewPlistDataFromContent(tt.content)
			require.NoError(t, err)
			_, err = PlistData(profile).GetProfileType()
			require.Error(t, err)
		})
	}
}

func TestPlistData_GetExportMethod(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    exportoptions.Method
	}{
		{
			name:    "no devices, no ProvisionsAllDevices: app-store",
			content: macosAppStoreProfileContent,
			want:    exportoptions.MethodAppStore,
		},
		{
			name:    "no devices, ProvisionsAllDevices=true: developer-id",
			content: macosDeveloperIDProfileContent,
			want:    exportoptions.MethodDeveloperID,
		},
		{
			name:    "has ProvisionedDevices: development",
			content: macosDevelopmentProfileContent,
			want:    exportoptions.MethodDevelopment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := plistutil.NewPlistDataFromContent(tt.content)
			require.NoError(t, err)
			require.Equal(t, tt.want, PlistData(profile).GetExportMethod())
		})
	}
}
