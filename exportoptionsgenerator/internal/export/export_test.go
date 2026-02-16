package export_test

import (
	"testing"

	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptionsgenerator/internal/export"
	"github.com/stretchr/testify/require"
)

func TestCreateSelectableCodeSignGroups(t *testing.T) {
	certDev := certificateutil.CertificateInfoModel{
		CommonName: "iPhone Distribution: Bitrise Test (ABCD1234)",
		TeamID:     "ABCD1234",
	}
	profileDev := profileutil.ProvisioningProfileInfoModel{
		Name:                  "Bitrise Test Profile",
		UUID:                  "PROFILE-UUID-1234",
		TeamID:                "ABCD1234",
		BundleID:              "io.bitrise.testapp",
		ExportType:            exportoptions.MethodAppStore,
		DeveloperCertificates: []certificateutil.CertificateInfoModel{certDev},
	}

	tests := []struct {
		name         string
		certificates []certificateutil.CertificateInfoModel
		profiles     []profileutil.ProvisioningProfileInfoModel
		bundleIDs    []string
		filter       export.SelectableCodeSignGroupFilter
		want         []export.SelectableCodeSignGroup
	}{
		{
			name:         "empty inputs",
			certificates: []certificateutil.CertificateInfoModel{},
			profiles:     []profileutil.ProvisioningProfileInfoModel{},
			bundleIDs:    []string{},
			want:         []export.SelectableCodeSignGroup{},
		},
		{
			name:         "single matching profile and certificate",
			certificates: []certificateutil.CertificateInfoModel{certDev},
			profiles:     []profileutil.ProvisioningProfileInfoModel{profileDev},
			bundleIDs:    []string{"io.bitrise.testapp"},
			want: []export.SelectableCodeSignGroup{
				{
					Certificate: certDev,
					BundleIDProfilesMap: map[string][]profileutil.ProvisioningProfileInfoModel{
						"io.bitrise.testapp": {profileDev},
					},
				},
			},
		},
		{
			name: "filter by team ID, no match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
			},
			bundleIDs: []string{"io.bitrise.testapp"},
			filter:    export.CreateTeamSelectableCodeSignGroupFilter("WRONGID"),
			want:      []export.SelectableCodeSignGroup{},
		},
		{
			name: "filter by team ID, match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
			},
			bundleIDs: []string{"io.bitrise.testapp"},
			filter:    export.CreateTeamSelectableCodeSignGroupFilter("ABCD1234"),
			want: []export.SelectableCodeSignGroup{
				{
					Certificate: certDev,
					BundleIDProfilesMap: map[string][]profileutil.ProvisioningProfileInfoModel{
						"io.bitrise.testapp": {profileDev},
					},
				},
			},
		},
		{
			name: "filter out app store distribution",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
			},
			bundleIDs: []string{"io.bitrise.testapp"},
			filter:    export.CreateExportMethodSelectableCodeSignGroupFilter(exportoptions.MethodAdHoc),
			want:      []export.SelectableCodeSignGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := export.CreateSelectableCodeSignGroups(tt.certificates, tt.profiles, tt.bundleIDs)
			got = export.FilterSelectableCodeSignGroups(got, tt.filter)
			require.Equal(t, tt.want, got)
		})
	}
}
