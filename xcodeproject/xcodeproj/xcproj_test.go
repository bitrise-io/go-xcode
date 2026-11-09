package xcodeproj

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/internal/xcproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

const (
	xcprojAppTargetID      = "16F0774C306FEF4F008B9392"
	xcprojUnitTestTargetID = "16F07759306FEF50008B9392"
	xcprojUITestTargetID   = "16F07763306FEF50008B9392"
	xcprojAppTargetName    = "sample-apps-ios-json-xcproj"
	xcprojFixtureFileName  = "ios-sample.xcproj"
)

func openXCProjFixture(t *testing.T) XcodeProj {
	t.Helper()

	return openXCProjFixtureFile(t, xcprojFixtureFileName)
}

func openXCProjFixtureFile(t *testing.T, name string) XcodeProj {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("internal", "xcproj", "testdata", name))
	require.NoError(t, err)

	return openProjectFiles(t, map[string][]byte{xcProjFileName: content})
}

func TestXCProj_Targets(t *testing.T) {
	project := openXCProjFixture(t)

	require.IsType(t, &XCProj{}, project)
	assert.Equal(t, "App", project.Name())

	targets := project.Targets()
	require.Len(t, targets, 3)

	app, unitTests, uiTests := targets[0], targets[1], targets[2]
	assert.Equal(t, []string{xcprojAppTargetName, xcprojAppTargetName + "Tests", xcprojAppTargetName + "UITests"}, targetNames(targets))
	assert.Equal(t, []string{xcprojAppTargetID, xcprojUnitTestTargetID, xcprojUITestTargetID}, []string{targets[0].ID, targets[1].ID, targets[2].ID})

	for _, target := range targets {
		assert.Equal(t, nativeTargetKind, target.kind, target.Name)
		assert.Equal(t, "Release", target.DefaultConfigurationName, target.Name)
	}

	assert.Equal(t, "com.apple.product-type.application", app.productType)
	assert.Equal(t, "sample-apps-ios-json-xcproj.app", app.productPath)
	assert.True(t, app.IsExecutableProduct())
	assert.Empty(t, app.dependencyTargetIDs)

	assert.Equal(t, "com.apple.product-type.bundle.unit-test", unitTests.productType)
	assert.True(t, unitTests.isTest())

	assert.Equal(t, "com.apple.product-type.bundle.ui-testing", uiTests.productType)
	assert.True(t, uiTests.IsUITestProduct())
	assert.True(t, uiTests.DependsOn(xcprojAppTargetID))
	assert.Equal(t, []string{xcprojAppTargetName}, targetNames(project.DependentTargetsOfTarget(uiTests)))
}

func TestXCProj_BuildConfigurations(t *testing.T) {
	project := openXCProjFixture(t)

	assert.Equal(t, "Release", project.DefaultConfigurationName())

	configurations := project.BuildConfigurations()
	require.Len(t, configurations, 2)
	debug, release := configurations[0], configurations[1]
	assert.Equal(t, "Debug", debug.Name)
	assert.Equal(t, "Release", release.Name)

	value, ok := debug.BuildSetting("DEBUG_INFORMATION_FORMAT")
	assert.True(t, ok)
	assert.Equal(t, "dwarf", value)

	value, ok = release.BuildSetting("DEBUG_INFORMATION_FORMAT")
	assert.True(t, ok)
	assert.Equal(t, "dwarf-with-dsym", value)

	_, ok = release.BuildSetting("ENABLE_TESTABILITY")
	assert.False(t, ok)

	value, ok = release.BuildSetting("SDKROOT")
	assert.True(t, ok)
	assert.Equal(t, "iphoneos", value)

	_, ok = debug.BuildSetting("GCC_PREPROCESSOR_DEFINITIONS")
	assert.False(t, ok, "array values are not strings")

	app, ok := project.Target(xcprojAppTargetID)
	require.True(t, ok)
	require.Len(t, app.BuildConfigurations, 2)
	value, ok = app.BuildConfigurations[1].BuildSetting("PRODUCT_BUNDLE_IDENTIFIER")
	assert.True(t, ok)
	assert.Equal(t, "io.bitrise.sample-apps-ios-json-xcproj", value)
}

func TestXCProj_BuildConfigurations_noBuildSettings(t *testing.T) {
	project := openProjectFiles(t, map[string][]byte{
		xcProjFileName: []byte(`{"configurations": ["Debug"], "targets": [{"name": "App", "id": "A1"}]}`),
	})

	require.Len(t, project.BuildConfigurations(), 1)
	assert.NotNil(t, project.BuildConfigurations()[0].buildSettings)

	require.Len(t, project.Targets(), 1)
	require.Len(t, project.Targets()[0].BuildConfigurations, 1)
	assert.NotNil(t, project.Targets()[0].BuildConfigurations[0].buildSettings)
}

func TestXCProj_unknownDependency(t *testing.T) {
	projectPath := writeProjectFiles(t, map[string][]byte{
		xcProjFileName: []byte(`{"targets": [
			{"name": "Lib", "id": "L1"},
			{"name": "App", "id": "A1", "dependencies": ["Lib", "Gone"]}
		]}`),
	})

	project, err := testFactory().Open(projectPath)
	require.ErrorContains(t, err, "target App depends on unknown target Gone")
	assert.True(t, project == nil, "got a non-nil %T", project)
}

func TestXCProj_productReferencedByID(t *testing.T) {
	project := openXCProjFixtureFile(t, "ios-sample-twin-and-xcconfig-targets.xcproj")

	assert.Equal(t, map[string]string{
		xcprojAppTargetName:             "sample-apps-ios-json-xcproj.app",
		xcprojAppTargetName + "Tests":   "sample-apps-ios-json-xcprojTests.xctest",
		xcprojAppTargetName + "UITests": "sample-apps-ios-json-xcprojUITests.xctest",
		"Twin":                          "Twin.app",
		"TwinCopy":                      "Twin.app",
		"XcconfigOnly":                  "XcconfigOnly.app",
	}, productPaths(project))

	twin := requireTarget(t, project, "Twin")
	assert.True(t, twin.IsExecutableProduct())

	assert.Equal(t, map[string]string{
		xcprojAppTargetName: "sample-apps-ios-json-xcproj.app",
		"Twin":              "Twin.app",
		"TwinCopy":          "Twin.app",
		"XcconfigOnly":      "XcconfigOnly.app",
	}, schemeBuildableNames(project.RecreateSchemes()))
}

func TestXCProj_productReferencedByUnknownID(t *testing.T) {
	project, err := xcproj.Decode([]byte(`{"targets": [{"name": "App", "id": "A1", "product-type": "application", "product": "id:P1"}]}`))
	require.NoError(t, err)
	var output bytes.Buffer

	targets := xcprojTargets(project, log.NewLogger(log.WithOutput(&output)))

	require.Len(t, targets, 1)
	assert.Empty(t, targets[0].productPath)
	assert.False(t, targets[0].IsExecutableProduct())
	assert.Contains(t, output.String(), "product of target App (A1) is referenced by ID P1, its path is unknown")
}

func TestXCProj_productWithoutName(t *testing.T) {
	project := openXCProjFixtureFile(t, "products-without-name.xcproj")

	assert.Equal(t, map[string]string{
		"Reader":                       "Reader.app",
		"ReaderTests":                  "ReaderTests.xctest",
		"Reader-iOS":                   "Reader-iOS.app",
		"Reader Subscribe Extension":   "Reader Subscribe Extension.appex",
		"Reader iOS Share Extension":   "Reader iOS Share Extension.appex",
		"Reader iOS Intents Extension": "Reader iOS Intents Extension.appex",
		"Reader iOS Widget Extension":  "Reader iOS Widget Extension.appex",
		"Reader-iOSTests":              "Reader-iOSTests.xctest",
		"Reader Share Extension":       "Reader Share Extension.appex",
	}, productPaths(project))

	schemes := project.RecreateSchemes()
	assert.Equal(t, "Reader.app", schemeBuildableNames(schemes)["Reader"])

	var testables []string
	for _, scheme := range schemes {
		for _, testable := range scheme.TestAction.Testables {
			testables = append(testables, testable.BuildableReference.BuildableName)
		}
	}
	assert.Equal(t, []string{"ReaderTests.xctest", "Reader-iOSTests.xctest"}, testables)
}

func TestXCProjProductName(t *testing.T) {
	assert.Equal(t, "App.app", xcprojProductName("App", ".app"))
	assert.Equal(t, "Other.app", xcprojProductName("App", "Other.app"))
	assert.Empty(t, xcprojProductName("App", ""))
}

func productPaths(project XcodeProj) map[string]string {
	paths := map[string]string{}
	for _, target := range project.Targets() {
		paths[target.Name] = target.productPath
	}
	return paths
}

func schemeBuildableNames(schemes []xcscheme.Scheme) map[string]string {
	names := map[string]string{}
	for _, scheme := range schemes {
		names[scheme.Name] = scheme.BuildAction.BuildActionEntries[0].BuildableReference.BuildableName
	}
	return names
}

func TestXCProj_targetKinds(t *testing.T) {
	project := openProjectFiles(t, map[string][]byte{
		xcProjFileName: []byte(`{"targets": [
			{"name": "App", "id": "A1"},
			{"name": "Native", "id": "N1", "kind": "native"},
			{"name": "Aggregate", "id": "G1", "kind": "aggregate"},
			{"name": "External", "id": "E1", "kind": "external-build-system"}
		]}`),
	})

	var kinds []targetKind
	for _, target := range project.Targets() {
		kinds = append(kinds, target.kind)
	}
	assert.Equal(t, []targetKind{nativeTargetKind, nativeTargetKind, aggregateTargetKind, externalBuildSystemTargetKind}, kinds)
}

func TestXCProj_RecreateSchemes(t *testing.T) {
	project := openXCProjFixture(t)

	schemes := project.RecreateSchemes()
	require.Len(t, schemes, 1)

	scheme := schemes[0]
	assert.Equal(t, xcprojAppTargetName, scheme.Name)
	require.Len(t, scheme.BuildAction.BuildActionEntries, 1)
	assert.Equal(t, "sample-apps-ios-json-xcproj.app", scheme.BuildAction.BuildActionEntries[0].BuildableReference.BuildableName)

	var testables []string
	for _, testable := range scheme.TestAction.Testables {
		testables = append(testables, testable.BuildableReference.BlueprintName)
	}
	assert.Equal(t, []string{xcprojAppTargetName + "Tests", xcprojAppTargetName + "UITests"}, testables)
}

func TestXCProj_SetBuildSetting(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("internal", "xcproj", "testdata", xcprojFixtureFileName))
	require.NoError(t, err)
	projectPath := writeProjectFiles(t, map[string][]byte{xcProjFileName: content})
	xcProjPath := filepath.Join(projectPath, xcProjFileName)
	require.NoError(t, os.Chmod(xcProjPath, 0664))

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	fetchedBefore := requireTarget(t, project, xcprojAppTargetName)

	require.NoError(t, project.SetBuildSetting(xcprojAppTargetName, "Release", "CODE_SIGN_STYLE", "Manual"))
	require.NoError(t, project.SetBuildSetting(xcprojAppTargetName, "Release", "CODE_SIGN_IDENTITY[sdk=iphoneos*]", "Apple Distribution"))
	require.NoError(t, project.SetBuildSetting(xcprojAppTargetName, "Debug", "MARKETING_VERSION", "2.0"))

	release, err := targetBuildSettings(fetchedBefore, "Release")
	require.NoError(t, err)
	value, ok := release.String("CODE_SIGN_STYLE")
	assert.True(t, ok)
	assert.Equal(t, "Manual", value, "targets fetched earlier see the change")

	require.NoError(t, project.Save())

	reopened, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	assert.Equal(t, reopened.Targets(), project.Targets())
	assert.Equal(t, reopened.BuildConfigurations(), project.BuildConfigurations())

	requireBuildSetting(t, reopened, xcprojAppTargetName, "Debug", "CODE_SIGN_STYLE", "Automatic")
	requireBuildSetting(t, reopened, xcprojAppTargetName, "Release", "CODE_SIGN_STYLE", "Manual")
	requireBuildSetting(t, reopened, xcprojAppTargetName, "Release", "CODE_SIGN_IDENTITY[sdk=iphoneos*]", "Apple Distribution")
	requireBuildSetting(t, reopened, xcprojAppTargetName, "Debug", "MARKETING_VERSION", "2.0")
	requireBuildSetting(t, reopened, xcprojAppTargetName, "Release", "MARKETING_VERSION", "1.0")

	info, err := os.Stat(xcProjPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0664), info.Mode().Perm(), "an existing file keeps its mode")
}

func TestXCProj_Save_noEdit(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("internal", "xcproj", "testdata", xcprojFixtureFileName))
	require.NoError(t, err)
	projectPath := writeProjectFiles(t, map[string][]byte{xcProjFileName: content})

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	require.NoError(t, project.Save())

	saved, err := os.ReadFile(filepath.Join(projectPath, xcProjFileName))
	require.NoError(t, err)
	assert.Equal(t, string(content), string(saved))
}

func TestXCProj_SetBuildSetting_errors(t *testing.T) {
	tests := []struct {
		name          string
		target        string
		configuration string
		key           string
		value         any
		wantErr       string
	}{
		{
			name:          "non-string value",
			target:        "App",
			configuration: "Debug",
			key:           "OTHER_LDFLAGS",
			value:         []string{"-ObjC"},
			wantErr:       "failed to set build setting OTHER_LDFLAGS: a []string value is not supported for project.xcproj, only a string",
		},
		{
			name:          "unknown target",
			target:        "Other",
			configuration: "Debug",
			key:           "KEY",
			value:         "value",
			wantErr:       "failed to find target with name: Other",
		},
		{
			name:          "unknown configuration",
			target:        "App",
			configuration: "Release",
			key:           "KEY",
			value:         "value",
			wantErr:       "failed to find build configuration Release of target App",
		},
		{
			name:          "target without build settings",
			target:        "Lib",
			configuration: "Debug",
			key:           "KEY",
			value:         "value",
			wantErr:       "target Lib has no build settings: adding the first build setting to project.xcproj is not supported",
		},
		{
			name:          "key with a configuration condition",
			target:        "App",
			configuration: "Debug",
			key:           "KEY[config=Debug]",
			value:         "value",
			wantErr:       "invalid build setting key: KEY[config=Debug]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := []byte(`{"configurations": ["Debug"], "targets": [
				{"name": "App", "id": "A1", "build-settings": {"A": "1"}},
				{"name": "Lib", "id": "L1"}
			]}`)
			project := openProjectFiles(t, map[string][]byte{xcProjFileName: content})

			err := project.SetBuildSetting(tt.target, tt.configuration, tt.key, tt.value)
			require.EqualError(t, err, tt.wantErr)

			for _, target := range project.Targets() {
				_, ok := target.BuildConfigurations[0].BuildSetting(tt.key)
				assert.False(t, ok, "a failed edit leaves target %s unchanged", target.Name)
			}

			require.NoError(t, project.Save())
			saved, err := os.ReadFile(filepath.Join(project.Path(), xcProjFileName))
			require.NoError(t, err)
			assert.Equal(t, string(content), string(saved), "a failed edit leaves the file unchanged")
		})
	}
}

func TestXCProj_ForceCodeSign(t *testing.T) {
	const content = `{
  "configurations": [
    "Debug",
    "Release",
  ],
  "targets": [
    {
      "build-settings": {
        "CODE_SIGN_IDENTITY[sdk=iphoneos*]": "iPhone Developer",
        "CODE_SIGN_STYLE": "Automatic",
        "DEVELOPMENT_TEAM": "TEAM123456",
        "MARKETING_VERSION": "1.0",
      },
      "id": "A1",
      "legacy-provisioning-style": "automatic",
      "legacy-team-id": "TEAM123456",
      "name": "App",
    },
  ],
}
`
	const want = `{
  "configurations": [
    "Debug",
    "Release",
  ],
  "targets": [
    {
      "build-settings": {
        "CODE_SIGN_IDENTITY[config=Debug][sdk=iphoneos*]": "iPhone Developer",
        "CODE_SIGN_IDENTITY[config=Release]": "Apple Development: John Doe (ASDF1234)",
        "CODE_SIGN_IDENTITY[config=Release][sdk=iphoneos*]": "Apple Development: John Doe (ASDF1234)",
        "CODE_SIGN_STYLE[config=Debug]": "Automatic",
        "CODE_SIGN_STYLE[config=Release]": "Manual",
        "DEVELOPMENT_TEAM[config=Debug]": "TEAM123456",
        "DEVELOPMENT_TEAM[config=Release]": "ABCD1234",
        "MARKETING_VERSION": "1.0",
        "PROVISIONING_PROFILE[config=Release]": "asdf56b6-e75a-4f86-bf25-101bfc2fasdf",
        "PROVISIONING_PROFILE_SPECIFIER[config=Release]": "",
      },
      "id": "A1",
      "legacy-provisioning-style": "automatic",
      "legacy-team-id": "TEAM123456",
      "name": "App",
    },
  ],
}
`
	project := openProjectFiles(t, map[string][]byte{xcProjFileName: []byte(content)})

	require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("App", "Release")))
	require.NoError(t, project.Save())

	saved, err := os.ReadFile(filepath.Join(project.Path(), xcProjFileName))
	require.NoError(t, err)
	assert.Equal(t, want, string(saved))

	reopened, err := testFactory().Open(project.Path())
	require.NoError(t, err)
	assert.Equal(t, reopened.Targets(), project.Targets())
}

func TestXCProj_ForceCodeSign_unknownTargetOrConfiguration(t *testing.T) {
	project := openProjectFiles(t, map[string][]byte{
		xcProjFileName: []byte(`{"configurations": ["Debug"], "targets": [{"name": "App", "id": "A1", "build-settings": {"A": "1"}}]}`),
	})

	require.EqualError(t, project.ForceCodeSign(forceCodeSignOptions("Nope", "Debug")), "failed to find target with name: Nope")
	require.EqualError(t, project.ForceCodeSign(forceCodeSignOptions("App", "Nope")), "failed to find build configuration Nope of target App")
}
