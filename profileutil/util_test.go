package profileutil

import (
	"testing"

	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/stretchr/testify/require"
)

func TestIsXcodeManaged(t *testing.T) {
	xcodeManagedNames := []string{
		"XC iOS: custom.bundle.id",
		"XC tvOS: custom.bundle.id",
		"iOS Team Provisioning Profile: another.custom.bundle.id",
		"tvOS Team Provisioning Profile: another.custom.bundle.id",
		"iOS Team Store Provisioning Profile: my.bundle.id",
		"tvOS Team Store Provisioning Profile: my.bundle.id",
		"Mac Team Provisioning Profile: my.bundle.id",
		"Mac Team Store Provisioning Profile: my.bundle.id",
		"Mac Catalyst Team Provisioning Profile: my.bundle.id",
	}
	nonXcodeManagedNames := []string{
		"Test Profile Name",
		"iOS Distribution Profile: test.bundle.id",
		"iOS Dev",
		"tvOS Distribution Profile: test.bundle.id",
		"tvOS Dev",
		"Mac Distribution Profile: test.bundle.id",
		"Mac Dev",
	}

	for _, profileName := range xcodeManagedNames {
		require.Equal(t, true, IsXcodeManaged(profileName))
	}

	for _, profileName := range nonXcodeManagedNames {
		require.Equal(t, false, IsXcodeManaged(profileName))
	}
}

func TestMatchTargetAndProfileEntitlements(t *testing.T) {
	tests := []struct {
		name                string
		targetEntitlements  plistutil.PlistData
		profileEntitlements plistutil.PlistData
		profileType         ProfileType
		want                []string
	}{
		{
			name:                "empty target entitlements",
			targetEntitlements:  plistutil.PlistData{},
			profileEntitlements: plistutil.PlistData{},
			profileType:         ProfileTypeIos,
			want:                []string{},
		},
		{
			name:                "known iOS key present in profile",
			targetEntitlements:  plistutil.PlistData{"aps-environment": "production"},
			profileEntitlements: plistutil.PlistData{"aps-environment": "production"},
			profileType:         ProfileTypeIos,
			want:                []string{},
		},
		{
			name:                "known iOS key missing from profile",
			targetEntitlements:  plistutil.PlistData{"aps-environment": "production"},
			profileEntitlements: plistutil.PlistData{},
			profileType:         ProfileTypeIos,
			want:                []string{"aps-environment"},
		},
		{
			name:                "unknown key is ignored",
			targetEntitlements:  plistutil.PlistData{"com.custom.entitlement": true},
			profileEntitlements: plistutil.PlistData{},
			profileType:         ProfileTypeIos,
			want:                []string{},
		},
		{
			name:                "macOS: known macOS key missing from profile",
			targetEntitlements:  plistutil.PlistData{"com.apple.developer.aps-environment": "production"},
			profileEntitlements: plistutil.PlistData{},
			profileType:         ProfileTypeMacOs,
			want:                []string{"com.apple.developer.aps-environment"},
		},
		{
			name:                "iOS key ignored for macOS profile",
			targetEntitlements:  plistutil.PlistData{"aps-environment": "production"},
			profileEntitlements: plistutil.PlistData{},
			profileType:         ProfileTypeMacOs,
			want:                []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchTargetAndProfileEntitlements(tt.targetEntitlements, tt.profileEntitlements, tt.profileType)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}
