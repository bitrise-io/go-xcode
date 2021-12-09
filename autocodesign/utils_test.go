package autocodesign

import (
	"testing"

	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/stretchr/testify/assert"
)

func Test_GivenCodeSignAssets_WhenMergingTwo_ThenValuesAreCorrect(t *testing.T) {
	dev1Profile := profile("base", "1")
	dev2Profile := profile("addition", "4")
	devUITest1Profile := profile("base", "2")
	devUITest2Profile := profile("addition-uitest", "5")
	enterprise1Profile := profile("enterprise", "1")
	adHoc1Profile := profile("ad-hoc", "1")

	certificate := certificateutil.CertificateInfoModel{}
	tests := []struct {
		name     string
		base     *AppCodesignAssets
		addition *AppCodesignAssets
		expected *AppCodesignAssets
	}{
		{
			name: "Two existing assets with overlapping values",
			base: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"dev-1": dev1Profile,
				},
				UITestTargetProfilesByBundleID: map[string]Profile{
					"dev-uitest-1": devUITest1Profile,
				},
				Certificate: certificate,
			},
			addition: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"dev-2": dev2Profile,
				},
				UITestTargetProfilesByBundleID: map[string]Profile{
					"dev-uitest-2": devUITest2Profile,
				},
				Certificate: certificate,
			},
			expected: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"dev-1": dev1Profile,
					"dev-2": dev2Profile,
				},
				UITestTargetProfilesByBundleID: map[string]Profile{
					"dev-uitest-1": devUITest1Profile,
					"dev-uitest-2": devUITest2Profile,
				},
				Certificate: certificate,
			},
		},
		{
			name: "Base value is empty",
			base: nil,
			addition: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"enterprise-1": enterprise1Profile,
				},
				UITestTargetProfilesByBundleID: nil,
				Certificate:                    certificate,
			},
			expected: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"enterprise-1": enterprise1Profile,
				},
				UITestTargetProfilesByBundleID: nil,
				Certificate:                    certificate,
			},
		},
		{
			name: "Additional value is empty",
			base: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"ad-hoc-1": adHoc1Profile,
				},
				UITestTargetProfilesByBundleID: nil,
				Certificate:                    certificate,
			},
			addition: nil,
			expected: &AppCodesignAssets{
				ArchivableTargetProfilesByBundleID: map[string]Profile{
					"ad-hoc-1": adHoc1Profile,
				},
				UITestTargetProfilesByBundleID: nil,
				Certificate:                    certificate,
			},
		},
		{
			name:     "Empty values",
			base:     nil,
			addition: nil,
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merged := mergeCodeSignAssets(test.base, test.addition)
			assert.Equal(t, test.expected, merged)
		})
	}
}

func profile(name, id string) Profile {
	return newMockProfile(profileArgs{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           name,
			UUID:           id,
			ProfileContent: []byte{},
			Platform:       "",
			ExpirationDate: appstoreconnect.Time{},
		},
		id:           id,
		appID:        appstoreconnect.BundleID{},
		devices:      nil,
		certificates: nil,
		entitlements: Entitlements{},
	})
}
