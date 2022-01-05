package projectmanager

import (
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign"
)

func TestCanGenerateProfileWithEntitlements(t *testing.T) {
	tests := []struct {
		name                   string
		entitlementsByBundleID map[string]autocodesign.Entitlements
		wantOk                 bool
		wantEntitlement        string
		wantBundleID           string
	}{
		{
			name: "no entitlements",
			entitlementsByBundleID: map[string]autocodesign.Entitlements{
				"com.bundleid": map[string]interface{}{},
			},
			wantOk:          true,
			wantEntitlement: "",
			wantBundleID:    "",
		},
		{
			name: "contains unsupported entitlement",
			entitlementsByBundleID: map[string]autocodesign.Entitlements{
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
			entitlementsByBundleID: map[string]autocodesign.Entitlements{
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
			entitlementsByBundleID: map[string]autocodesign.Entitlements{
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
