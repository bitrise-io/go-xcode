package autocodesign

import (
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
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
			got, err := createWildcardBundleID(tt.bundleID)
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
		wantErr     bool
	}{
		{
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS ad-hoc - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS enterprise - (io.bitrise.app)",
			wantErr:     false,
		},

		{
			profileType: appstoreconnect.TvOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS ad-hoc - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS enterprise - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.ProfileType("unknown"),
			bundleID:    "io.bitrise.app",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.profileType), func(t *testing.T) {
			got, err := profileName(tt.profileType, tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("profileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("profileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findMissingContainers(t *testing.T) {
	tests := []struct {
		name        string
		projectEnts serialized.Object
		profileEnts serialized.Object
		want        []string
		wantErr     bool
	}{
		{
			name: "equal without container",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "equal with container",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "profile has more containers than project",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),

			want:    nil,
			wantErr: false,
		},
		{
			name: "project has more containers than profile",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{},
			}),

			want:    []string{"container1"},
			wantErr: false,
		},
		{
			name: "project has containers but profile doesn't",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": []interface{}{"container1"},
			}),
			profileEnts: serialized.Object(map[string]interface{}{
				"otherentitlement": "",
			}),

			want:    []string{"container1"},
			wantErr: false,
		},
		{
			name: "error check",
			projectEnts: serialized.Object(map[string]interface{}{
				"com.apple.developer.icloud-container-identifiers": "break",
			}),

			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findMissingContainers(tt.projectEnts, tt.profileEnts)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, got, tt.want)
		})
	}
}

type MockProfile struct {
	attributes appstoreconnect.ProfileAttributes
}

func (m MockProfile) ID() string {
	return ""
}

func (m MockProfile) Attributes() appstoreconnect.ProfileAttributes {
	return m.attributes
}

func (m MockProfile) CertificateIDs() (map[string]bool, error) {
	return nil, nil
}

func (m MockProfile) DeviceIDs() (map[string]bool, error) {
	return nil, nil
}

func (m MockProfile) BundleID() (appstoreconnect.BundleID, error) {
	return appstoreconnect.BundleID{}, nil
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
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(5 * time.Hour))}},
			want:                false,
		},
		{
			name:                "no days set - profile expiry date before current time",
			minProfileDaysValid: 0,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(-5 * time.Hour))}},
			want:                true,
		},
		{
			name:                "days set - profile expiry date after current time + days set",
			minProfileDaysValid: 2,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(5 * 24 * time.Hour))}},
			want:                false,
		},
		{
			name:                "days set - profile expiry date before current time + days set",
			minProfileDaysValid: 2,
			prof:                MockProfile{attributes: appstoreconnect.ProfileAttributes{ExpirationDate: appstoreconnect.Time(time.Now().Add(1 * 24 * time.Hour))}},
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

func TestCanGenerateProfileWithEntitlements(t *testing.T) {
	tests := []struct {
		name                   string
		entitlementsByBundleID map[string]serialized.Object
		wantOk                 bool
		wantEntitlement        string
		wantBundleID           string
	}{
		{
			name: "no entitlements",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{},
			},
			wantOk:          true,
			wantEntitlement: "",
			wantBundleID:    "",
		},
		{
			name: "contains unsupported entitlement",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{
					"com.entitlement-ignored":            true,
					"com.apple.developer.contacts.notes": true,
				},
			},
			wantOk:          false,
			wantEntitlement: "com.apple.developer.contacts.notes",
			wantBundleID:    "com.bundleid",
		},
		{
			name: "contains unsupported entitlement, multiple bundle IDs",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{
					"aps-environment": true,
				},
				"com.bundleid2": map[string]interface{}{
					"com.entitlement-ignored":            true,
					"com.apple.developer.contacts.notes": true,
				},
			},
			wantOk:          false,
			wantEntitlement: "com.apple.developer.contacts.notes",
			wantBundleID:    "com.bundleid2",
		},
		{
			name: "all entitlements supported",
			entitlementsByBundleID: map[string]serialized.Object{
				"com.bundleid": map[string]interface{}{
					"aps-environment": true,
				},
			},
			wantOk:          true,
			wantEntitlement: "",
			wantBundleID:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, gotEntilement, gotBundleID := CanGenerateProfileWithEntitlements(tt.entitlementsByBundleID)
			if gotOk != tt.wantOk {
				t.Errorf("CanGenerateProfileWithEntitlements() got = %v, want %v", gotOk, tt.wantOk)
			}
			if gotEntilement != tt.wantEntitlement {
				t.Errorf("CanGenerateProfileWithEntitlements() got1 = %v, want %v", gotEntilement, tt.wantEntitlement)
			}
			if gotBundleID != tt.wantBundleID {
				t.Errorf("CanGenerateProfileWithEntitlements() got2 = %v, want %v", gotBundleID, tt.wantBundleID)
			}
		})
	}
}
