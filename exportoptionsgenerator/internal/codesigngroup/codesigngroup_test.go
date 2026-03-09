package codesigngroup_test

import (
	"testing"

	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptionsgenerator/internal/codesigngroup"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
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
		filter       codesigngroup.SelectableCodeSignGroupFilter
		want         []codesigngroup.SelectableCodeSignGroup
	}{
		{
			name:         "empty inputs",
			certificates: []certificateutil.CertificateInfoModel{},
			profiles:     []profileutil.ProvisioningProfileInfoModel{},
			bundleIDs:    []string{},
			want:         []codesigngroup.SelectableCodeSignGroup(nil),
		},
		{
			name:         "single matching profile and certificate",
			certificates: []certificateutil.CertificateInfoModel{certDev},
			profiles:     []profileutil.ProvisioningProfileInfoModel{profileDev},
			bundleIDs:    []string{"io.bitrise.testapp"},
			want: []codesigngroup.SelectableCodeSignGroup{
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
			filter:    codesigngroup.CreateTeamSelectableCodeSignGroupFilter("WRONGID"),
			want:      []codesigngroup.SelectableCodeSignGroup(nil),
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
			filter:    codesigngroup.CreateTeamSelectableCodeSignGroupFilter("ABCD1234"),
			want: []codesigngroup.SelectableCodeSignGroup{
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
			filter:    codesigngroup.CreateExportMethodSelectableCodeSignGroupFilter(exportoptions.MethodAdHoc),
			want:      []codesigngroup.SelectableCodeSignGroup(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codesigngroup.BuildFilterableList(tt.certificates, tt.profiles, tt.bundleIDs)
			got = codesigngroup.Filter(got, tt.filter)
			require.Equal(t, tt.want, got)
		})
	}
}
