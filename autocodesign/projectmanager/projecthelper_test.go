package projectmanager

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcodeproj"
	"github.com/stretchr/testify/require"
)

var schemeCases []string
var targetCases []string
var xcProjCases []xcodeproj.XcodeProj
var projectCases []string
var projHelpCases []ProjectHelper
var configCases []string

func TestMain(m *testing.M) {
	var err error
	schemeCases, _, xcProjCases, projHelpCases, configCases, projectCases, err = initTestCases()
	if err != nil {
		fmt.Printf("Failed to initialize test cases: %s\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func Test_GivenXcode13_WhenProjectHelperInitialised_ThenUITestTargetsConsidered(t *testing.T) {
	// ios_project_files/Xcode-10_default.xcworkspace
	project := projectCases[0]
	scheme := "Xcode-10_default"
	configuration := "Debug"

	projectHelper, err := NewProjectHelper(project, scheme, configuration, 13)
	require.NoError(t, err)

	require.Equal(t, 1, len(projectHelper.TestTargets))

	testTarget := projectHelper.TestTargets[0]
	require.Equal(t, "Xcode-10_defaultUITests", testTarget.Name)
}

func Test_GivenXcode14_WhenProjectHelperInitialised_ThenUnitAndUITestTargetsConsidered(t *testing.T) {
	// ios_project_files/Xcode-10_default.xcworkspace
	project := projectCases[0]
	scheme := "Xcode-10_default"
	configuration := "Debug"

	projectHelper, err := NewProjectHelper(project, scheme, configuration, 14)
	require.NoError(t, err)

	require.Equal(t, 2, len(projectHelper.TestTargets))

	testTarget := projectHelper.TestTargets[0]
	require.Equal(t, "Xcode-10_defaultTests", testTarget.Name)

	testTarget = projectHelper.TestTargets[1]
	require.Equal(t, "Xcode-10_defaultUITests", testTarget.Name)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name              string
		projOrWSPath      string
		schemeName        string
		configurationName string
		wantConfiguration string
		wantErr           bool
	}{
		{
			name:              "Xcode 10 workspace - iOS",
			projOrWSPath:      projectCases[0],
			schemeName:        "Xcode-10_default",
			configurationName: "Debug",
			wantConfiguration: "Debug",
			wantErr:           false,
		},
		{
			name:              "Xcode 10 workspace - iOS - Default configuration",
			projOrWSPath:      projectCases[0],
			schemeName:        "Xcode-10_default",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "Xcode 10 workspace - iOS - Scheme in workspace",
			projOrWSPath:      projectCases[6],
			schemeName:        "Xcode-10_default",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "Xcode 10 workspace - iOS - Default configuration - Gdańsk scheme",
			projOrWSPath:      projectCases[0],
			schemeName:        "Gdańsk",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "Xcode-10_mac project - MacOS - Debug configuration",
			projOrWSPath:      projectCases[2],
			schemeName:        "Xcode-10_mac",
			configurationName: "Debug",
			wantConfiguration: "Debug",
			wantErr:           false,
		},
		{
			name:              "Xcode-10_mac project - MacOS - Default configuration",
			projOrWSPath:      projectCases[2],
			schemeName:        "Xcode-10_mac",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "TV_OS.xcodeproj project - TVOS - Default configuration",
			projOrWSPath:      projectCases[4],
			schemeName:        "TV_OS",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projHelp, err := NewProjectHelper(tt.projOrWSPath, tt.schemeName, tt.configurationName, 13)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if projHelp == nil {
				t.Errorf("New() error = No projectHelper was generated")
			}
			if projHelp.Configuration != tt.wantConfiguration {
				t.Errorf("New() got1 = %v, want %v", projHelp.Configuration, tt.wantConfiguration)
			}
		})
	}
}

func TestProjectHelper_ProjectTeamID_withoutTargetAttributes(t *testing.T) {
	log.SetEnableDebugLog(true)

	helper := ProjectHelper{
		MainTarget: xcodeproj.Target{Name: "AppTarget"},
		// bypass calling xcodebuild -showBuildSettings
		buildSettingsCache: map[string]map[string]serialized.Object{"AppTarget": {"Debug": {}}},
		// project withouth TargetAttributes
		XcProj: xcodeproj.XcodeProj{Proj: xcodeproj.Proj{Attributes: xcodeproj.ProjectAtributes{TargetAttributes: nil}}},
	}
	_, err := helper.ProjectTeamID("Debug")
	require.NoError(t, err)
}

func TestProjectHelper_ProjectTeamID(t *testing.T) {
	log.SetEnableDebugLog(true)

	tests := []struct {
		name    string
		config  string
		want    string
		wantErr bool
	}{
		{
			name:    schemeCases[0] + " Debug",
			config:  configCases[0],
			want:    "72SA8V3WYL",
			wantErr: false,
		},
		{
			name:    schemeCases[1] + " Release",
			config:  configCases[1],
			want:    "72SA8V3WYL",
			wantErr: false,
		},
		{
			name:    schemeCases[2] + " Debug",
			config:  configCases[2],
			want:    "72SA8V3WYL",
			wantErr: false,
		},
		{
			name:    schemeCases[3] + " Release",
			config:  configCases[3],
			want:    "72SA8V3WYL",
			wantErr: false,
		},
		{
			name:    schemeCases[4] + " Debug",
			config:  configCases[4],
			want:    "72SA8V3WYL",
			wantErr: false,
		},
		{
			name:    schemeCases[5] + " Release",
			config:  configCases[5],
			want:    "72SA8V3WYL",
			wantErr: false,
		},
	}

	for i, tt := range tests {
		p := projHelpCases[i]

		t.Run(tt.name, func(t *testing.T) {
			got, err := p.ProjectTeamID(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectHelper.ProjectTeamID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProjectHelper.ProjectTeamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_expandTargetSetting(t *testing.T) {
	const productName = "Sample"
	tests := []struct {
		name          string
		value         string
		buildSettings map[string]interface{}
		want          string
		wantErr       bool
	}{
		{
			name:  "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
			value: "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["PRODUCT_NAME"] = productName
				return m
			}(),
			want:    fmt.Sprintf("Bitrise.%s", productName),
			wantErr: false,
		},
		{
			name:  "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
			value: "Bitrise.$(PRODUCT_NAME:rfc1034identifier)",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["PRODUCT_NAME"] = productName
				m["a"] = productName
				return m
			}(),
			want:    fmt.Sprintf("Bitrise.%s", productName),
			wantErr: false,
		},
		{
			name:  "Bitrise.Test.$(PRODUCT_NAME:rfc1034identifier).Suffix",
			value: "Bitrise.Test.$(PRODUCT_NAME:rfc1034identifier).Suffix",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["PRODUCT_NAME"] = productName
				m["a"] = productName
				return m
			}(),
			want:    fmt.Sprintf("Bitrise.Test.%s.Suffix", productName),
			wantErr: false,
		},
		{
			name:  "iCloud.$(CFBundleIdentifier)",
			value: "iCloud.$(CFBundleIdentifier)",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["CFBundleIdentifier"] = productName
				return m
			}(),
			want:    fmt.Sprintf("iCloud.%s", productName),
			wantErr: false,
		},
		{
			name:  "iCloud.${CFBundleIdentifier}",
			value: "iCloud.${CFBundleIdentifier}",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["CFBundleIdentifier"] = productName
				return m
			}(),
			want:    fmt.Sprintf("iCloud.%s", productName),
			wantErr: false,
		},
		{
			name:  "${CFBundleIdentifier}.Suffix",
			value: "${CFBundleIdentifier}.Suffix",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["CFBundleIdentifier"] = productName
				return m
			}(),
			want:    fmt.Sprintf("%s.Suffix", productName),
			wantErr: false,
		},
		{
			name:  "$(CFBundleIdentifier)",
			value: "$(CFBundleIdentifier)",
			buildSettings: func() map[string]interface{} {
				m := make(map[string]interface{})
				m["CFBundleIdentifier"] = productName
				return m
			}(),
			want:    productName,
			wantErr: false,
		},
		{
			name:          "iCloud.bundle.id",
			value:         "iCloud.bundle.id",
			buildSettings: nil,
			want:          "",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandTargetSetting(tt.value, tt.buildSettings)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandTargetSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("expandTargetSetting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectHelper_TargetBundleID(t *testing.T) {
	for i, schemeCase := range schemeCases {
		xcProj, _, err := findBuiltProject(
			projectCases[i],
			schemeCase,
		)
		if err != nil {
			t.Fatalf("Failed to generate XcodeProj for test case: %s", err)
		}
		xcProjCases = append(xcProjCases, xcProj)

		projHelp, err := NewProjectHelper(
			projectCases[i],
			schemeCase,
			configCases[i],
			13,
		)
		if err != nil {
			t.Fatalf("Failed to generate projectHelper for test case: %s", err)
		}
		projHelpCases = append(projHelpCases, *projHelp)
	}

	tests := []struct {
		name       string
		targetName string
		conf       string
		want       string
		wantErr    bool
	}{
		{
			name:       targetCases[0] + " Debug",
			targetName: targetCases[0],
			conf:       configCases[0],
			want:       "com.bitrise.Xcode-10-default",
			wantErr:    false,
		},
		{
			name:       targetCases[1] + " Release",
			targetName: targetCases[1],
			conf:       configCases[1],
			want:       "com.bitrise.Xcode-10-default",
			wantErr:    false,
		},
		{
			name:       targetCases[2] + " Release",
			targetName: targetCases[2],
			conf:       configCases[2],
			want:       "com.bitrise.Xcode-10-mac",
			wantErr:    false,
		},
		{
			name:       targetCases[3] + " Release",
			targetName: targetCases[3],
			conf:       configCases[3],
			want:       "com.bitrise.Xcode-10-mac",
			wantErr:    false,
		},
		{
			name:       targetCases[4] + " Release",
			targetName: targetCases[4],
			conf:       configCases[4],
			want:       "com.bitrise.TV-OS",
			wantErr:    false,
		},
		{
			name:       targetCases[5] + " Release",
			targetName: targetCases[5],
			conf:       configCases[5],
			want:       "com.bitrise.TV-OS",
			wantErr:    false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := projHelpCases[i]

			got, err := p.TargetBundleID(tt.targetName, tt.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectHelper.TargetBundleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProjectHelper.TargetBundleID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectHelper_targetEntitlements(t *testing.T) {
	tests := []struct {
		name          string
		targetName    string
		conf          string
		bundleID      string
		want          autocodesign.Entitlements
		projectHelper ProjectHelper
		wantErr       bool
	}{
		{
			name:          targetCases[2] + " Release",
			targetName:    targetCases[2],
			conf:          configCases[2],
			projectHelper: projHelpCases[2],
			want: func() autocodesign.Entitlements {
				m := make(map[string]interface{})
				m["com.apple.security.app-sandbox"] = true
				m["com.apple.security.files.user-selected.read-only"] = true
				return m
			}(),
			wantErr: false,
		},
		{
			name:          targetCases[3] + " Release",
			targetName:    targetCases[3],
			conf:          configCases[3],
			projectHelper: projHelpCases[3],
			want: func() autocodesign.Entitlements {
				m := make(map[string]interface{})
				m["com.apple.security.app-sandbox"] = true
				m["com.apple.security.files.user-selected.read-only"] = true
				return m
			}(),
			wantErr: false,
		},
		{
			name:          targetCases[4] + " Release",
			targetName:    targetCases[4],
			conf:          configCases[4],
			projectHelper: projHelpCases[4],
			want:          nil,
			wantErr:       false,
		},
		{
			name:          targetCases[5] + " Release",
			targetName:    targetCases[5],
			conf:          configCases[5],
			projectHelper: projHelpCases[5],
			want:          nil,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.projectHelper.targetEntitlements(tt.targetName, tt.conf, tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectHelper.targetEntitlements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProjectHelper.targetEntitlements() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveEntitlementVariables(t *testing.T) {
	type args struct {
		entitlements autocodesign.Entitlements
		bundleID     string
	}
	tests := []struct {
		name    string
		args    args
		want    autocodesign.Entitlements
		wantErr bool
	}{
		{
			name: "Existing entitlememts are unchanged",
			args: args{
				entitlements: map[string]interface{}{
					"com.apple.developer.contacts.notes": true,
				},
			},
			want: map[string]interface{}{
				"com.apple.developer.contacts.notes": true,
			},
		},
		{
			name: "iCloud entitlememts are unchanged, if service is in use",
			args: args{
				entitlements: map[string]interface{}{
					"com.apple.developer.icloud-services": []interface{}{"CloudKit"},
					"com.apple.developer.icloud-container-identifiers": []interface{}{
						"iCloud.bundle.id",
					},
				},
			},
			want: map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{"CloudKit"},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.bundle.id",
				},
			},
		},
		{
			name: "iCloud entitlememts are unchanged, if service is not in use",
			args: args{
				entitlements: map[string]interface{}{
					"com.apple.developer.icloud-services": []interface{}{},
					"com.apple.developer.icloud-container-identifiers": []interface{}{
						"iCloud.bundle.id",
					},
				},
			},
			want: map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.bundle.id",
				},
			},
		},
		{
			name: "iCloud containers CFBundleIdentifier variable is expanded",
			args: args{
				entitlements: map[string]interface{}{
					"com.apple.developer.icloud-services": []interface{}{"CloudKit"},
					"com.apple.developer.icloud-container-identifiers": []interface{}{
						"iCloud.${CFBundleIdentifier}",
					},
				},
				bundleID: "bundle.id",
			},
			want: map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{"CloudKit"},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.bundle.id",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveEntitlementVariables(tt.args.entitlements, tt.args.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveEntitlementVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolveEntitlementVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func initTestCases() ([]string, []string, []xcodeproj.XcodeProj, []ProjectHelper, []string, []string, error) {
	//
	// If the test cases already initialized return them
	if schemeCases != nil {
		return schemeCases, targetCases, xcProjCases, projHelpCases, configCases, projectCases, nil
	}

	p, err := pathutil.NormalizedOSTempDirPath("_autoprov")
	if err != nil {
		log.Errorf("Failed to create tmp dir error: %s", err)
	}

	gitRepo, err := git.New(p)
	if err != nil {
		log.Errorf("failed to init git repo: %w", err)
	}

	branch := "project"
	cmd := gitRepo.CloneTagOrBranch("https://github.com/bitrise-io/sample-artifacts.git", branch)
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to git clone the sample project files error: %s", err)
	}
	//
	// Init test cases
	targetCases = []string{
		"Xcode-10_default",
		"Xcode-10_default",
		"Xcode-10_mac",
		"Xcode-10_mac",
		"TV_OS",
		"TV_OS",
		"Xcode-10-default",
	}

	schemeCases = []string{
		"Xcode-10_default",
		"Xcode-10_default",
		"Xcode-10_mac",
		"Xcode-10_mac",
		"TV_OS",
		"TV_OS",
		"Xcode-10_default",
	}
	configCases = []string{
		"Debug",
		"Release",
		"Debug",
		"Release",
		"Debug",
		"Release",
		"Debug",
	}
	projectCases = []string{
		p + "/ios_project_files/Xcode-10_default.xcworkspace",
		p + "/ios_project_files/Xcode-10_default.xcworkspace",
		p + "/ios_project_files/Xcode-10_mac.xcodeproj",
		p + "/ios_project_files/Xcode-10_mac.xcodeproj",
		p + "/ios_project_files/TV_OS.xcodeproj",
		p + "/ios_project_files/TV_OS.xcodeproj",
		p + "/ios_project_files/Xcode-10_with_scheme.xcworkspace",
	}
	var xcProjCases []xcodeproj.XcodeProj
	var projHelpCases []ProjectHelper

	for i, schemeCase := range schemeCases {
		xcProj, _, err := findBuiltProject(
			projectCases[i],
			schemeCase,
		)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to generate XcodeProj for test case: %s", err)
		}
		xcProjCases = append(xcProjCases, xcProj)

		projHelp, err := NewProjectHelper(
			projectCases[i],
			schemeCase,
			configCases[i],
			13,
		)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("failed to generate projectHelper for test case: %s", err)
		}
		projHelpCases = append(projHelpCases, *projHelp)
	}

	return schemeCases, targetCases, xcProjCases, projHelpCases, configCases, projectCases, nil
}
