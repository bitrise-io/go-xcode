package codesigngroup_test

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptionsgenerator/internal/codesigngroup"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/stretchr/testify/require"
)

func TestCreateSelectableCodeSignGroups(t *testing.T) {
	printer := codesigngroup.NewPrinter(log.NewLogger())
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
	profileExt := profileutil.ProvisioningProfileInfoModel{
		Name:                  "Bitrise Test Profile 2",
		UUID:                  "PROFILE-UUID-1235",
		TeamID:                "ABCD1234",
		BundleID:              "io.bitrise.testapp.appext",
		ExportType:            exportoptions.MethodAppStore,
		DeveloperCertificates: []certificateutil.CertificateInfoModel{certDev},
	}
	wildcarDev := profileutil.ProvisioningProfileInfoModel{
		Name:                  "Bitrise Test Profile *",
		UUID:                  "PROFILE-UUID-1236",
		TeamID:                "ABCD1234",
		BundleID:              "io.bitrise.*",
		ExportType:            exportoptions.MethodAppStore,
		DeveloperCertificates: []certificateutil.CertificateInfoModel{certDev},
	}
	managedProflile := profileutil.ProvisioningProfileInfoModel{
		Name:                  "iOS Team Provisioning Profile: io.bitrise.testapp", // managed by Xcode
		UUID:                  "PROFILE-UUID-1237",
		TeamID:                "ABCD1234",
		BundleID:              "io.bitrise.testapp",
		ExportType:            exportoptions.MethodAppStore,
		DeveloperCertificates: []certificateutil.CertificateInfoModel{certDev},
	}

	tests := []struct {
		name         string
		certificates []certificateutil.CertificateInfoModel
		profiles     []profileutil.ProvisioningProfileInfoModel
		bundleIDs    map[string]plistutil.PlistData
		filter       codesigngroup.GroupMapFunc
		want         []codesigngroup.Selectable
	}{
		{
			name:         "empty inputs",
			certificates: []certificateutil.CertificateInfoModel{},
			profiles:     []profileutil.ProvisioningProfileInfoModel{},
			bundleIDs:    map[string]plistutil.PlistData{},
			want:         []codesigngroup.Selectable(nil),
		},
		{
			name:         "single matching profile and certificate",
			certificates: []certificateutil.CertificateInfoModel{certDev},
			profiles:     []profileutil.ProvisioningProfileInfoModel{profileDev},
			bundleIDs: map[string]plistutil.PlistData{
				"io.bitrise.testapp": {},
			},
			want: []codesigngroup.Selectable{
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
			bundleIDs: map[string]plistutil.PlistData{"io.bitrise.testapp": {}},
			filter:    codesigngroup.CreateTeamIDFilter("WRONGID"),
			want:      []codesigngroup.Selectable(nil),
		},
		{
			name: "filter by team ID, match. prefers longer bundle ID match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
				profileExt,
				wildcarDev,
			},
			bundleIDs: map[string]plistutil.PlistData{
				"io.bitrise.testapp":        {},
				"io.bitrise.testapp.appext": {},
			},
			filter: codesigngroup.CreateTeamIDFilter("ABCD1234"),
			want: []codesigngroup.Selectable{
				{
					Certificate: certDev,
					BundleIDProfilesMap: map[string][]profileutil.ProvisioningProfileInfoModel{
						"io.bitrise.testapp":        {profileDev, wildcarDev},
						"io.bitrise.testapp.appext": {profileExt, wildcarDev},
					},
				},
			},
		},
		{
			name: "filter for ad-hoc dist, no match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
			},
			bundleIDs: map[string]plistutil.PlistData{"io.bitrise.testapp": {}},
			filter:    codesigngroup.CreateExportMethodFilter(exportoptions.MethodAdHoc),
			want:      []codesigngroup.Selectable(nil),
		},
		{
			name: "filter out non xcode managed profiles, no match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
				profileExt,
				managedProflile,
			},
			bundleIDs: map[string]plistutil.PlistData{
				"io.bitrise.testapp":        {},
				"io.bitrise.testapp.appext": {}, // no managed profile provided
			},
			filter: codesigngroup.CreateXcodeManagedFilter(),
			want:   []codesigngroup.Selectable(nil),
		},
		{
			name: "filter out xcode managed profiles, match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
				managedProflile,
			},
			bundleIDs: map[string]plistutil.PlistData{"io.bitrise.testapp": {}},
			filter:    codesigngroup.CreateNonXcodeManagedFilter(),
			want: []codesigngroup.Selectable{
				{
					Certificate: certDev,
					BundleIDProfilesMap: map[string][]profileutil.ProvisioningProfileInfoModel{
						"io.bitrise.testapp": {profileDev},
					},
				},
			},
		},
		{
			name: "filter out profile by name, match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
				profileExt,
			},
			bundleIDs: map[string]plistutil.PlistData{"io.bitrise.testapp": {}},
			filter:    codesigngroup.CreateExcludeProfileNameFilter("Nonmathcing Profile"),
			want: []codesigngroup.Selectable{
				{
					Certificate: certDev,
					BundleIDProfilesMap: map[string][]profileutil.ProvisioningProfileInfoModel{
						"io.bitrise.testapp": {profileDev},
					},
				},
			},
		},
		{
			name: "filter out profile by name, no match",
			certificates: []certificateutil.CertificateInfoModel{
				certDev,
			},
			profiles: []profileutil.ProvisioningProfileInfoModel{
				profileDev,
				profileExt,
			},
			bundleIDs: map[string]plistutil.PlistData{
				"io.bitrise.testapp":        {},
				"io.bitrise.testapp.appext": {},
			},
			filter: codesigngroup.CreateExcludeProfileNameFilter("Bitrise Test Profile"),
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codesigngroup.BuildFilterableList(tt.certificates, tt.profiles, tt.bundleIDs)
			t.Logf("groups: %s", printer.ListToDebugString(got))
			got = codesigngroup.MapGroups(got, tt.filter)
			t.Logf("filtered groups: %s", printer.ListToDebugString(got))
			require.JSONEq(t, printer.ListToDebugString(tt.want), printer.ListToDebugString(got))
		})
	}
}
