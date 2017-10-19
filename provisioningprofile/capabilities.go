package provisioningprofile

import (
	"github.com/bitrise-tools/go-xcode/plistutil"
)

// MatchTargetAndProfileEntitlements ...
func MatchTargetAndProfileEntitlements(targetEntitlements plistutil.PlistData, profileEntitlements plistutil.PlistData) []string {
	missingEntitlements := []string{}
	for key := range targetEntitlements {
		_, found := profileEntitlements[key]
		if !found {
			missingEntitlements = append(missingEntitlements, key)
		}
	}
	return missingEntitlements
}

// KnownTargetCapabilityProfileCapabilityMapping ...
var KnownTargetCapabilityProfileCapabilityMapping = map[string]interface{}{
	"com.apple.ApplePay":                         "com.apple.developer.in-app-payments",
	"com.apple.ApplicationGroups.iOS":            "com.apple.security.application-groups",
	"com.apple.BackgroundModes":                  "",
	"com.apple.DataProtection":                   "com.apple.developer.default-data-protection",
	"com.apple.GameCenter":                       "",
	"com.apple.HealthKit":                        "com.apple.developer.healthkit",
	"com.apple.HomeKit":                          "com.apple.developer.homekit",
	"com.apple.HotspotConfiguration":             "com.apple.developer.networking.HotspotConfiguration",
	"com.apple.InAppPurchase":                    "",
	"com.apple.InterAppAudio":                    "inter-app-audio",
	"com.apple.Keychain":                         "keychain-access-groups",
	"com.apple.Maps.iOS":                         "",
	"com.apple.Multipath":                        "com.apple.developer.networking.multipath",
	"com.apple.NearFieldCommunicationTagReading": "com.apple.developer.nfc.readersession.formats",
	"com.apple.NetworkExtensions.iOS":            "com.apple.developer.networking.networkextension",
	"com.apple.Push":                             "aps-environment",
	"com.apple.SafariKeychain":                   "com.apple.developer.associated-domains",
	"com.apple.Siri":                             "com.apple.developer.siri",
	"com.apple.VPNLite":                          "com.apple.developer.networking.vpn.api",
	"com.apple.WAC":                              "com.apple.external-accessory.wireless-configuration",
	"com.apple.Wallet":                           "com.apple.developer.pass-type-identifiers",
	"com.apple.iCloud":                           "com.apple.developer.icloud-container-identifiers",

	"com.apple.BackgroundModes.watchos.extension": "",
	"com.apple.HealthKit.watchos":                 "com.apple.developer.healthkit",
}
