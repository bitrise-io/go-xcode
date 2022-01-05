package autocodesign

import (
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/stretchr/testify/require"
)

func Test_createWildcardBundleID(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     string
		wantErr  bool
	}{
		{
			name:     "Invalid bundle id: empty",
			bundleID: "",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid bundle id: does not contain *",
			bundleID: "my_app",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "2 component bundle id",
			bundleID: "com.my_app",
			want:     "com.*",
			wantErr:  false,
		},
		{
			name:     "multi component bundle id",
			bundleID: "com.bitrise.my_app.uitest",
			want:     "com.bitrise.my_app.*",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateWildcardBundleID(tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("createWildcardBundleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createWildcardBundleID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_profileName(t *testing.T) {
	tests := []struct {
		profileType appstoreconnect.ProfileType
		bundleID    string
		want        string
	}{
		{
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS development - (io.bitrise.app)",
		},
		{
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS app-store - (io.bitrise.app)",
		},
		{
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS ad-hoc - (io.bitrise.app)",
		},
		{
			profileType: appstoreconnect.IOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS enterprise - (io.bitrise.app)",
		},

		{
			profileType: appstoreconnect.TvOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS development - (io.bitrise.app)",
		},
		{
			profileType: appstoreconnect.TvOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS app-store - (io.bitrise.app)",
		},
		{
			profileType: appstoreconnect.TvOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS ad-hoc - (io.bitrise.app)",
		},
		{
			profileType: appstoreconnect.TvOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS enterprise - (io.bitrise.app)",
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.profileType), func(t *testing.T) {
			got := profileName(tt.profileType, tt.bundleID)
			if got != tt.want {
				t.Errorf("profileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findMissingContainers(t *testing.T) {
	tests := []struct {
		name        string
		appEnts     Entitlements
		profileEnts Entitlements
		want        []string
		wantErr     bool
	}{
		{
			name: "equal without container",
			appEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),
			profileEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "equal with container",
			appEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "profile has more containers than project",
			appEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),
			profileEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "project has more containers than profile",
			appEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),

			want:    []string{"container1"},
			wantErr: false,
		},
		{
			name: "project has containers but profile doesn't",
			appEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: Entitlements(map[string]interface{}{
				"otherentitlement": "",
			}),

			want:    []string{"container1"},
			wantErr: false,
		},
		{
			name: "error check",
			appEnts: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": "break",
			}),

			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindMissingContainers(tt.appEnts, tt.profileEnts)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, got, tt.want)
		})
	}
}

func Test_IsProfileExpired(t *testing.T) {
	tests := []struct {
		prof                Profile
		minProfileDaysValid int
		name                string
		want                bool
	}{
		{
			name:                "no days set - profile expiry date after current time",
			minProfileDaysValid: 0,
			prof:                newMockProfile(profileArgs{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(5 * time.Hour))}}),
			want:                false,
		},
		{
			name:                "no days set - profile expiry date before current time",
			minProfileDaysValid: 0,
			prof:                newMockProfile(profileArgs{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(-5 * time.Hour))}}),
			want:                true,
		},
		{
			name:                "days set - profile expiry date after current time + days set",
			minProfileDaysValid: 2,
			prof:                newMockProfile(profileArgs{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(5 * 24 * time.Hour))}}),
			want:                false,
		},
		{
			name:                "days set - profile expiry date before current time + days set",
			minProfileDaysValid: 2,
			prof:                newMockProfile(profileArgs{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(1 * 24 * time.Hour))}}),
			want:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isProfileExpired(tt.prof, tt.minProfileDaysValid); got != tt.want {
				t.Errorf("checkProfileExpiry() = %v, want %v", got, tt.want)
			}
		})
	}
}
