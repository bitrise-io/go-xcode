package appstoreconnectclient

import (
	"testing"

	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
)

func Test_checkBundleIDEntitlements(t *testing.T) {
	tests := []struct {
		name                 string
		bundleIDEntitlements []appstoreconnect.BundleIDCapability
		projectEntitlements  autocodesign.Entitlement
		wantErr              bool
	}{
		{
			name:                 "Check known entitlements, which does not need to be registered on the Developer Portal",
			bundleIDEntitlements: []appstoreconnect.BundleIDCapability{},
			projectEntitlements: autocodesign.Entitlement(map[string]interface{}{
				"keychain-access-groups":                             "",
				"com.apple.developer.ubiquity-kvstore-identifier":    "",
				"com.apple.developer.icloud-container-identifiers":   "",
				"com.apple.developer.ubiquity-container-identifiers": "",
			}),
			wantErr: false,
		},
		{
			name:                 "Needed to register entitlements",
			bundleIDEntitlements: []appstoreconnect.BundleIDCapability{},
			projectEntitlements: autocodesign.Entitlement(map[string]interface{}{
				"com.apple.developer.applesignin": "",
			}),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkBundleIDEntitlements(tt.bundleIDEntitlements, tt.projectEntitlements)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkBundleIDEntitlements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if mErr, ok := err.(autocodesign.NonmatchingProfileError); !ok {
					t.Errorf("checkBundleIDEntitlements() error = %v, it is not expected type", mErr)
				}
			}
		})
	}
}
