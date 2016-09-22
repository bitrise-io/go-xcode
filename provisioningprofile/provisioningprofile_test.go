package provisioningprofile

import (
	"testing"

	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/stretchr/testify/require"
)

func TestNewFromProfileContent(t *testing.T) {
	t.Log("it creates model from development profile content")
	{
		profile, err := newFromProfileContent(developmentProfileContent)
		require.NoError(t, err)
		require.NotNil(t, profile.Name)
		require.Equal(t, "Bitrise Test Development", *profile.Name)
		require.NotNil(t, profile.ProvisionedDevices)
		require.Equal(t, 1, len(*profile.ProvisionedDevices))
		require.Equal(t, "b13813075ad9b298cb9a9f28555c49573d8bc322", (*profile.ProvisionedDevices)[0])
		require.Nil(t, profile.ProvisionsAllDevices)

		require.NotNil(t, profile.Entitlements)

		require.NotNil(t, (*profile.Entitlements).GetTaskAllow)
		require.Equal(t, true, *(*profile.Entitlements).GetTaskAllow)

		require.NotNil(t, (*profile.Entitlements).DeveloperTeamID)
		require.NotNil(t, "9NS44DLTN7", (*profile.Entitlements).DeveloperTeamID)
	}

	t.Log("it creates model from app store profile content")
	{
		profile, err := newFromProfileContent(appStoreProfileContent)
		require.NoError(t, err)
		require.NotNil(t, profile.Name)
		require.Equal(t, "Bitrise Test App Store", *profile.Name)
		require.Nil(t, profile.ProvisionedDevices)
		require.Nil(t, profile.ProvisionsAllDevices)

		require.NotNil(t, profile.Entitlements)

		require.NotNil(t, (*profile.Entitlements).GetTaskAllow)
		require.Equal(t, false, *(*profile.Entitlements).GetTaskAllow)

		require.NotNil(t, (*profile.Entitlements).DeveloperTeamID)
		require.NotNil(t, "9NS44DLTN7", (*profile.Entitlements).DeveloperTeamID)
	}

	t.Log("it creates model from ad hoc profile content")
	{
		profile, err := newFromProfileContent(adHocProfileContent)
		require.NoError(t, err)
		require.NotNil(t, profile.Name)
		require.Equal(t, "Bitrise Test Ad Hoc", *profile.Name)
		require.NotNil(t, profile.ProvisionedDevices)
		require.Equal(t, 1, len(*profile.ProvisionedDevices))
		require.Equal(t, "b13813075ad9b298cb9a9f28555c49573d8bc322", (*profile.ProvisionedDevices)[0])
		require.Nil(t, profile.ProvisionsAllDevices)

		require.NotNil(t, profile.Entitlements)

		require.NotNil(t, (*profile.Entitlements).GetTaskAllow)
		require.Equal(t, false, *(*profile.Entitlements).GetTaskAllow)

		require.NotNil(t, (*profile.Entitlements).DeveloperTeamID)
		require.NotNil(t, "9NS44DLTN7", (*profile.Entitlements).DeveloperTeamID)
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := newFromProfileContent(enterpriseProfileContent)
		require.NoError(t, err)
		require.NotNil(t, profile.Name)
		require.Equal(t, "PaintSpeciPadDistProf", *profile.Name)
		require.Nil(t, profile.ProvisionedDevices)
		require.NotNil(t, profile.ProvisionsAllDevices)
		require.Equal(t, true, *profile.ProvisionsAllDevices)

		require.NotNil(t, profile.Entitlements)

		require.NotNil(t, (*profile.Entitlements).GetTaskAllow)
		require.Equal(t, false, *(*profile.Entitlements).GetTaskAllow)

		require.NotNil(t, (*profile.Entitlements).DeveloperTeamID)
		require.NotNil(t, "PF3BP78LQ8", (*profile.Entitlements).DeveloperTeamID)
	}
}

func TestGetExportMethod(t *testing.T) {
	t.Log("development profile specifies development export method")
	{
		profile, err := newFromProfileContent(developmentProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodDevelopment, profile.GetExportMethod())
	}

	t.Log("app store profile specifies app-store export method")
	{
		profile, err := newFromProfileContent(appStoreProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodAppStore, profile.GetExportMethod())
	}

	t.Log("ad hoc profile specifies ad-hoc export method")
	{
		profile, err := newFromProfileContent(adHocProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodAdHoc, profile.GetExportMethod())
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := newFromProfileContent(enterpriseProfileContent)
		require.NoError(t, err)
		require.Equal(t, exportoptions.MethodEnterprise, profile.GetExportMethod())
	}
}

func TestGetDeveloperTeam(t *testing.T) {
	t.Log("development profile specifies development export method")
	{
		profile, err := newFromProfileContent(developmentProfileContent)
		require.NoError(t, err)
		require.Equal(t, "9NS44DLTN7", profile.GetDeveloperTeam())
	}

	t.Log("app store profile specifies app-store export method")
	{
		profile, err := newFromProfileContent(appStoreProfileContent)
		require.NoError(t, err)
		require.Equal(t, "9NS44DLTN7", profile.GetDeveloperTeam())
	}

	t.Log("ad hoc profile specifies ad-hoc export method")
	{
		profile, err := newFromProfileContent(adHocProfileContent)
		require.NoError(t, err)
		require.Equal(t, "9NS44DLTN7", profile.GetDeveloperTeam())
	}

	t.Log("it creates model from enterprise profile content")
	{
		profile, err := newFromProfileContent(enterpriseProfileContent)
		require.NoError(t, err)
		require.Equal(t, "PF3BP78LQ8", profile.GetDeveloperTeam())
	}
}
