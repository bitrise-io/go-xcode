package autocodesign

import (
	"testing"

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
