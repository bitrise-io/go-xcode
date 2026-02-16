package export_test

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptionsgenerator/internal/export"
	"github.com/stretchr/testify/require"
)

func TestCodeSignGroupPrinter_ToDebugString(t *testing.T) {
	tests := []struct {
		name  string
		group export.SelectableCodeSignGroup
		want  string
	}{
		{
			name: "empty group",
			group: export.SelectableCodeSignGroup{
				Certificate: certificateutil.CertificateInfoModel{
					CommonName: "CN",
					Serial:     "SERIAL",
					TeamID:     "TEAMID",
				},
				BundleIDProfilesMap: map[string][]profileutil.ProvisioningProfileInfoModel{
					"com.example.app": {
						{
							Name: "Profile 1",
							UUID: "UUID1",
						},
						{
							Name: "Profile 2",
							UUID: "UUID2",
						},
					},
					"com.example.appext": {{
						Name: "Profile 3",
						UUID: "UUID3",
					}},
				},
			},
			want: `{
	"bundle_id_profiles": {
		"com.example.app": [
			"Profile 1 (UUID1)",
			"Profile 2 (UUID2)"
		],
		"com.example.appext": [
			"Profile 3 (UUID3)"
		]
	},
	"certificate": "CN (SERIAL)",
	"team": " (TEAMID)"
}`,
		},
	}
	logger := log.NewLogger()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := export.NewCodeSignGroupPrinter(logger)
			got := printer.ToDebugString(tt.group)
			require.JSONEq(t, tt.want, got)
		})
	}
}
