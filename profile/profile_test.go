package profile

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fullsailor/pkcs7"

	"github.com/bitrise-io/go-plist"
	"github.com/stretchr/testify/require"
)

func TestNewProfile(t *testing.T) {
	tests := []struct {
		name         string
		pth          string
		wantType     Type
		wantPlatform []string
	}{
		{
			name:         "iOS App Development",
			pth:          "iOS_App_Development.plist",
			wantType:     Development,
			wantPlatform: []string{string(IOS), string(XROS), string(VisionOS)},
		},
		{
			name:         "iOS App Development with certificates for Xcode 11 and later",
			pth:          "iOS_App_Development_with_new_cert.plist",
			wantType:     Development,
			wantPlatform: []string{string(IOS), string(XROS), string(VisionOS)},
		},
		{
			name:         "iOS App Development with Mac devices",
			pth:          "iOS_App_Development_with_Mac.plist",
			wantType:     Development,
			wantPlatform: []string{string(IOS), string(XROS), string(VisionOS)},
		},
		{
			name:         "tvOS App Development",
			pth:          "tvOS_App_Development.plist",
			wantType:     Development,
			wantPlatform: []string{string(TVOS)},
		},
		{
			name:         "macOS App Development type Mac",
			pth:          "macOS_App_Development_type_Mac.plist",
			wantType:     Development,
			wantPlatform: []string{string(OSX)},
		},
		{
			name:         "macOS App Development type Mac Catalyst",
			pth:          "macOS_App_Development_type_Mac_Catalyst.plist",
			wantType:     Development,
			wantPlatform: []string{string(OSX), string(XROS), string(VisionOS)},
		},
		{
			name:         "Ad Hoc",
			pth:          "Ad_Hoc.plist",
			wantType:     AdHoc,
			wantPlatform: []string{string(IOS), string(XROS), string(VisionOS)},
		},
		{
			name:         "tvOS Ad Hoc",
			pth:          "tvOS_Ad_Hoc.plist",
			wantType:     AdHoc,
			wantPlatform: []string{string(TVOS)},
		},
		{
			name:         "App Store",
			pth:          "App_Store.plist",
			wantType:     AppStore,
			wantPlatform: []string{string(IOS), string(XROS), string(VisionOS)},
		},
		{
			name:         "tvOS App Store",
			pth:          "tvOS_App_Store.plist",
			wantType:     AppStore,
			wantPlatform: []string{string(TVOS)},
		},
		{
			name:         "Mac App Store type Mac",
			pth:          "Mac_App_Store_type_Mac.plist",
			wantType:     AppStore,
			wantPlatform: []string{string(OSX)},
		},
		{
			name:         "Mac App Store type Mac Catalyst",
			pth:          "Mac_App_Store_type_Mac_Catalyst.plist",
			wantType:     AppStore,
			wantPlatform: []string{string(OSX), string(XROS), string(VisionOS)},
		},
		{
			name:         "Developer ID Application",
			pth:          "Developer_ID_Application.plist",
			wantType:     DeveloperID,
			wantPlatform: []string{string(OSX)},
		},
		{
			name:         "DriverKit App Development",
			pth:          "DriverKit_App_Development.plist",
			wantType:     Development,
			wantPlatform: []string{string(OSX), string(IOS)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", "plist", tt.pth))
			require.NoError(t, err)

			b, err := io.ReadAll(f)
			require.NoError(t, err)

			pkcs7Profile := pkcs7.PKCS7{}
			pkcs7Profile.Content = b

			profile := newProfile(&pkcs7Profile)
			require.NotNil(t, profile)

			details, err := profile.Details()
			require.NoError(t, err)
			verifyDetails(t, details, b)
			require.Equal(t, tt.wantType, details.Type())
			require.Equal(t, tt.wantPlatform, details.Platform)
		})
	}
}

// verifyProfileModel verifies if the profile models is complete (contains all information from the profile plist data).
func verifyDetails(t *testing.T, details *Details, expectedProfilePlistData []byte) {
	b, err := plist.MarshalIndent(details, plist.XMLFormat, "\t")
	require.NoError(t, err)

	var profileData map[string]interface{}
	_, err = plist.Unmarshal(b, &profileData)
	require.NoError(t, err)

	var expectedData map[string]interface{}
	_, err = plist.Unmarshal(expectedProfilePlistData, &expectedData)
	require.NoError(t, err)

	require.Equal(t, expectedData, profileData)
}

//// Test_convertProfileToPlist unwraps the provisioning profile's plist content,
//// redacts sensitive information and writes to file.
//// testdata/Ad_Hoc.mobileprovision -> testdata/plist/Ad_Hoc.plist
//func Test_convertProfileToPlist(t *testing.T) {
//	profiles := []string{
//		"testdata/Ad_Hoc.mobileprovision",
//		"testdata/App_Store.mobileprovision",
//		"...",
//	}
//
//	for profileIdx, profilePth := range profiles {
//		rootDir := filepath.Dir(profilePth)
//		f, err := os.Open(profilePth)
//		require.NoError(t, err)
//
//		profile, err := NewProfileFromFile(f)
//		require.NoError(t, err)
//
//		profile = redactSensitiveInfo(profileIdx, profile)
//
//		b, err := plist.MarshalIndent(profile, plist.XMLFormat, "\t")
//		require.NoError(t, err)
//
//		fileName := strings.TrimSuffix(filepath.Base(profilePth), filepath.Ext(profilePth)) + ".plist"
//		plistPth := filepath.Join(rootDir, "plist", fileName)
//
//		err = os.WriteFile(plistPth, b, os.ModePerm)
//		require.NoError(t, err)
//	}
//}
//
//func redactSensitiveInfo(id int, profile *Profile) *Profile {
//	entitlements := profile.Entitlements["get-task-allow"]
//	profile.Entitlements = map[string]interface{}{"get-task-allow": entitlements}
//
//	var devices []string
//	for i := range profile.ProvisionedDevices {
//		devices = append(devices, fmt.Sprintf("device_%d", i))
//	}
//	profile.ProvisionedDevices = devices
//	profile.DeveloperCertificates = nil
//	profile.TeamName = "Dev Team"
//	profile.UUID = fmt.Sprintf("uuid_%d", id)
//	profile.AppIDName = fmt.Sprintf("app_id_%d", id)
//
//	var appIDPrefixes []string
//	for range profile.ApplicationIdentifierPrefix {
//		appIDPrefixes = append(appIDPrefixes, profile.AppIDName)
//	}
//	profile.ApplicationIdentifierPrefix = appIDPrefixes
//
//	var teamIDs []string
//	for i := range profile.TeamIdentifier {
//		teamIDs = append(teamIDs, fmt.Sprintf("team_id_%d", i))
//	}
//	profile.TeamIdentifier = teamIDs
//	profile.DEREncodedProfile = nil
//
//	return profile
//}
