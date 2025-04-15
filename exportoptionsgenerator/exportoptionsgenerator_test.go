package exportoptionsgenerator

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/exportoptions"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/exportoptionsgenerator/mocks"
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
		bundleID = "io.bundle.id"
		teamID   = "TEAM123"
	)

	logger := log.NewLogger()
	logger.EnableDebugLog(true)

	tests := []struct {
		name                          string
		generatorFactory              func() ExportOptionsGenerator
		exportMethod                  exportoptions.Method
		containerEnvironment          string
		xcodeVersion                  int64
		testFlightInternalTestingOnly bool
		want                          string
		wantErr                       bool
	}{
		{
			name:         "Default development exportOptions",
			exportMethod: exportoptions.MethodDevelopment,
			generatorFactory: func() ExportOptionsGenerator {
				targetInfoProvider := MockTargetInfoProvider{
					mainBundleID: bundleID,
				}
				g := NewWithTargetInfoProvider(targetInfoProvider, newXcodeVersionReader(t, 15), logger)

				return g
			},
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
		/*{
					name:         "Default app store exportOptions",
					exportMethod: exportoptions.MethodAppStore,
					generatorFactory: func() ExportOptionsGenerator {
						applicationTarget := givenApplicationTarget(nil)
						xcodeProj := givenXcodeproj([]xcodeproj.Target{applicationTarget})
						scheme := givenScheme(applicationTarget)
						xcodeVersionReader := newXcodeVersionReader(t, 15)

						g := New(&xcodeProj, &scheme, "", xcodeVersionReader, logger)
						g.targetInfoProvider = MockTargetInfoProvider{
							bundleID: map[string]string{"Application": bundleID},
						}

						return g
					},
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
			</dict>
		</plist>`,
				},
				{
					name:                 "When the app uses iCloud services",
					exportMethod:         exportoptions.MethodDevelopment,
					containerEnvironment: string(exportoptions.ICloudContainerEnvironmentProduction),
					generatorFactory: func() ExportOptionsGenerator {
						applicationTarget := givenApplicationTarget(nil)
						xcodeProj := givenXcodeproj([]xcodeproj.Target{applicationTarget})
						scheme := givenScheme(applicationTarget)
						xcodeVersionReader := newXcodeVersionReader(t, 15)

						g := New(&xcodeProj, &scheme, "", xcodeVersionReader, logger)
						g.targetInfoProvider = MockTargetInfoProvider{
							bundleID:             map[string]string{"Application": bundleID},
							codesignEntitlements: map[string]serialized.Object{"Application": map[string]interface{}{"com.apple.developer.icloud-services": []string{"CloudKit"}}},
						}

						return g
					},
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
					name:                          "When exporting for TestFlight internal testing only",
					exportMethod:                  exportoptions.MethodAppStore,
					testFlightInternalTestingOnly: true,
					generatorFactory: func() ExportOptionsGenerator {
						applicationTarget := givenApplicationTarget(nil)
						xcodeProj := givenXcodeproj([]xcodeproj.Target{applicationTarget})
						scheme := givenScheme(applicationTarget)
						xcodeVersionReader := newXcodeVersionReader(t, 15)

						g := New(&xcodeProj, &scheme, "", xcodeVersionReader, logger)
						g.targetInfoProvider = MockTargetInfoProvider{
							bundleID: map[string]string{"Application": bundleID},
						}

						return g
					},
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
				},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			gotOpts, err := tt.generatorFactory().GenerateApplicationExportOptions(tt.exportMethod, tt.containerEnvironment, teamID, true, true, false, exportoptions.SigningStyleAutomatic, tt.testFlightInternalTestingOnly)

			// Assert
			require.NoError(t, err)

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
			xcodeVersionReader := newXcodeVersionReader(t, tt.xcodeVersion)

			cloudKitEntitlement := map[string]interface{}{"com.apple.developer.icloud-services": []string{"CloudKit"}}
			targetInfoProvider := MockTargetInfoProvider{
				mainBundleID: bundleID,
				bundleIDtoEntitlements: map[string]plistutil.PlistData{
					bundleID:     cloudKitEntitlement,
					bundleIDClip: nil,
				},
			}

			g := NewWithTargetInfoProvider(targetInfoProvider, xcodeVersionReader, logger)
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

			// Act
			gotOpts, err := g.GenerateApplicationExportOptions(tt.exportMethod, "Production", teamID, true, true, false, exportoptions.SigningStyleManual, true)

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
			targetInfoProvider := MockTargetInfoProvider{
				mainBundleID: bundleID,
				bundleIDtoEntitlements: map[string]plistutil.PlistData{
					bundleID:     cloudKitEntitlement,
					bundleIDClip: nil,
				},
			}
			g := NewWithTargetInfoProvider(targetInfoProvider, xcodeVersionReader, logger)

			g.certificateProvider = MockCodesignIdentityProvider{
				[]certificateutil.CertificateInfoModel{certificate},
			}
			g.profileProvider = MockProvisioningProfileProvider{}

			// Act
			gotOpts, err := g.GenerateApplicationExportOptions(tt.exportMethod, "Production", teamID, true, true, false, exportoptions.SigningStyleManual, true)

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

type MockTargetInfoProvider struct {
	mainBundleID           string
	bundleIDtoEntitlements map[string]plistutil.PlistData
}

func (b MockTargetInfoProvider) applicationTargetsAndEntitlements(exportMethod exportoptions.Method) (string, map[string]plistutil.PlistData, error) {
	return b.mainBundleID, b.bundleIDtoEntitlements, nil
}

// func givenAppClipTarget() xcodeproj.Target {
// 	return xcodeproj.Target{
// 		ID:               "app_clip_id",
// 		Name:             "App Clip",
// 		ProductReference: xcodeproj.ProductReference{Path: "Fruta iOS Clip.app"},
// 		ProductType:      AppClipProductType,
// 	}
// }

// func givenApplicationTarget(dependentTargets []xcodeproj.Target) xcodeproj.Target {
// 	var dependencies []xcodeproj.TargetDependency
// 	for _, target := range dependentTargets {
// 		dependencies = append(dependencies, xcodeproj.TargetDependency{TargetID: target.ID})
// 	}

// 	return xcodeproj.Target{
// 		ID:               "application_id",
// 		Name:             "Application",
// 		Dependencies:     dependencies,
// 		ProductReference: xcodeproj.ProductReference{Path: "Fruta.app"},
// 	}
// }

// func givenXcodeproj(targets []xcodeproj.Target) xcodeproj.XcodeProj {
// 	return xcodeproj.XcodeProj{
// 		Proj: xcodeproj.Proj{
// 			Targets: targets,
// 		},
// 	}
// }

// func givenScheme(archivableTarget xcodeproj.Target) xcscheme.Scheme {
// 	return xcscheme.Scheme{
// 		BuildAction: xcscheme.BuildAction{
// 			BuildActionEntries: []xcscheme.BuildActionEntry{
// 				{
// 					BuildForArchiving: "YES",
// 					BuildableReference: xcscheme.BuildableReference{
// 						BuildableName:       archivableTarget.ProductReference.Path,
// 						BlueprintIdentifier: archivableTarget.ID,
// 					},
// 				},
// 			},
// 		},
// 	}
// }
