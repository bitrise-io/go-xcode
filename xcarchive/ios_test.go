package xcarchive

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/stretchr/testify/require"
)

var tmpDir = ""

func sampleRepoPath(t *testing.T) string {
	dir := ""
	if tmpDir != "" {
		dir = tmpDir
	} else {
		var err error
		dir, err = pathutil.NewPathProvider().CreateTempDir(tempDirName)
		require.NoError(t, err)
		sampleArtifactsGitURI := "https://github.com/bitrise-io/sample-artifacts.git"

		cmd := command.NewFactory(env.NewRepository()).Create("git", []string{"clone", sampleArtifactsGitURI, dir}, nil)
		output, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			t.Log(output)
			t.Errorf("git clone failed: %s", err)
		}
		tmpDir = dir
	}
	t.Logf("sample artifcats dir: %s\n", dir)
	return dir
}

func TestNewIosArchive(t *testing.T) {
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)
	require.Equal(t, 5, len(archive.InfoPlist))

	app := archive.Application
	require.Equal(t, 26, len(app.InfoPlist))
	require.Equal(t, 4, len(app.Entitlements))
	require.Equal(t, "*", app.ProvisioningProfile.BundleID)

	require.Equal(t, 1, len(app.Extensions))
	extension := app.Extensions[0]
	require.Equal(t, 23, len(extension.InfoPlist))
	require.Equal(t, 4, len(extension.Entitlements))
	require.Equal(t, "*", extension.ProvisioningProfile.BundleID)

	require.NotNil(t, app.WatchApplication)
	watchApp := *app.WatchApplication
	require.Equal(t, 24, len(watchApp.InfoPlist))
	require.Equal(t, 4, len(watchApp.Entitlements))
	require.Equal(t, "*", watchApp.ProvisioningProfile.BundleID)

	require.Equal(t, 1, len(watchApp.Extensions))
	watchExtension := watchApp.Extensions[0]
	require.Equal(t, 23, len(watchExtension.InfoPlist))
	require.Equal(t, 4, len(watchExtension.Entitlements))
	require.Equal(t, "*", watchExtension.ProvisioningProfile.BundleID)
}

func TestNewAppClipArchive(t *testing.T) {
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/Fruta.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)
	require.Equal(t, 5, len(archive.InfoPlist))

	app := archive.Application
	require.Equal(t, 30, len(app.InfoPlist))
	require.Equal(t, 6, len(app.Entitlements))
	require.Equal(t, "io.bitrise.appcliptest", app.ProvisioningProfile.BundleID)

	require.Equal(t, 1, len(app.Extensions))
	extension := app.Extensions[0]
	require.Equal(t, 24, len(extension.InfoPlist))
	require.Equal(t, 4, len(extension.Entitlements))
	require.Equal(t, "io.bitrise.appcliptest.ios-widgets", extension.ProvisioningProfile.BundleID)

	require.NotNil(t, app.ClipApplication)
	clipApp := *app.ClipApplication
	require.Equal(t, 31, len(clipApp.InfoPlist))
	require.Equal(t, 8, len(clipApp.Entitlements))
	require.Equal(t, "io.bitrise.appcliptest.Clip", clipApp.ProvisioningProfile.BundleID)
}

func TestIsXcodeManaged(t *testing.T) {
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)

	require.Equal(t, false, archive.IsXcodeManaged())
}

func TestSigningIdentity(t *testing.T) {
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)

	require.Equal(t, "iPhone Developer: Bitrise Bot (VV2J4SV8V4)", archive.SigningIdentity())
}

func TestBundleIDEntitlementsMap(t *testing.T) {
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)

	bundleIDEntitlementsMap := archive.BundleIDEntitlementsMap()
	require.Equal(t, 4, len(bundleIDEntitlementsMap))

	bundleIDs := []string{"com.bitrise.code-sign-test.share-extension", "com.bitrise.code-sign-test.watchkitapp", "com.bitrise.code-sign-test.watchkitapp.watchkitextension", "com.bitrise.code-sign-test"}
	for _, bundleID := range bundleIDs {
		_, ok := bundleIDEntitlementsMap[bundleID]
		require.True(t, ok, fmt.Sprintf("%v", bundleIDEntitlementsMap))
	}
}

func TestBundleIDProfileInfoMap(t *testing.T) {
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)

	bundleIDProfileInfoMap := archive.BundleIDProfileInfoMap()
	require.Equal(t, 4, len(bundleIDProfileInfoMap))

	bundleIDs := []string{"com.bitrise.code-sign-test.share-extension", "com.bitrise.code-sign-test.watchkitapp", "com.bitrise.code-sign-test.watchkitapp.watchkitextension", "com.bitrise.code-sign-test"}
	for _, bundleID := range bundleIDs {
		_, ok := bundleIDProfileInfoMap[bundleID]
		require.True(t, ok, fmt.Sprintf("%v", bundleIDProfileInfoMap))
	}
}

func TestFindDSYMs(t *testing.T) {
	// base case: dsyms for apps and frameworks
	iosArchivePth := filepath.Join(sampleRepoPath(t), "archives/Fruta.xcarchive")
	archive, err := NewIosArchive(iosArchivePth)
	require.NoError(t, err)

	appDsym, otherDsyms, err := archive.FindDSYMs()
	require.NoError(t, err)
	require.Equal(t, 2, len(appDsym))
	require.Equal(t, 2, len(otherDsyms))

	// no app dsym case: something has changed since the
	// initial implementation of the function under test,
	// and is causing dsyms with filenames to be generated
	// even when dsym generation is turned off -- we don't care about
	// other dsyms in this case, only whether the app dsym
	// path is empty
	noDSYMArchivePth := filepath.Join(sampleRepoPath(t), "archives/ios.ios-simple-objc.noappdsym.xcarchive")
	archive, err = NewIosArchive(noDSYMArchivePth)
	require.NoError(t, err)

	appDsym, _, err = archive.FindDSYMs()
	require.NoError(t, err)
	require.Empty(t, appDsym)
}

func Test_applicationFromArchive(t *testing.T) {
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "{}GlobControlChars:a-b[ab]?*", "test.xcarchive")
	appDir := filepath.Join(archivePath, "Products", "Applications")
	appPath := filepath.Join(appDir, "test.app")
	t.Logf("Test app path: %s", appPath)
	err := os.MkdirAll(appDir, os.ModePerm)
	if err != nil {
		t.Errorf("setup: failed to create directory: %s, error: %s", appDir, err)
	}
	file, err := os.Create(appPath)
	if err != nil {
		t.Errorf("setup: failed to create test archive: %s, error: %s", appPath, err)
	}
	if err := file.Close(); err != nil {
		t.Errorf("setup: failed to close file, error: %s", err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "[] glob control characters in path",
			args: args{
				path: archivePath,
			},
			want:    appPath,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applicationFromArchive(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("applicationFromArchive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applicationFromArchive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applicationFromPlist(t *testing.T) {
	infoPlist, err := plistutil.NewPlistDataFromFile(filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive/Info.plist"))
	const appRelativePathToProduct = "Applications/code-sign-test.app"
	if err != nil {
		t.Errorf("setup: could not read plist, error: %s", infoPlist)
	}

	type args struct {
		InfoPlist plistutil.PlistData
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "normal case",
			args: args{
				infoPlist,
			},
			want:  appRelativePathToProduct,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := applicationFromPlist(tt.args.InfoPlist)
			if got != tt.want {
				t.Errorf("applicationFromPlist() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("applicationFromPlist() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestIosArchive_GetAppLayout(t *testing.T) {
	tests := []struct {
		name    string
		archive IosArchive
		want    autocodesign.AppLayout
		wantErr bool
	}{
		{
			name: "Single target app",
			archive: IosArchive{
				Application: IosApplication{
					IosBaseApplication: IosBaseApplication{
						InfoPlist: map[string]interface{}{
							"CFBundleIdentifier": "io.bitrise.app",
							"DTPlatformName":     "iphoneos",
						},
						ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
							TeamID: "1234ASDF",
						},
					},
				},
			},
			want: autocodesign.AppLayout{
				Platform: autocodesign.IOS,
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					"io.bitrise.app": nil,
				},
			},
		},
		{
			name: "Multi target app",
			archive: IosArchive{
				Application: IosApplication{
					IosBaseApplication: IosBaseApplication{
						InfoPlist: map[string]interface{}{
							"CFBundleIdentifier": "io.bitrise.app",
							"DTPlatformName":     "iphoneos",
						},
						ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
							TeamID: "1234ASDF",
						},
					},
					WatchApplication: &IosWatchApplication{
						IosBaseApplication: IosBaseApplication{
							InfoPlist: map[string]interface{}{
								"CFBundleIdentifier": "io.bitrise.watchapp",
								"DTPlatformName":     "watchos",
							},
							ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
								TeamID: "1234ASDF",
							},
						},
						Extensions: []IosExtension{
							{
								IosBaseApplication: IosBaseApplication{
									InfoPlist: map[string]interface{}{
										"CFBundleIdentifier": "io.bitrise.watch-widget",
										"DTPlatformName":     "watchos",
									},
									ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
										TeamID: "1234ASDF",
									},
								},
							},
						},
					},
					ClipApplication: &IosClipApplication{
						IosBaseApplication: IosBaseApplication{
							InfoPlist: map[string]interface{}{
								"CFBundleIdentifier": "io.bitrise.clip",
								"DTPlatformName":     "iphoneos",
							},
							ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
								TeamID: "1234ASDF",
							},
						},
					},
					Extensions: []IosExtension{
						{
							IosBaseApplication: IosBaseApplication{
								InfoPlist: map[string]interface{}{
									"CFBundleIdentifier": "io.bitrise.ios-widget1",
									"DTPlatformName":     "iphoneos",
								},
								ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
									TeamID: "1234ASDF",
								},
							},
						},
						{
							IosBaseApplication: IosBaseApplication{
								InfoPlist: map[string]interface{}{
									"CFBundleIdentifier": "io.bitrise.ios-widget2",
									"DTPlatformName":     "iphoneos",
								},
								ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
									TeamID: "1234ASDF",
								},
							},
						},
					},
				},
			},
			want: autocodesign.AppLayout{
				Platform: autocodesign.IOS,
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					"io.bitrise.app":          nil,
					"io.bitrise.watchapp":     nil,
					"io.bitrise.watch-widget": nil,
					"io.bitrise.clip":         nil,
					"io.bitrise.ios-widget1":  nil,
					"io.bitrise.ios-widget2":  nil,
				},
			},
		},
		{
			name: "Single target app with capabilities",
			archive: IosArchive{
				Application: IosApplication{
					IosBaseApplication: IosBaseApplication{
						InfoPlist: map[string]interface{}{
							"CFBundleIdentifier": "io.bitrise.app",
							"DTPlatformName":     "iphoneos",
						},
						ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
							TeamID: "1234ASDF",
						},
						Entitlements: map[string]interface{}{
							"get-task-allow":                        false,
							"com.apple.security.application-groups": []string{"group.io.bitrise.app"},
						},
					},
				},
			},
			want: autocodesign.AppLayout{
				Platform: autocodesign.IOS,
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					"io.bitrise.app": {
						"get-task-allow":                        false,
						"com.apple.security.application-groups": []string{"group.io.bitrise.app"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.archive.GetAppLayout(false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCodesignParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadCodesignParameters() got = %v, want %v", got, tt.want)
			}
		})
	}
}
