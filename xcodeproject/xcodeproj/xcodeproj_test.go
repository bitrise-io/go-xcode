package xcodeproj

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/norm"
)

func TestResolve(t *testing.T) {

	t.Log("resolves bundle id in format: prefix.${ENV_KEY}.$ENV_KEY_2")
	{
		bundleID := `prefix.${PRODUCT_NAME}.$VERSION`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
			"VERSION":      "beta",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "prefix.ios-simple-objc.beta", resolved)
	}

	t.Log("resolves bundle id in format: prefix.{text.${ENV_KEY}.text}")
	{
		bundleID := `prefix.{text.${PRODUCT_NAME}.text}`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "prefix.{text.ios-simple-objc.text}", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEY")
	{
		bundleID := `auto_provision.$PRODUCT_NAME`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEYsuffix")
	{
		bundleID := `auto_provision.$PRODUCT_NAMEsuffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objcsuffix", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEYsuffix$ENV_KEY")
	{
		bundleID := `auto_provision.$PRODUCT_NAMEsuffix$PRODUCT_NAME`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objcsuffixios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEY$ENV_KEY_2")
	{
		bundleID := `auto_provision.$PRODUCT_NAME$VERSION`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
			"VERSION":      "beta",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objcbeta", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier)")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEYtest.suffix")
	{
		bundleID := `auto_provision.$PRODUCT_NAMEtest.suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objctest.suffix", resolved)
	}

	t.Log("resolves bundle id with cross env reference")
	{
		bundleID := `auto_provision.$(BUNDLE_ID:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
			"BUNDLE_ID":    "$(PRODUCT_NAME:rfc1034identifier)",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("detects env refernce cycle")
	{
		bundleID := `auto_provision.$(BUNDLE_ID:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "$(BUNDLE_ID:rfc1034identifier)",
			"BUNDLE_ID":    "$(PRODUCT_NAME:rfc1034identifier)",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.EqualError(t, err, "bundle id reference cycle found")
		require.Equal(t, "", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix", resolved)
	}

	t.Log("resolves bundle id in format: $(ENV_KEY:rfc1034identifier)")
	{
		bundleID := `$(PRODUCT_NAME:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier)")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id with cross env reference")
	{
		bundleID := `auto_provision.$(BUNDLE_ID:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
			"BUNDLE_ID":    "$(PRODUCT_NAME:rfc1034identifier)",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("detects env refernce cycle")
	{
		bundleID := `auto_provision.${BUNDLE_ID:rfc1034identifier}`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "${BUNDLE_ID:rfc1034identifier}",
			"BUNDLE_ID":    "${PRODUCT_NAME:rfc1034identifier}",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.EqualError(t, err, "bundle id reference cycle found")
		require.Equal(t, "", resolved)
	}

	t.Log("resolves bundle id in format: prefix.${ENV_KEY:rfc1034identifier}.suffix")
	{
		bundleID := `auto_provision.${PRODUCT_NAME:rfc1034identifier}.suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix", resolved)
	}

	t.Log("resolves bundle id in format: ${ENV_KEY:rfc1034identifier}")
	{
		bundleID := `${PRODUCT_NAME:rfc1034identifier}`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix.$(ENV_KEY:rfc1034identifier)")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix.$(PRODUCT_NAME:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix.$(ENV_KEY_2:rfc1034identifier)")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix.$(VERSION:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
			"VERSION":      "beta",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix.beta", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEY.suffix")
	{
		bundleID := `auto_provision.$PRODUCT_NAME.suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix", resolved)
	}
	t.Log("resolves bundle id in format: prefix.second.${ENV_KEY}.suffix")
	{
		bundleID := `prefix.second.${PRODUCT_NAME}.suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "prefix.second.ios-simple-objc.suffix", resolved)
	}
	t.Log("resolves bundle id in format: prefix.second.third.${ENV_KEY}")
	{
		bundleID := `prefix.second.third.${PRODUCT_NAME}`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "prefix.second.third.ios-simple-objc", resolved)
	}
	t.Log("resolves bundle id in format: prefix.second.third.fourth.${ENV_KEY}")
	{
		bundleID := `prefix.second.third.fourth.${PRODUCT_NAME}`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := resolve(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "prefix.second.third.fourth.ios-simple-objc", resolved)
	}
}

func TestExpand(t *testing.T) {

	// Complex env
	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix.$(ENV_KEY:rfc1034identifier)")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix.$(PRODUCT_NAME:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expand(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expand(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix", resolved)
	}

	// Simple env
	t.Log("resolves bundle id in format: prefix.$ENV_KEY")
	{
		bundleID := `auto_provision.$PRODUCT_NAME`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expand(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEYsuffix")
	{
		bundleID := `auto_provision.$PRODUCT_NAMEsuffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expand(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objcsuffix", resolved)
	}
}

func TestExpandComplexEnv(t *testing.T) {
	t.Log("resolves bundle id in format: prefix.{text.${ENV_KEY}.text}")
	{
		bundleID := `prefix.{text.${PRODUCT_NAME}.text}`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expandComplexEnv(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "prefix.{text.ios-simple-objc.text}", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix.$(ENV_KEY:rfc1034identifier)")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix.$(PRODUCT_NAME:rfc1034identifier)`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expandComplexEnv(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$(ENV_KEY:rfc1034identifier).suffix")
	{
		bundleID := `auto_provision.$(PRODUCT_NAME:rfc1034identifier).suffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expandComplexEnv(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc.suffix", resolved)
	}
}

func TestExpandSimpleEnv(t *testing.T) {
	t.Log("resolves bundle id in format: prefix.$ENV_KEY")
	{
		bundleID := `auto_provision.$PRODUCT_NAME`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expandSimpleEnv(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objc", resolved)
	}

	t.Log("resolves bundle id in format: prefix.$ENV_KEYsuffix")
	{
		bundleID := `auto_provision.$PRODUCT_NAMEsuffix`
		buildSettings := serialized.Object{
			"PRODUCT_NAME": "ios-simple-objc",
		}
		resolved, err := expandSimpleEnv(bundleID, buildSettings)
		require.NoError(t, err)
		require.Equal(t, "auto_provision.ios-simple-objcsuffix", resolved)
	}
}

func TestTargets(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	project, err := Open(filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj"))
	require.NoError(t, err)

	{
		// Target with dependencies
		target, ok := project.Proj.Target("7D0342D720F4B5AD0050B6A6")
		require.True(t, ok)

		dependentTargets := project.DependentTargetsOfTarget(target)
		require.Equal(t, 2, len(dependentTargets))
		require.Equal(t, "WatchKitApp", dependentTargets[0].Name)
		require.Equal(t, "WatchKitApp Extension", dependentTargets[1].Name)
	}

	{
		// Target with no dependencies
		target, ok := project.Proj.Target("7D03432F20F4BBBD0050B6A6")
		require.True(t, ok)

		dependentTargets := project.DependentTargetsOfTarget(target)
		require.Empty(t, dependentTargets)
	}

	{
		settings, err := project.TargetBuildSettings("SubProject", "Debug")
		require.NoError(t, err)
		require.True(t, len(settings) > 0)

		bundleID, err := settings.String("PRODUCT_BUNDLE_IDENTIFIER")
		require.NoError(t, err)
		require.Equal(t, "com.bitrise.SubProject", bundleID)

		infoPlist, err := settings.String("INFOPLIST_PATH")
		require.NoError(t, err)
		require.Equal(t, "SubProject.app/Info.plist", infoPlist)
	}

	{
		bundleID, err := project.TargetBundleID("SubProject", "Debug")
		require.NoError(t, err)
		require.Equal(t, "com.bitrise.SubProject", bundleID)
	}

	{
		properties, _, err := project.ReadTargetInfoplist("SubProject", "Debug")
		require.NoError(t, err)
		require.Equal(t, serialized.Object{"CFBundlePackageType": "APPL",
			"UISupportedInterfaceOrientations":      []interface{}{"UIInterfaceOrientationPortrait", "UIInterfaceOrientationLandscapeLeft", "UIInterfaceOrientationLandscapeRight"},
			"CFBundleInfoDictionaryVersion":         "6.0",
			"CFBundleName":                          "$(PRODUCT_NAME)",
			"UISupportedInterfaceOrientations~ipad": []interface{}{"UIInterfaceOrientationPortrait", "UIInterfaceOrientationPortraitUpsideDown", "UIInterfaceOrientationLandscapeLeft", "UIInterfaceOrientationLandscapeRight"},
			"CFBundleDevelopmentRegion":             "$(DEVELOPMENT_LANGUAGE)",
			"CFBundleExecutable":                    "$(EXECUTABLE_NAME)",
			"CFBundleShortVersionString":            "1.0",
			"CFBundleVersion":                       "1",
			"LSRequiresIPhoneOS":                    true,
			"UIMainStoryboardFile":                  "Main",
			"UIRequiredDeviceCapabilities":          []interface{}{"armv7"},
			"CFBundleIdentifier":                    "$(PRODUCT_BUNDLE_IDENTIFIER)",
			"UILaunchStoryboardName":                "LaunchScreen"}, properties)
	}

	{
		entitlements, err := project.TargetCodeSignEntitlements("WatchKitApp", "Debug")
		require.NoError(t, err)
		require.Equal(t, serialized.Object{"com.apple.security.application-groups": []interface{}{}}, entitlements)

	}
}

func TestScheme(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	pth := filepath.Join(dir, "XcodeProj.xcodeproj")
	project, err := Open(pth)
	require.NoError(t, err)

	{
		scheme, container, err := project.Scheme("ProjectTodayExtensionScheme")
		require.NoError(t, err)
		require.Equal(t, "ProjectTodayExtensionScheme", scheme.Name)
		require.Equal(t, pth, container)
	}

	{
		scheme, container, err := project.Scheme("NotExistScheme")
		require.EqualError(t, err, "scheme NotExistScheme not found in XcodeProj")
		require.Equal(t, (*xcscheme.Scheme)(nil), scheme)
		require.Equal(t, "", container)
	}

	{
		// Gdańsk represented in High Sierra
		b := []byte{71, 100, 97, 197, 132, 115, 107}
		scheme, container, err := project.Scheme(string(b))
		require.NoError(t, err)
		require.Equal(t, norm.NFC.String(string(b)), norm.NFC.String(scheme.Name))
		require.Equal(t, pth, container)
	}

	{
		// Gdańsk represented in Mojave
		b := []byte{71, 100, 97, 110, 204, 129, 115, 107}
		scheme, container, err := project.Scheme(string(b))
		require.NoError(t, err)
		require.Equal(t, norm.NFC.String(string(b)), norm.NFC.String(scheme.Name))
		require.Equal(t, pth, container)
	}
}

func TestSchemes(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	require.NoError(t, err)

	schemes, err := project.Schemes()
	require.NoError(t, err)
	require.Equal(t, 3, len(schemes))

	// Gdańsk represented in High Sierra
	b := []byte{71, 100, 97, 197, 132, 115, 107}
	require.Equal(t, norm.NFC.String(string(b)), norm.NFC.String(schemes[0].Name))
	require.Equal(t, "ProjectScheme", schemes[1].Name)

	// Gdańsk represented in Mojave
	b = []byte{71, 100, 97, 110, 204, 129, 115, 107}
	require.Equal(t, norm.NFC.String(string(b)), norm.NFC.String(schemes[0].Name))
	require.Equal(t, "ProjectScheme", schemes[1].Name)
}

func TestOpenXcodeproj(t *testing.T) {
	t.Log("Opening Pods.xcodeproj in sample-apps-ios-workspace-swift.git")
	{
		dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/sample-apps-ios-workspace-swift.git")
		project, err := Open(filepath.Join(dir, "Pods", "Pods.xcodeproj"))
		require.NoError(t, err)
		require.Equal(t, filepath.Join(dir, "Pods", "Pods.xcodeproj"), project.Path)
		require.Equal(t, "Pods", project.Name)
	}
	t.Log("Opening XcodeProj.xcodeproj in xcode-project-test.git")
	{
		dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
		project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
		require.NoError(t, err)
		require.Equal(t, filepath.Join(dir, "XcodeProj.xcodeproj"), project.Path)
		require.Equal(t, "XcodeProj", project.Name)
	}
}

func TestIsXcodeProj(t *testing.T) {
	require.True(t, IsXcodeProj("./BitriseSample.xcodeproj"))
	require.False(t, IsXcodeProj("./BitriseSample.xcworkspace"))
}

func TestXcodeProj_forceBundleID(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	if err != nil {
		t.Fatalf("Failed to init project for test case, error: %s", err)
	}

	tests := []struct {
		name          string
		target        string
		configuration string
		bundleID      string
		wantErr       bool
	}{
		{
			name:          "Update bundle ID for target and configuration",
			target:        "XcodeProj",
			configuration: "Release",
			bundleID:      "io.bitrise.test.XcodeProj",
			wantErr:       false,
		},
		{
			name:    "Target not found",
			target:  "NON_EXISTENT_TARGET",
			wantErr: true,
		},
		{
			name:          "Configuration not found",
			target:        "XcodeProj",
			configuration: "NON_EXISTENT_CONFIGURATION",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := project.ForceTargetBundleID(tt.target, tt.configuration, tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error: %s", err)
			} else if tt.wantErr {
				return
			}

			got, err := project.TargetBundleID(tt.target, tt.configuration)
			assert.NoError(t, err)
			assert.Equal(t, tt.bundleID, got)
		})
	}
}

func TestXcodePrj_forceTargetCodeSignEntitlement(t *testing.T) {
	dir := testhelper.GitCloneIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git")
	project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	if err != nil {
		t.Fatalf("Failed to init project for test case, error: %s", err)
	}

	tests := []struct {
		name          string
		target        string
		configuration string
		entitlement   string
		value         string
		wantErr       bool
	}{
		{
			name:          "Update entitlement",
			target:        "TodayExtension",
			configuration: "Release",
			entitlement:   "com.apple.security.application-groups",
			value:         "io.bitrise.test",
			wantErr:       false,
		},
		{
			name:    "Target not found",
			target:  "NON_EXISTENT_TARGET",
			wantErr: true,
		},
		{
			name:          "Configuration not found",
			configuration: "NON_EXISTENT_CONFIGURATION",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := project.ForceTargetCodeSignEntitlement(tt.target, tt.configuration, tt.entitlement, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error: %s, but wanErr = %t", err, tt.wantErr)
			}

			if got, err := project.TargetCodeSignEntitlements(tt.target, tt.configuration); (err != nil) != tt.wantErr {
				t.Fatalf("error validating test: %s", err)
			} else if err == nil && got[tt.entitlement] != tt.value {
				t.Fatalf("got %s, want %s", got[tt.entitlement], tt.value)
			}
		})
	}
}

func TestXcodeProj_ForceCodeSign(t *testing.T) {
	// arrange
	proj, err := parsePBXProjContent([]byte(pbxprojWithouthTargetAttributes))
	require.NoError(t, err)

	configurationName := "Debug"
	targetName := "Target"

	team := "ABCD1234"
	signingIdentity := "Apple Development: John Doe (ASDF1234)"
	provisioningProfile := "asdf56b6-e75a-4f86-bf25-101bfc2fasdf"

	// act
	err = proj.ForceCodeSign(configurationName, targetName, team, signingIdentity, provisioningProfile)
	require.NoError(t, err)

	// assert
	target := findTarget(t, proj, targetName)

	targetAttr := projectTargetAttributes(t, proj, target.ID)
	ensureValue(t, targetAttr, "ProvisioningStyle", "Manual")
	ensureValue(t, targetAttr, "DevelopmentTeam", team)
	ensureValue(t, targetAttr, "DevelopmentTeamName", "")

	targetBuildConfig := findBuildConfiguration(t, target, configurationName)
	ensureValue(t, targetBuildConfig.BuildSettings, "CODE_SIGN_STYLE", "Manual")
	ensureValue(t, targetBuildConfig.BuildSettings, "DEVELOPMENT_TEAM", team)
	ensureValue(t, targetBuildConfig.BuildSettings, "CODE_SIGN_IDENTITY", signingIdentity)
	ensureValue(t, targetBuildConfig.BuildSettings, "CODE_SIGN_IDENTITY[sdk=iphoneos*]", signingIdentity)
	ensureValue(t, targetBuildConfig.BuildSettings, "PROVISIONING_PROFILE_SPECIFIER", "")
	ensureValue(t, targetBuildConfig.BuildSettings, "PROVISIONING_PROFILE", provisioningProfile)
}

func TestXcodeProj_ForceCodeSign_WithouthTargetAttributes(t *testing.T) {
	// arrange
	proj, err := parsePBXProjContent([]byte(pbxprojWithouthTargetAttributes))
	require.NoError(t, err)

	configurationName := "Debug"
	targetName := "TargetWithouthTargetAttributes"

	team := "ABCD1234"
	signingIdentity := "Apple Development: John Doe (ASDF1234)"
	provisioningProfile := "asdf56b6-e75a-4f86-bf25-101bfc2fasdf"

	// act
	err = proj.ForceCodeSign(configurationName, targetName, team, signingIdentity, provisioningProfile)
	require.NoError(t, err)

	// assert
	target := findTarget(t, proj, targetName)

	targetBuildConfig := findBuildConfiguration(t, target, configurationName)
	ensureValue(t, targetBuildConfig.BuildSettings, "CODE_SIGN_STYLE", "Manual")
	ensureValue(t, targetBuildConfig.BuildSettings, "DEVELOPMENT_TEAM", team)
	ensureValue(t, targetBuildConfig.BuildSettings, "CODE_SIGN_IDENTITY", signingIdentity)
	ensureValue(t, targetBuildConfig.BuildSettings, "PROVISIONING_PROFILE_SPECIFIER", "")
	ensureValue(t, targetBuildConfig.BuildSettings, "PROVISIONING_PROFILE", provisioningProfile)
}

func TestXcodeProj_ForceCodeSign_OverridesSigningBuildSettingsOnly(t *testing.T) {
	// arrange
	proj, err := parsePBXProjContent([]byte(pbxprojWithouthTargetAttributes))
	require.NoError(t, err)

	configurationName := "Debug"
	targetName := "TargetWithouthTargetAttributes"

	team := "ABCD1234"
	signingIdentity := "Apple Development: John Doe (ASDF1234)"
	provisioningProfile := "asdf56b6-e75a-4f86-bf25-101bfc2fasdf"

	// act
	err = proj.ForceCodeSign(configurationName, targetName, team, signingIdentity, provisioningProfile)
	require.NoError(t, err)

	// assert
	target := findTarget(t, proj, targetName)

	targetBuildConfig := findBuildConfiguration(t, target, configurationName)
	ensureValue(t, targetBuildConfig.BuildSettings, "INFOPLIST_FILE", "Target copy-Info.plist")
}

func ensureValue(t *testing.T, obj serialized.Object, key, value string) {
	v, err := obj.String(key)
	require.NoError(t, err)
	require.Equal(t, value, v)
}

func findTarget(t *testing.T, project *XcodeProj, name string) Target {
	var target Target
	for _, t := range project.Proj.Targets {
		if t.Name == name {
			target = t
			break
		}
	}
	require.NotNil(t, target)
	return target
}

func projectTargetAttributes(t *testing.T, project *XcodeProj, targetID string) serialized.Object {
	attr, err := project.Proj.Attributes.TargetAttributes.Object(targetID)
	require.NoError(t, err)
	return attr
}

func findBuildConfiguration(t *testing.T, target Target, name string) BuildConfiguration {
	var config BuildConfiguration
	for _, c := range target.BuildConfigurationList.BuildConfigurations {
		if c.Name == name {
			config = c
		}
	}
	require.NotNil(t, config)
	return config
}

func TestXcodeProjOpen_AposthropeSupported(t *testing.T) {
	// Arrange
	dir := testhelper.GitCloneBranchIntoTmpDir(t, "https://github.com/bitrise-io/xcode-project-test.git", "special-character")
	project, err := Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	if err != nil {
		t.Fatalf("Failed to init project for test case, error: %s", err)
	}

	// Act + Assert
	objects, errObjects := project.RawProj.Object("objects")
	if errObjects != nil {
		t.Errorf("Failed reading test project, error: %s", errObjects)
	}
	jsonGroup, errGroup := objects.Object("0F8FD97B23F5831A006C13DE")
	if errGroup != nil {
		t.Errorf("Failed reading test project, error: %s", errGroup)
	}
	path, errPath := jsonGroup.String("path")
	if errObjects != nil {
		t.Errorf("Failed reading test project, error: %s", errPath)
	}

	if path != "JSON's" {
		t.Errorf("Test project modified, file does not contain special characters")
	}

	if err := project.Save(); err != nil {
		t.Errorf("Failed to save project, error: %s", err)
	}
	_, err = Open(filepath.Join(dir, "XcodeProj.xcodeproj"))
	if err != nil {
		t.Fatalf("Failed to reopen project after saving it, error: %s", err)
	}
}

func TestXcodeProj_perObjectModify(t *testing.T) {
	tests := []struct {
		name                  string
		projContent           string
		configuration, target string
		want                  []byte
		wantErr               bool
	}{
		{
			name:          "No target attributes",
			projContent:   pbxprojWithouthTargetAttributes,
			configuration: "Debug",
			target:        "TargetWithouthTargetAttributes",
			want:          []byte(pbxprojWTAafterPerObjectModify),
		},
		{
			name:          "Will change 2 objects (as Target attributes is included in the project)",
			projContent:   testhelper.XcodeProjectTest,
			configuration: "Debug",
			target:        "XcodeProj",
			want:          []byte(testhelper.XcodeProjectTestChanged),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, err := parsePBXProjContent([]byte(tt.projContent))
			require.NoError(t, err)

			team := "ABCD1234"
			signingIdentity := "Apple Development: John Doe (ASDF1234)"
			provisioningProfile := "asdf56b6-e75a-4f86-bf25-101bfc2fasdf"

			err = proj.ForceCodeSign(tt.configuration, tt.target, team, signingIdentity, provisioningProfile)
			require.NoError(t, err)

			got, err := proj.perObjectModify()
			if (err != nil) != tt.wantErr {
				t.Errorf("XcodeProj.perObjectModify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, string(tt.want), string(got), "XcodeProj.perObjectModify() =")
		})
	}
}

func Test_removeCustomInfo(t *testing.T) {
	tests := []struct {
		o    interface{}
		name string
		want interface{}
	}{
		{
			name: "map",
			o: map[string]interface{}{
				"a":                 "b",
				customAnnotationKey: "dsad",
			},
			want: map[string]interface{}{
				"a": "b",
			},
		},
		{
			name: "map -> map",
			o: map[string]interface{}{
				"a": map[string]interface{}{
					"a":                 "b",
					customAnnotationKey: "dsad",
				},
			},
			want: map[string]interface{}{
				"a": map[string]interface{}{
					"a": "b",
				},
			},
		},
		{
			name: "array -> map",
			o: []interface{}{
				map[string]interface{}{
					"a":                 "b",
					customAnnotationKey: "dsad",
				},
			},
			want: []interface{}{
				map[string]interface{}{
					"a": "b",
				},
			},
		},
		{
			name: "map -> array -> map",
			o: map[string]interface{}{
				"a": []interface{}{
					map[string]interface{}{
						"a": map[string]interface{}{
							"a":                 "b",
							customAnnotationKey: "dsad",
						},
					},
				},
			},
			want: map[string]interface{}{
				"a": []interface{}{
					map[string]interface{}{
						"a": map[string]interface{}{
							"a": "b",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeCustomInfo(tt.o)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_deepCopyObject(t *testing.T) {
	object := map[string]interface{}{
		"d": []interface{}{
			map[string]interface{}{"a": "b"},
		},
	}

	got := deepCopyObject(object)
	require.EqualValues(t, object, got, "deepCopyObject() returns identical value as input")

	internalArray, ok := object["d"].([]interface{})
	require.True(t, ok)
	internalMap, ok := internalArray[0].(map[string]interface{})
	require.True(t, ok)
	internalMap["a"] = "c"

	require.NotEqual(t, object, got, "deepCopyObject() changing copied object does not change original")
}

func Test_deduplicateTargetList(t *testing.T) {
	tests := []struct {
		name    string
		targets []Target
		want    []Target
	}{
		{
			"Empty slice",
			[]Target{},
			[]Target{},
		},
		{
			"Single item",
			[]Target{{ID: "610F554E26158E0A001D3AA0"}},
			[]Target{{ID: "610F554E26158E0A001D3AA0"}},
		},
		{
			"Multiple unique items",
			[]Target{
				{ID: "610F554E26158E0A001D3AA0"},
				{ID: "612F3D832615A2F400137D77"},
			},
			[]Target{
				{ID: "610F554E26158E0A001D3AA0"},
				{ID: "612F3D832615A2F400137D77"},
			},
		},
		{
			"Duplicated items",
			[]Target{
				{ID: "610F554E26158E0A001D3AA0"},
				{ID: "612F3D832615A2F400137D77"},
				{ID: "610F554E26158E0A001D3AA0"},
				{ID: "612F3D832615A2F400137D77"},
			},
			[]Target{
				{ID: "610F554E26158E0A001D3AA0"},
				{ID: "612F3D832615A2F400137D77"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, deduplicateTargetList(tt.targets), "deduplicateTargetList(%v)", tt.targets)
		})
	}
}
