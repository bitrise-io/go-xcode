package xcodeproj

import (
	"path/filepath"
	"reflect"
	"testing"

	plist "github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/xcodeproject/pretty"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestParseConfigurationList(t *testing.T) {
	var raw serialized.Object
	_, err := plist.Unmarshal([]byte(rawConfigurationList), &raw)
	require.NoError(t, err)

	configurationList, err := parseConfigurationList("13E76E3A1F4AC90A0028096E", raw)
	require.NoError(t, err)
	// fmt.Printf("configurationList:\n%s\n", pretty.Object(configurationList))
	require.Equal(t, expectedConfigurationList, pretty.Object(configurationList))
}

const rawConfigurationList = `
{
	13E76E3A1F4AC90A0028096E /* Build configuration list for PBXNativeTarget "code-sign-test" */ = {
		isa = XCConfigurationList;
		buildConfigurations = (
			13E76E3B1F4AC90A0028096E /* Debug */,
			13E76E3C1F4AC90A0028096E /* Release */,
		);
		defaultConfigurationIsVisible = 0;
		defaultConfigurationName = Release;
	};

	13E76E3B1F4AC90A0028096E /* Debug */ = {
		isa = XCBuildConfiguration;
		buildSettings = {
			ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
			"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
			CODE_SIGN_STYLE = Automatic;
			DEVELOPMENT_TEAM = 72SA8V3WYL;
			INFOPLIST_FILE = "code-sign-test/Info.plist";
			LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
			PRODUCT_BUNDLE_IDENTIFIER = "com.bitrise.code-sign-test";
			PRODUCT_NAME = "$(TARGET_NAME)";
			PROVISIONING_PROFILE = "";
			PROVISIONING_PROFILE_SPECIFIER = "";
			TARGETED_DEVICE_FAMILY = "1,2";
		};
		name = Debug;
	};

	13E76E3C1F4AC90A0028096E /* Release */ = {
		isa = XCBuildConfiguration;
		buildSettings = {
			ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
			"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "iPhone Developer";
			CODE_SIGN_STYLE = Automatic;
			DEVELOPMENT_TEAM = 72SA8V3WYL;
			INFOPLIST_FILE = "code-sign-test/Info.plist";
			LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
			PRODUCT_BUNDLE_IDENTIFIER = "com.bitrise.code-sign-test";
			PRODUCT_NAME = "$(TARGET_NAME)";
			PROVISIONING_PROFILE = "";
			PROVISIONING_PROFILE_SPECIFIER = "";
			TARGETED_DEVICE_FAMILY = "1,2";
		};
		name = Release;
	};
}`

const expectedConfigurationList = `{
	"ID": "13E76E3A1F4AC90A0028096E",
	"DefaultConfigurationName": "Release",
	"BuildConfigurations": [
		{
			"ID": "13E76E3B1F4AC90A0028096E",
			"Name": "Debug",
			"BuildSettings": {
				"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
				"CODE_SIGN_STYLE": "Automatic",
				"DEVELOPMENT_TEAM": "72SA8V3WYL",
				"INFOPLIST_FILE": "code-sign-test/Info.plist",
				"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks",
				"PRODUCT_BUNDLE_IDENTIFIER": "com.bitrise.code-sign-test",
				"PRODUCT_NAME": "$(TARGET_NAME)",
				"PROVISIONING_PROFILE": "",
				"PROVISIONING_PROFILE_SPECIFIER": "",
				"TARGETED_DEVICE_FAMILY": "1,2"
			}
		},
		{
			"ID": "13E76E3C1F4AC90A0028096E",
			"Name": "Release",
			"BuildSettings": {
				"ASSETCATALOG_COMPILER_APPICON_NAME": "AppIcon",
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
				"CODE_SIGN_STYLE": "Automatic",
				"DEVELOPMENT_TEAM": "72SA8V3WYL",
				"INFOPLIST_FILE": "code-sign-test/Info.plist",
				"LD_RUNPATH_SEARCH_PATHS": "$(inherited) @executable_path/Frameworks",
				"PRODUCT_BUNDLE_IDENTIFIER": "com.bitrise.code-sign-test",
				"PRODUCT_NAME": "$(TARGET_NAME)",
				"PROVISIONING_PROFILE": "",
				"PROVISIONING_PROFILE_SPECIFIER": "",
				"TARGETED_DEVICE_FAMILY": "1,2"
			}
		}
	]
}`

func TestXcodeProjBuildConfigurationList(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	if err != nil {
		t.Fatalf("Failed to init project for test case, error: %s", err)
	}
	tests := []struct {
		name     string
		targetID string
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "Fetch xcode-project-test sample's buildConfiguration list",
			targetID: "7D5B35FB20E28EE80022BAE6",
			want: map[string]interface{}{
				"buildConfigurations": []string{
					"7D5B360F20E28EEA0022BAE6",
					"7D5B361020E28EEA0022BAE6",
				},
				"defaultConfigurationIsVisible": "0",
				"defaultConfigurationName":      "Release",
				"isa":                           "XCConfigurationList",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := project.BuildConfigurationList(tt.targetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("XcodeProj.BuildConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			slice, err := got.StringSlice("buildConfigurations")
			if err != nil {
				t.Errorf("failed to get buildConfigurations string slice for test case, error: %s", err)
			}
			if !reflect.DeepEqual(slice, tt.want["buildConfigurations"]) {
				t.Errorf("XcodeProj.BuildConfigurations() = %s, want %s", pretty.Object(got), pretty.Object(tt.want))
			}
		})
	}
}

func TestXcodeProjBuildConfigurations(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	if err != nil {
		t.Fatalf("Failed to init project for test case, error: %s", err)
	}

	buildConfigurationList, err := project.BuildConfigurationList("7D5B35FB20E28EE80022BAE6")
	if err != nil {
		t.Fatalf("Failed to init buildConfigurationList for test case, error: %s", err)
	}

	tests := []struct {
		name                   string
		buildConfigurationList serialized.Object
		want                   []serialized.Object
		wantErr                bool
	}{
		{
			name:                   "Fetch xcode-project-test sample's buildConfigurations",
			buildConfigurationList: buildConfigurationList,
			want: []serialized.Object{
				serialized.Object{
					"buildSettings": map[string]interface{}{
						"ALWAYS_EMBED_SWIFT_STANDARD_LIBRARIES": "YES",
						"ASSETCATALOG_COMPILER_APPICON_NAME":    "AppIcon",
						"CODE_SIGN_STYLE":                       "Automatic",
						"DEVELOPMENT_TEAM":                      "72SA8V3WYL",
						"INFOPLIST_FILE":                        "XcodeProj/Info.plist",
						"LD_RUNPATH_SEARCH_PATHS": []interface{}{
							"$(inherited)",
							"@executable_path/Frameworks",
						},
						"PRODUCT_BUNDLE_IDENTIFIER": "com.bitrise.XcodeProj",
						"PRODUCT_NAME":              "$(TARGET_NAME)",
						"SWIFT_VERSION":             "4.0",
						"TARGETED_DEVICE_FAMILY":    "1,2",
					},
					"isa":  "XCBuildConfiguration",
					"name": "Debug",
				},
				serialized.Object{
					"buildSettings": map[string]interface{}{
						"ALWAYS_EMBED_SWIFT_STANDARD_LIBRARIES": "YES",
						"ASSETCATALOG_COMPILER_APPICON_NAME":    "AppIcon",
						"CODE_SIGN_STYLE":                       "Automatic",
						"DEVELOPMENT_TEAM":                      "72SA8V3WYL",
						"INFOPLIST_FILE":                        "XcodeProj/Info.plist",
						"LD_RUNPATH_SEARCH_PATHS": []interface{}{
							"$(inherited)",
							"@executable_path/Frameworks",
						},
						"PRODUCT_BUNDLE_IDENTIFIER": "com.bitrise.XcodeProj",
						"PRODUCT_NAME":              "$(TARGET_NAME)",
						"SWIFT_VERSION":             "4.0",
						"TARGETED_DEVICE_FAMILY":    "1,2",
					},
					"isa":  "XCBuildConfiguration",
					"name": "Release",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := project.BuildConfigurations(tt.buildConfigurationList)
			if (err != nil) != tt.wantErr {
				t.Errorf("XcodeProj.BuildConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("XcodeProj.BuildConfigurations() = %v, want %v", pretty.Object(got), pretty.Object(tt.want))
				t.Logf("Diff %s", cmp.Diff(got, tt.want))
			}
		})
	}
}
