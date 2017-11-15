package profileutil

import (
	"github.com/bitrise-tools/go-xcode/plistutil"
)

// MatchTargetAndProfileEntitlements ...
func MatchTargetAndProfileEntitlements(targetEntitlements plistutil.PlistData, profileEntitlements plistutil.PlistData) []string {
	missingEntitlements := []string{}
	for key := range targetEntitlements {
		_, exclude := KnownTargetCapabilityExclusions[key]
		if exclude {
			continue
		}
		_, found := profileEntitlements[key]
		if !found {
			missingEntitlements = append(missingEntitlements, key)
		}
	}
	return missingEntitlements
}

// KnownTargetCapabilityExclusions ...
var KnownTargetCapabilityExclusions = map[string]bool{
	"com.apple.security.get-task-allow": true,
}

// KnownProfileCapabilitiesMap ...
var KnownProfileCapabilitiesMap = map[string]bool{
	"com.apple.developer.in-app-payments":                 true,
	"com.apple.security.application-groups":               true,
	"com.apple.developer.default-data-protection":         true,
	"com.apple.developer.healthkit":                       true,
	"com.apple.developer.homekit":                         true,
	"com.apple.developer.networking.HotspotConfiguration": true,
	"inter-app-audio":                                     true,
	"keychain-access-groups":                              true,
	"com.apple.developer.networking.multipath":            true,
	"com.apple.developer.nfc.readersession.formats":       true,
	"com.apple.developer.networking.networkextension":     true,
	"aps-environment":                                     true,
	"com.apple.developer.associated-domains":              true,
	"com.apple.developer.siri":                            true,
	"com.apple.developer.networking.vpn.api":              true,
	"com.apple.external-accessory.wireless-configuration": true,
	"com.apple.developer.pass-type-identifiers":           true,
	"com.apple.developer.icloud-container-identifiers":    true,
}
