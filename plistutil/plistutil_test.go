package plistutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetBool(t *testing.T) {
	profileData, err := NewPlistDataFromContent(enterpriseProfileContent)
	require.NoError(t, err)

	allDevices, ok := profileData.GetBool("ProvisionsAllDevices")
	require.Equal(t, true, ok)
	require.Equal(t, true, allDevices)
}

func TestGetTime(t *testing.T) {
	profileData, err := NewPlistDataFromContent(developmentProfileContent)
	require.NoError(t, err)

	expire, ok := profileData.GetTime("ExpirationDate")
	require.Equal(t, true, ok)

	// 2017-09-22T11:28:46Z
	desiredExpire, err := time.Parse("2006-01-02T15:04:05Z", "2017-09-22T11:28:46Z")
	require.NoError(t, err)
	require.Equal(t, true, expire.Equal(desiredExpire))
}

func TestGetInt(t *testing.T) {
	profileData, err := NewPlistDataFromContent(developmentProfileContent)
	require.NoError(t, err)

	version, ok := profileData.GetInt("Version")
	require.Equal(t, true, ok)
	require.Equal(t, uint64(1), version)
}

func TestGetStringArray(t *testing.T) {
	profileData, err := NewPlistDataFromContent(developmentProfileContent)
	require.NoError(t, err)

	devices, ok := profileData.GetStringArray("ProvisionedDevices")
	require.Equal(t, true, ok)
	require.Equal(t, 1, len(devices))
	require.Equal(t, "b13813075ad9b298cb9a9f28555c49573d8bc322", devices[0])
}

func TestGetMapStringInterface(t *testing.T) {
	profileData, err := NewPlistDataFromContent(developmentProfileContent)
	require.NoError(t, err)

	entitlements, ok := profileData.GetMapStringInterface("Entitlements")
	require.Equal(t, true, ok)

	teamID, ok := entitlements.GetString("com.apple.developer.team-identifier")
	require.Equal(t, true, ok)
	require.Equal(t, "9NS44DLTN7", teamID)
}
