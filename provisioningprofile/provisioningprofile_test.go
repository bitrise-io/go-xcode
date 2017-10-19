package provisioningprofile

import (
	"testing"

	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"github.com/stretchr/testify/require"
)

func TestPlistData(t *testing.T) {
	t.Log("development profile specifies development export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(developmentProfileContent)
		require.NoError(t, err)
		require.Equal(t, "4b617a5f-e31e-4edc-9460-718a5abacd05", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test Development", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.*", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "*", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodDevelopment, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "2017-09-22T11:28:46Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string{"b13813075ad9b298cb9a9f28555c49573d8bc322"}, PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
	}

	t.Log("app store profile specifies app-store export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(appStoreProfileContent)
		require.NoError(t, err)
		require.Equal(t, "a60668dd-191a-4770-8b1e-b453b87aa60b", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test App Store", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.*", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "*", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodAppStore, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "2017-09-21T13:20:06Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string(nil), PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
	}

	t.Log("ad hoc profile specifies ad-hoc export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(adHocProfileContent)
		require.NoError(t, err)
		require.Equal(t, "26668300-5743-46a1-8e00-7023e2e35c7d", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test Ad Hoc", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.*", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "*", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodAdHoc, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "2017-09-21T13:20:06Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string{"b13813075ad9b298cb9a9f28555c49573d8bc322"}, PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := plistutil.NewPlistDataFromContent(enterpriseProfileContent)
		require.NoError(t, err)
		require.Equal(t, "8d6caa15-ac49-48f9-9bd3-ce9244add6a0", PlistData(profile).GetUUID())
		require.Equal(t, "Bitrise Test Enterprise", PlistData(profile).GetName())
		require.Equal(t, "9NS44DLTN7.com.Bitrise.Test", PlistData(profile).GetApplicationIdentifier())
		require.Equal(t, "com.Bitrise.Test", PlistData(profile).GetBundleIdentifier())
		require.Equal(t, exportoptions.MethodEnterprise, PlistData(profile).GetExportMethod())
		require.Equal(t, "9NS44DLTN7", PlistData(profile).GetTeamID())
		require.Equal(t, "2016-10-04T13:32:46Z", PlistData(profile).GetExpirationDate().Format("2006-01-02T15:04:05Z"))
		require.Equal(t, []string(nil), PlistData(profile).GetProvisionedDevices())
		require.Equal(t, [][]uint8{[]uint8{}}, PlistData(profile).GetDeveloperCertificates())
	}
}
