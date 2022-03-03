package profileutil

import (
	"testing"

	"github.com/fullsailor/pkcs7"

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

func TestProvisioningProfilePlatform(t *testing.T) {
	tests := []struct {
		name           string
		profileContent string
		want           ProfileType
	}{
		{
			name:           "iOS",
			profileContent: iosProfileContent,
			want:           ProfileTypeIos,
		},
		{
			name:           "macOS",
			profileContent: macosProfileContent,
			want:           ProfileTypeMacOs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profilePkcs7 := pkcs7.PKCS7{Content: []byte(tt.profileContent)}
			got, err := NewProvisioningProfileInfo(profilePkcs7)

			require.NoError(t, err)
			require.Equal(t, tt.want, got.Type)
		})
	}
}

const iosProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>Bitrise Test</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>9NS44DLTN7</string>
	</array>
	<key>CreationDate</key>
	<date>2016-09-22T11:28:46Z</date>
	<key>Platform</key>
	<array>
		<string>iOS</string>
	</array>
	<key>DeveloperCertificates</key>
	<array>
		<data></data>
	</array>
	<key>Entitlements</key>
	<dict>
		<key>keychain-access-groups</key>
		<array>
			<string>9NS44DLTN7.*</string>
		</array>
		<key>get-task-allow</key>
		<true/>
		<key>application-identifier</key>
		<string>9NS44DLTN7.*</string>
		<key>com.apple.developer.team-identifier</key>
		<string>9NS44DLTN7</string>
	</dict>
	<key>ExpirationDate</key>
	<date>2017-09-22T11:28:46Z</date>
	<key>Name</key>
	<string>Bitrise Test Development</string>
	<key>ProvisionedDevices</key>
	<array>
		<string>b13813075ad9b298cb9a9f28555c49573d8bc322</string>
	</array>
	<key>TeamIdentifier</key>
	<array>
		<string>9NS44DLTN7</string>
	</array>
	<key>TeamName</key>
	<string>Some Dude</string>
	<key>TimeToLive</key>
	<integer>365</integer>
	<key>UUID</key>
	<string>4b617a5f-e31e-4edc-9460-718a5abacd05</string>
	<key>Version</key>
	<integer>1</integer>
</dict>`

const macosProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>AppIDName</key>
	<string>XC io bitrise mobile ios QuickActionsTodayExtension</string>
	<key>ApplicationIdentifierPrefix</key>
	<array>
	<string>72SA8V3WYL</string>
	</array>
	<key>CreationDate</key>
	<date>2022-02-28T10:35:39Z</date>
	<key>Platform</key>
	<array>
			<string>OSX</string>
	</array>
	<key>IsXcodeManaged</key>
	<false/>
	<key>DeveloperCertificates</key>
	<array>
		<data></data>
	</array>

	<key>DER-Encoded-Profile</key>
	<data></data>
														
	<key>Entitlements</key>
	<dict>
				
				<key>com.apple.developer.game-center</key>
		<true/>
				
				<key>com.apple.security.application-groups</key>
		<array>
				<string>group.io.bitrise.statistics</string>
		</array>
				
				<key>application-identifier</key>
		<string>72SA8V3WYL.io.bitrise.mobile.ios.QuickActionsTodayExtension</string>
				
				<key>com.apple.application-identifier</key>
		<string>72SA8V3WYL.io.bitrise.mobile.ios.QuickActionsTodayExtension</string>
				
				<key>keychain-access-groups</key>
		<array>
				<string>72SA8V3WYL.*</string>
				<string>com.apple.token</string>
		</array>
				
				<key>get-task-allow</key>
		<true/>
				
				<key>com.apple.developer.team-identifier</key>
		<string>72SA8V3WYL</string>

	</dict>
	<key>ExpirationDate</key>
	<date>2023-02-28T10:35:39Z</date>
	<key>Name</key>
	<string>_profile_bug_type_catalyst</string>
	<key>ProvisionedDevices</key>
	<array>
		<string>BA0EC799-F254-5574-B335-E70B8A2FA5E7</string>
	</array>
	<key>TeamIdentifier</key>
	<array>
		<string>72SA8V3WYL</string>
	</array>
	<key>TeamName</key>
	<string>BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG</string>
	<key>TimeToLive</key>
	<integer>365</integer>
	<key>UUID</key>
	<string>dea6a48c-d7d3-4624-9f6b-e0c3b3ce517d</string>
	<key>Version</key>
	<integer>1</integer>
</dict>
</plist>`
