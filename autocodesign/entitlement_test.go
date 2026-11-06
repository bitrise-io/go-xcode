package autocodesign

import (
	"errors"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/require"
)

func TestICloudContainers(t *testing.T) {
	tests := []struct {
		name                string
		projectEntitlements Entitlements
		want                []string
		errHandler          func(require.TestingT, error, ...interface{})
	}{
		{
			name:                "no containers",
			projectEntitlements: Entitlements(map[string]interface{}{}),
			want:                nil,
			errHandler:          require.NoError,
		},
		{
			name: "no containers - CloudDocuments",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudDocuments",
				},
			}),
			want:       nil,
			errHandler: require.NoError,
		},
		{
			name: "no containers - CloudKit",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
				},
			}),
			want:       nil,
			errHandler: require.NoError,
		},
		{
			name: "no containers - CloudKit and CloudDocuments",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
					"CloudDocuments",
				},
			}),
			want:       nil,
			errHandler: require.NoError,
		},
		{
			name: "has containers - CloudDocuments",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudDocuments",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
					"iCloud.test.container.id2"},
			}),
			want:       []string{"iCloud.test.container.id", "iCloud.test.container.id2"},
			errHandler: require.NoError,
		},
		{
			name: "has containers - CloudKit",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
					"iCloud.test.container.id2"},
			}),
			want:       []string{"iCloud.test.container.id", "iCloud.test.container.id2"},
			errHandler: require.NoError,
		},
		{
			name: "has containers - CloudKit and CloudDocuments",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
					"CloudDocuments",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
					"iCloud.test.container.id2"},
			}),
			want:       []string{"iCloud.test.container.id", "iCloud.test.container.id2"},
			errHandler: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.projectEntitlements.ICloudContainers()
			require.Equal(t, got, tt.want)
			tt.errHandler(t, err)
		})
	}
}

func TestCapability_HealthKitAccessIgnored(t *testing.T) {
	ent := Entitlement(map[string]interface{}{
		"com.apple.developer.healthkit.access": []interface{}{"health-records"},
	})
	cap, err := ent.Capability()
	require.NoError(t, err)
	require.Nil(t, cap)
}

func TestCapability_UnknownKeyReturnsSentinel(t *testing.T) {
	ent := Entitlement(map[string]interface{}{
		"com.apple.developer.totally-made-up-entitlement": true,
	})

	cap, err := ent.Capability()
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnknownEntitlementKey))
	require.Contains(t, err.Error(), "com.apple.developer.totally-made-up-entitlement")
	require.Nil(t, cap)
}

func TestCapability_IgnoredKeys(t *testing.T) {
	keys := []string{
		"com.apple.developer.kernel.extended-virtual-addressing",
		"com.apple.developer.kernel.increased-memory-limit",
		"com.apple.developer.authentication-services.credential-provider-ui",
	}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			ent := Entitlement(map[string]interface{}{key: true})
			cap, err := ent.Capability()
			require.NoError(t, err)
			require.Nil(t, cap)
			require.False(t, ent.AppearsOnDeveloperPortal())
		})
	}
}

func TestICloudContainers_wrongType(t *testing.T) {
	tests := []struct {
		name                string
		projectEntitlements Entitlements
		wantErr             string
	}{
		{
			name: "key-value store identifier is not a string",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.ubiquity-kvstore-identifier": true,
			}),
			wantErr: "com.apple.developer.ubiquity-kvstore-identifier is not a string: true (bool)",
		},
		{
			name: "iCloud services is not a string array",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services": "CloudKit",
			}),
			wantErr: `com.apple.developer.icloud-services is not a string array: "CloudKit" (string)`,
		},
		{
			name: "container identifiers is not a string array",
			projectEntitlements: Entitlements(map[string]interface{}{
				"com.apple.developer.icloud-services":              []interface{}{"CloudKit"},
				"com.apple.developer.icloud-container-identifiers": "iCloud.com.bitrise.app",
			}),
			wantErr: `com.apple.developer.icloud-container-identifiers is not a string array: "iCloud.com.bitrise.app" (string)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.projectEntitlements.ICloudContainers()
			require.EqualError(t, err, tt.wantErr)
			require.True(t, errors.As(err, &serialized.WrongTypeError{}), "callers can tell a wrong type apart")
			require.Nil(t, got)
		})
	}
}

func TestDataProtection_wrongType(t *testing.T) {
	ent := Entitlement(map[string]interface{}{
		"com.apple.developer.default-data-protection": true,
	})
	const wantErr = "com.apple.developer.default-data-protection is not a string: true (bool)"

	cap, err := ent.Capability()
	require.EqualError(t, err, wantErr)
	require.Nil(t, cap)

	equal, err := ent.Equal(appstoreconnect.BundleIDCapability{
		Attributes: appstoreconnect.BundleIDCapabilityAttributes{CapabilityType: appstoreconnect.DataProtection},
	}, Entitlements(ent))
	require.EqualError(t, err, wantErr)
	require.False(t, equal)
}
