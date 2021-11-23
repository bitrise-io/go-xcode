package autocodesign

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/stretchr/testify/assert"
)

func Test_GivenCodeSignAssets_WhenMergingTwo_ThenValuesAreCorrect(t *testing.T) {
	certificate := certificateutil.CertificateInfoModel{}
	tests := []struct {
		name     string
		base     map[DistributionType]AppCodesignAssets
		addition map[DistributionType]AppCodesignAssets
		expected map[DistributionType]AppCodesignAssets
	}{
		{
			name: "Two existing assets with overlapping values",
			base: map[DistributionType]AppCodesignAssets{
				Development: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"dev-1": profile("base", "1"),
					},
					UITestTargetProfilesByBundleID: map[string]Profile{
						"dev-uitest-1": profile("base", "2"),
					},
					Certificate: certificate,
				}},
			addition: map[DistributionType]AppCodesignAssets{
				Development: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"dev-2": profile("addition", "4"),
					},
					UITestTargetProfilesByBundleID: map[string]Profile{
						"dev-uitest-2": profile("addition-uitest", "5"),
					},
					Certificate: certificate,
				},
				AppStore: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"app-store-1": profile("base", "3"),
					},
					UITestTargetProfilesByBundleID: nil,
					Certificate:                    certificate,
				},
			},
			expected: map[DistributionType]AppCodesignAssets{
				Development: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"dev-1": profile("base", "1"),
						"dev-2": profile("addition", "4"),
					},
					UITestTargetProfilesByBundleID: map[string]Profile{
						"dev-uitest-1": profile("base", "2"),
						"dev-uitest-2": profile("addition-uitest", "5"),
					},
					Certificate: certificate,
				},
				AppStore: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"app-store-1": profile("base", "3"),
					},
					UITestTargetProfilesByBundleID: nil,
					Certificate:                    certificate,
				},
			},
		},
		{
			name: "Base value is empty",
			base: map[DistributionType]AppCodesignAssets{},
			addition: map[DistributionType]AppCodesignAssets{
				Enterprise: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"enterprise-1": profile("enterprise", "1"),
					},
					UITestTargetProfilesByBundleID: nil,
					Certificate:                    certificate,
				},
			},
			expected: map[DistributionType]AppCodesignAssets{
				Enterprise: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"enterprise-1": profile("enterprise", "1"),
					},
					UITestTargetProfilesByBundleID: nil,
					Certificate:                    certificate,
				},
			},
		},
		{
			name: "Additional value is empty",
			base: map[DistributionType]AppCodesignAssets{
				AdHoc: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"ad-hoc-1": profile("ad-hoc", "1"),
					},
					UITestTargetProfilesByBundleID: nil,
					Certificate:                    certificate,
				},
			},
			addition: map[DistributionType]AppCodesignAssets{},
			expected: map[DistributionType]AppCodesignAssets{
				AdHoc: {
					ArchivableTargetProfilesByBundleID: map[string]Profile{
						"ad-hoc-1": profile("ad-hoc", "1"),
					},
					UITestTargetProfilesByBundleID: nil,
					Certificate:                    certificate,
				},
			},
		},
		{
			name:     "Empty values",
			base:     map[DistributionType]AppCodesignAssets{},
			addition: map[DistributionType]AppCodesignAssets{},
			expected: map[DistributionType]AppCodesignAssets{},
		},
	}

	for _, test := range tests {
		merged := mergeCodeSignAssets(test.base, test.addition)
		assert.Equal(t, test.expected, merged)
	}
}

func profile(name, id string) testProfile {
	return testProfile{
		attributes: appstoreconnect.ProfileAttributes{
			Name:           name,
			UUID:           id,
			ProfileContent: []byte{},
			Platform:       "",
			ExpirationDate: appstoreconnect.Time{},
		},
		id:             id,
		bundleID:       fmt.Sprintf("bundle-id-%s", id),
		certificateIDs: nil,
		deviceIDs:      nil,
	}
}

type testProfile struct {
	attributes     appstoreconnect.ProfileAttributes
	id             string
	bundleID       string
	deviceIDs      []string
	certificateIDs []string
}

// ID ...
func (p testProfile) ID() string {
	return p.id
}

// Attributes ...
func (p testProfile) Attributes() appstoreconnect.ProfileAttributes {
	return p.attributes
}

// CertificateIDs ...
func (p testProfile) CertificateIDs() ([]string, error) {
	return p.certificateIDs, nil
}

// DeviceIDs ...
func (p testProfile) DeviceIDs() ([]string, error) {
	return p.deviceIDs, nil
}

// BundleID ...
func (p testProfile) BundleID() (appstoreconnect.BundleID, error) {
	return appstoreconnect.BundleID{}, nil
}

// Entitlements ...
func (p testProfile) Entitlements() (Entitlements, error) {
	return Entitlements{}, nil
}
