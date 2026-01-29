package exportoptionsgenerator

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/v2/exportoptionsgenerator/mocks"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
	"github.com/stretchr/testify/require"
)

const (
	expectedDevelopmentXcode11ExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>development</string>
		<key>provisioningProfiles</key>
		<dict>
			<key>io.bundle.id</key>
			<string>Development Application Profile</string>
		</dict>
		<key>signingCertificate</key>
		<string>Development Certificate</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`
	expectedDevelopmentExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id</string>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>development</string>
		<key>provisioningProfiles</key>
		<dict>
			<key>io.bundle.id</key>
			<string>Development Application Profile</string>
		</dict>
		<key>signingCertificate</key>
		<string>Development Certificate</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`
	expectedAdHocExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id</string>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>ad-hoc</string>
		<key>provisioningProfiles</key>
		<dict>
			<key>io.bundle.id</key>
			<string>Development Application Profile</string>
		</dict>
		<key>signingCertificate</key>
		<string>Development Certificate</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`
	expectedXcode12AppStoreExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>app-store</string>
		<key>provisioningProfiles</key>
		<dict>
			<key>io.bundle.id</key>
			<string>Development Application Profile</string>
			<key>io.bundle.id.AppClipID</key>
			<string>Development App Clip Profile</string>
		</dict>
		<key>signingCertificate</key>
		<string>Development Certificate</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`
	expectedXcode13AppStoreExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>manageAppVersionAndBuildNumber</key>
		<false/>
		<key>method</key>
		<string>app-store</string>
		<key>provisioningProfiles</key>
		<dict>
			<key>io.bundle.id</key>
			<string>Development Application Profile</string>
			<key>io.bundle.id.AppClipID</key>
			<string>Development App Clip Profile</string>
		</dict>
		<key>signingCertificate</key>
		<string>Development Certificate</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`
	expectedNoProfilesDevelopmentXcode11ExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>development</string>
	</dict>
</plist>`
	expectedNoProfilesDevelopmentXcode16ExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id</string>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>debugging</string>
	</dict>
</plist>`
	expectedNoProfilesXcode13AppStoreExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>manageAppVersionAndBuildNumber</key>
		<false/>
		<key>method</key>
		<string>app-store</string>
	</dict>
</plist>`
	expectedNoProfilesXcode16AppStoreExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>manageAppVersionAndBuildNumber</key>
		<false/>
		<key>method</key>
		<string>app-store-connect</string>
	</dict>
</plist>`
	expectedNoProfilesAdHocExportOptions = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id</string>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>ad-hoc</string>
	</dict>
</plist>`
)

func newXcodeVersionReader(t *testing.T, major int64) xcodeversion.Reader {
	reader := mocks.NewXcodeVersionReader(t)
	reader.On("GetVersion").Return(xcodeversion.Version{Major: major}, nil)
	return reader
}

func TestExportOptionsGenerator_GenerateApplicationExportOptions_ForAutomaticSigningStyle(t *testing.T) {
	// Arrange
	const (
		bundleID     = "io.bundle.id"
		bundleIDClip = "io.bundle.id.AppClipID"
		teamID       = "TEAM123"
	)

	logger := log.NewLogger()
	logger.EnableDebugLog(true)

	tests := []struct {
		name                          string
		exportProduct                 ExportProduct
		archiveInfo                   ArchiveInfo
		exportMethod                  exportoptions.Method
		containerEnvironment          string
		xcodeVersion                  int64
		testFlightInternalTestingOnly bool
		ManageVersionAndBuildNumber   bool
		want                          string
		wantErr                       bool
	}{
		{
			name:          "Default development exportOptions",
			exportProduct: ExportProductApp,
			archiveInfo: ArchiveInfo{
				AppBundleID: bundleID,
			},
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 15,
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id</string>
		<key>method</key>
		<string>development</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`,
		},
		{
			name:          "App Clip, Default development exportOptions",
			exportProduct: ExportProductAppClip,
			archiveInfo: ArchiveInfo{
				AppBundleID:     bundleID,
				AppClipBundleID: bundleIDClip,
			},
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 15,
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id.AppClipID</string>
		<key>method</key>
		<string>development</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`,
		},
		{
			name:          "App Clip, Clip not present in archive",
			exportProduct: ExportProductAppClip,
			archiveInfo: ArchiveInfo{
				AppBundleID: bundleID,
			},
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 15,
			wantErr:      true,
		},
		{
			name:          "app store exportOptions, with managed version",
			exportProduct: ExportProductApp,
			archiveInfo: ArchiveInfo{
				AppBundleID: bundleID,
			},
			exportMethod:                exportoptions.MethodAppStore,
			ManageVersionAndBuildNumber: true,
			xcodeVersion:                15,
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>method</key>
		<string>app-store</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`,
		},
		{
			name:          "When the app uses iCloud services",
			exportProduct: ExportProductApp,
			archiveInfo: ArchiveInfo{
				AppBundleID: bundleID,
				EntitlementsByBundleID: map[string]plistutil.PlistData{
					bundleID: {"com.apple.developer.icloud-services": []string{"CloudKit"}},
				},
			},
			exportMethod:         exportoptions.MethodDevelopment,
			containerEnvironment: string(exportoptions.ICloudContainerEnvironmentProduction),
			xcodeVersion:         15,
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>distributionBundleIdentifier</key>
		<string>io.bundle.id</string>
		<key>iCloudContainerEnvironment</key>
		<string>Production</string>
		<key>method</key>
		<string>development</string>
		<key>teamID</key>
		<string>TEAM123</string>
	</dict>
</plist>`,
		},
		{
			name:          "When exporting for TestFlight internal testing only",
			exportProduct: ExportProductApp,
			archiveInfo: ArchiveInfo{
				AppBundleID: bundleID,
			},
			exportMethod:                  exportoptions.MethodAppStore,
			testFlightInternalTestingOnly: true,
			xcodeVersion:                  15,
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>manageAppVersionAndBuildNumber</key>
		<false/>
		<key>method</key>
		<string>app-store</string>
		<key>teamID</key>
		<string>TEAM123</string>
		<key>testFlightInternalTestingOnly</key>
		<true/>
	</dict>
</plist>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			g := New(newXcodeVersionReader(t, tt.xcodeVersion), logger)
			opts := Opts{
				ContainerEnvironment:             tt.containerEnvironment,
				TeamID:                           teamID,
				UploadBitcode:                    true,
				CompileBitcode:                   true,
				ArchivedWithXcodeManagedProfiles: false,
				TestFlightInternalTestingOnly:    tt.testFlightInternalTestingOnly,
				ManageVersionAndBuildNumber:      tt.ManageVersionAndBuildNumber,
			}

			// Act
			gotOpts, err := g.GenerateApplicationExportOptions(tt.exportProduct, tt.archiveInfo, tt.exportMethod, exportoptions.SigningStyleAutomatic, opts)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			got, err := gotOpts.String()
			require.NoError(t, err)
			fmt.Println(got)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExportOptionsGenerator_GenerateApplicationExportOptions(t *testing.T) {
	const (
		bundleID     = "io.bundle.id"
		bundleIDClip = "io.bundle.id.AppClipID"
		teamID       = "TEAM123"
	)

	certificate := certificateutil.CertificateInfoModel{Serial: "serial", CommonName: "Development Certificate", TeamID: teamID}

	tests := []struct {
		name         string
		exportMethod exportoptions.Method
		xcodeVersion int64
		want         string
		wantErr      bool
	}{
		{
			name:         "Development Xcode 11",
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 11,
			want:         expectedDevelopmentXcode11ExportOptions,
		},
		{
			name:         "Development Xcode > 12",
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 13,
			want:         expectedDevelopmentExportOptions,
		},
		{
			name:         "Ad-hoc",
			exportMethod: exportoptions.MethodAdHoc,
			xcodeVersion: 13,
			want:         expectedAdHocExportOptions,
		},
		{
			name:         "App-store Xcode 12",
			exportMethod: exportoptions.MethodAppStore,
			xcodeVersion: 12,
			want:         expectedXcode12AppStoreExportOptions,
		},
		{
			name:         "App-store Xcode 13",
			exportMethod: exportoptions.MethodAppStore,
			xcodeVersion: 13,
			want:         expectedXcode13AppStoreExportOptions,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := log.NewLogger()
			logger.EnableDebugLog(true)

			g := New(newXcodeVersionReader(t, tt.xcodeVersion), logger)
			g.certificateProvider = MockCodesignIdentityProvider{
				[]certificateutil.CertificateInfoModel{certificate},
			}
			profile := profileutil.ProvisioningProfileInfoModel{
				BundleID:              bundleID,
				TeamID:                teamID,
				ExportType:            tt.exportMethod,
				Name:                  "Development Application Profile",
				DeveloperCertificates: []certificateutil.CertificateInfoModel{certificate},
			}
			profileForClip := profileutil.ProvisioningProfileInfoModel{
				BundleID:              bundleIDClip,
				TeamID:                teamID,
				ExportType:            tt.exportMethod,
				Name:                  "Development App Clip Profile",
				DeveloperCertificates: []certificateutil.CertificateInfoModel{certificate},
			}
			g.profileProvider = MockProvisioningProfileProvider{
				[]profileutil.ProvisioningProfileInfoModel{
					profile,
					profileForClip,
				},
			}

			archiveInfo := ArchiveInfo{
				AppBundleID: bundleID,
				EntitlementsByBundleID: map[string]plistutil.PlistData{
					bundleID:     {"com.apple.developer.icloud-services": []string{"CloudKit"}},
					bundleIDClip: nil,
				},
				AppClipBundleID: bundleIDClip,
			}
			opts := Opts{
				ContainerEnvironment:             string(exportoptions.ICloudContainerEnvironmentProduction),
				TeamID:                           teamID,
				UploadBitcode:                    true,
				CompileBitcode:                   true,
				ArchivedWithXcodeManagedProfiles: false,
				TestFlightInternalTestingOnly:    false,
			}

			// Act
			gotOpts, err := g.GenerateApplicationExportOptions(ExportProductApp, archiveInfo, tt.exportMethod, exportoptions.SigningStyleManual, opts)

			// Assert
			require.NoError(t, err)

			got, err := gotOpts.String()
			require.NoError(t, err)
			fmt.Println(got)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExportOptionsGenerator_GenerateApplicationExportOptions_WhenNoProfileFound(t *testing.T) {
	const (
		bundleID     = "io.bundle.id"
		bundleIDClip = "io.bundle.id.AppClipID"
		teamID       = "TEAM123"
	)

	certificate := certificateutil.CertificateInfoModel{Serial: "serial", CommonName: "Development Certificate", TeamID: teamID}

	tests := []struct {
		name         string
		exportMethod exportoptions.Method
		xcodeVersion int64
		want         string
		wantErr      bool
	}{
		{
			name:         "When no profiles found, Xcode 16, using new export method name",
			exportMethod: exportoptions.MethodAppStore,
			xcodeVersion: 16,
			want:         expectedNoProfilesXcode16AppStoreExportOptions,
		},
		{
			name:         "When no profiles found, Xcode 13, then manageAppVersionAndBuildNumber is included",
			exportMethod: exportoptions.MethodAppStore,
			xcodeVersion: 13,
			want:         expectedNoProfilesXcode13AppStoreExportOptions,
		},
		{
			name:         "When no profiles found, Xcode > 12, distributionBundleIdentifier included",
			exportMethod: exportoptions.MethodAdHoc,
			xcodeVersion: 13,
			want:         expectedNoProfilesAdHocExportOptions,
		},
		{
			name:         "When no profiles found, Xcode 16, usess new export method name",
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 16,
			want:         expectedNoProfilesDevelopmentXcode16ExportOptions,
		},
		{
			name:         "When no profiles found, Xcode 11",
			exportMethod: exportoptions.MethodDevelopment,
			xcodeVersion: 11,
			want:         expectedNoProfilesDevelopmentXcode11ExportOptions,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := log.NewLogger()
			logger.EnableDebugLog(true)
			xcodeVersionReader := newXcodeVersionReader(t, tt.xcodeVersion)

			cloudKitEntitlement := map[string]interface{}{"com.apple.developer.icloud-services": []string{"CloudKit"}}
			g := New(xcodeVersionReader, logger)

			g.certificateProvider = MockCodesignIdentityProvider{
				[]certificateutil.CertificateInfoModel{certificate},
			}
			g.profileProvider = MockProvisioningProfileProvider{}

			archiveInfo := ArchiveInfo{
				AppBundleID: bundleID,
				EntitlementsByBundleID: map[string]plistutil.PlistData{
					bundleID:     cloudKitEntitlement,
					bundleIDClip: nil,
				},
				AppClipBundleID: bundleIDClip,
			}
			opts := Opts{
				ContainerEnvironment:             string(exportoptions.ICloudContainerEnvironmentProduction),
				TeamID:                           teamID,
				UploadBitcode:                    true,
				CompileBitcode:                   true,
				ArchivedWithXcodeManagedProfiles: false,
				TestFlightInternalTestingOnly:    true,
			}

			// Act
			gotOpts, err := g.GenerateApplicationExportOptions(ExportProductApp, archiveInfo, tt.exportMethod, exportoptions.SigningStyleManual, opts)

			// Assert
			require.NoError(t, err)

			got, err := gotOpts.String()
			require.NoError(t, err)
			fmt.Println(got)
			require.Equal(t, tt.want, got)
		})
	}
}

type MockCodesignIdentityProvider struct {
	codesignIdentities []certificateutil.CertificateInfoModel
}

func (p MockCodesignIdentityProvider) ListCodesignIdentities() ([]certificateutil.CertificateInfoModel, error) {
	return p.codesignIdentities, nil
}

type MockProvisioningProfileProvider struct {
	profileInfos []profileutil.ProvisioningProfileInfoModel
}

func (p MockProvisioningProfileProvider) ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error) {
	return p.profileInfos, nil
}

func (p MockProvisioningProfileProvider) GetDefaultProvisioningProfile() (profileutil.ProvisioningProfileInfoModel, error) {
	return profileutil.ProvisioningProfileInfoModel{}, nil
}
