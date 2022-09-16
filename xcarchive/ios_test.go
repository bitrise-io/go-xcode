package xcarchive

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	v1xcarchive "github.com/bitrise-io/go-xcode/xcarchive"
)

func TestIosArchive_ReadCodesignParameters(t *testing.T) {
	tests := []struct {
		name    string
		archive IosArchive
		want    *autocodesign.AppLayout
		wantErr bool
	}{
		{
			name: "Single target app",
			archive: IosArchive{
				IosArchive: v1xcarchive.IosArchive{
					Application: v1xcarchive.IosApplication{
						IosBaseApplication: v1xcarchive.IosBaseApplication{
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
			},
			want: &autocodesign.AppLayout{
				Platform: autocodesign.IOS,
				EntitlementsByArchivableTargetBundleID: map[string]autocodesign.Entitlements{
					"io.bitrise.app": nil,
				},
			},
		},
		{
			name: "Multi target app",
			archive: IosArchive{
				IosArchive: v1xcarchive.IosArchive{
					Application: v1xcarchive.IosApplication{
						IosBaseApplication: v1xcarchive.IosBaseApplication{
							InfoPlist: map[string]interface{}{
								"CFBundleIdentifier": "io.bitrise.app",
								"DTPlatformName":     "iphoneos",
							},
							ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
								TeamID: "1234ASDF",
							},
						},
						WatchApplication: &v1xcarchive.IosWatchApplication{
							IosBaseApplication: v1xcarchive.IosBaseApplication{
								InfoPlist: map[string]interface{}{
									"CFBundleIdentifier": "io.bitrise.watchapp",
									"DTPlatformName":     "watchos",
								},
								ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
									TeamID: "1234ASDF",
								},
							},
							Extensions: []v1xcarchive.IosExtension{
								{
									IosBaseApplication: v1xcarchive.IosBaseApplication{
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
						ClipApplication: &v1xcarchive.IosClipApplication{
							IosBaseApplication: v1xcarchive.IosBaseApplication{
								InfoPlist: map[string]interface{}{
									"CFBundleIdentifier": "io.bitrise.clip",
									"DTPlatformName":     "iphoneos",
								},
								ProvisioningProfile: profileutil.ProvisioningProfileInfoModel{
									TeamID: "1234ASDF",
								},
							},
						},
						Extensions: []v1xcarchive.IosExtension{
							{
								IosBaseApplication: v1xcarchive.IosBaseApplication{
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
								IosBaseApplication: v1xcarchive.IosBaseApplication{
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
			},
			want: &autocodesign.AppLayout{
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
				IosArchive: v1xcarchive.IosArchive{
					Application: v1xcarchive.IosApplication{
						IosBaseApplication: v1xcarchive.IosBaseApplication{
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
			},
			want: &autocodesign.AppLayout{
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
			got, err := tt.archive.ReadCodesignParameters()
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
