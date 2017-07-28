package provisioningprofile

import (
	"testing"

	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
	"github.com/stretchr/testify/require"
)

func TestGetExportMethod(t *testing.T) {
	t.Log("development profile specifies development export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(developmentProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodDevelopment, GetExportMethod(profile))
	}

	t.Log("app store profile specifies app-store export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(appStoreProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodAppStore, GetExportMethod(profile))
	}

	t.Log("ad hoc profile specifies ad-hoc export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(adHocProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodAdHoc, GetExportMethod(profile))
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := plistutil.NewPlistDataFromContent(enterpriseProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodEnterprise, GetExportMethod(profile))
	}
}

func TestGetDeveloperTeam(t *testing.T) {
	t.Log("development profile specifies development export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(developmentProfileContent)
		require.NoError(t, err)
		require.Equal(t, "9NS44DLTN7", GetDeveloperTeam(profile))
	}

	t.Log("app store profile specifies app-store export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(appStoreProfileContent)
		require.NoError(t, err)
		require.Equal(t, "9NS44DLTN7", GetDeveloperTeam(profile))
	}

	t.Log("ad hoc profile specifies ad-hoc export method")
	{
		profile, err := plistutil.NewPlistDataFromContent(adHocProfileContent)
		require.NoError(t, err)
		require.Equal(t, "9NS44DLTN7", GetDeveloperTeam(profile))
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := plistutil.NewPlistDataFromContent(enterpriseProfileContent)
		require.NoError(t, err)
		require.Equal(t, "PF3BP78LQ8", GetDeveloperTeam(profile))
	}
}

func TestParseBuildSettingsOut(t *testing.T) {
	buildSettings, err := parseBuildSettingsOut(buildSettingsOut)
	require.NoError(t, err)
	require.Equal(t, 384, len(buildSettings))
	require.Equal(t, "Bitrise.ios-simple-objc", buildSettings["PRODUCT_BUNDLE_IDENTIFIER"])
	require.Equal(t, "ios-simple-objc", buildSettings["PRODUCT_NAME"])
}
