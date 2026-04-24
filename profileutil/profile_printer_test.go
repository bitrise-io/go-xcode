package profileutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/stretchr/testify/require"
)

func TestProfilePrinter_PrintableProfile(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	futureExpiry := fixedTime.Add(24 * time.Hour)

	baseProfile := ProvisioningProfileInfoModel{
		UUID:           "test-uuid",
		Name:           "Test Profile",
		TeamName:       "Test Team",
		TeamID:         "TEAM123",
		BundleID:       "com.example.app",
		ExportType:     exportoptions.MethodDevelopment,
		ExpirationDate: futureExpiry,
	}

	printer := NewProfilePrinter(log.NewLogger(), fakeTimeProvider{fixedTime})

	t.Run("valid profile contains expected keys", func(t *testing.T) {
		output := printer.PrintableProfile(baseProfile)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.Contains(t, result, "name")
		require.Contains(t, result, "export_type")
		require.Contains(t, result, "team")
		require.Contains(t, result, "bundle_id")
		require.Contains(t, result, "expiry")
		require.Contains(t, result, "is_xcode_managed")
		require.Contains(t, result, "capabilities")
		require.Contains(t, result, "certificates")
		require.NotContains(t, result, "errors")
	})

	t.Run("known capability entitlement appears in capabilities", func(t *testing.T) {
		profile := baseProfile
		profile.Entitlements = plistutil.PlistData{"aps-environment": "production"}
		output := printer.PrintableProfile(profile)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		capabilities, ok := result["capabilities"].(map[string]interface{})
		require.True(t, ok)
		require.Contains(t, capabilities, "aps-environment")
	})

	t.Run("unknown entitlement key absent from capabilities", func(t *testing.T) {
		profile := baseProfile
		profile.Entitlements = plistutil.PlistData{"com.custom.entitlement": true}
		output := printer.PrintableProfile(profile)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		capabilities, ok := result["capabilities"].(map[string]interface{})
		require.True(t, ok)
		require.NotContains(t, capabilities, "com.custom.entitlement")
	})

	t.Run("provisioned devices present in output", func(t *testing.T) {
		profile := baseProfile
		profile.ProvisionedDevices = []string{"device-udid-1"}
		output := printer.PrintableProfile(profile)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.Contains(t, result, "devices")
	})

	t.Run("nil provisioned devices absent from output", func(t *testing.T) {
		output := printer.PrintableProfile(baseProfile)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.NotContains(t, result, "devices")
	})

	t.Run("certificates list reflects developer certificates", func(t *testing.T) {
		profile := baseProfile
		profile.DeveloperCertificates = []certificateutil.CertificateInfoModel{
			{CommonName: "Apple Development: Test User", Serial: "abc123", TeamID: "TEAM123"},
		}
		output := printer.PrintableProfile(profile)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		certs, ok := result["certificates"].([]interface{})
		require.True(t, ok)
		require.Len(t, certs, 1)
		cert, ok := certs[0].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Apple Development: Test User", cert["name"])
		require.Equal(t, "abc123", cert["serial"])
		require.Equal(t, "TEAM123", cert["team_id"])
	})

	t.Run("expired profile has errors", func(t *testing.T) {
		expired := baseProfile
		expired.ExpirationDate = fixedTime.Add(-time.Second)
		output := printer.PrintableProfile(expired)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.Contains(t, result, "errors")
	})

	t.Run("matching installed certificate has no cert error", func(t *testing.T) {
		profile := baseProfile
		profile.DeveloperCertificates = []certificateutil.CertificateInfoModel{{Serial: "abc"}}
		output := printer.PrintableProfile(profile, certificateutil.CertificateInfoModel{Serial: "abc"})
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.NotContains(t, result, "errors")
	})

	t.Run("no matching installed certificate has cert error", func(t *testing.T) {
		profile := baseProfile
		profile.DeveloperCertificates = []certificateutil.CertificateInfoModel{{Serial: "abc"}}
		output := printer.PrintableProfile(profile, certificateutil.CertificateInfoModel{Serial: "xyz"})
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.Contains(t, result, "errors")
	})
}

type fakeTimeProvider struct{ t time.Time }

func (f fakeTimeProvider) Now() time.Time { return f.t }
