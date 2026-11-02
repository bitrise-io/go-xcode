package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Target IDs from internal/xcproj/testdata/ios-sample.xcproj.
const (
	xcprojAppTargetID      = "16F0774C306FEF4F008B9392"
	xcprojUnitTestTargetID = "16F07759306FEF50008B9392"
	xcprojUITestTargetID   = "16F07763306FEF50008B9392"
	xcprojAppTargetName    = "sample-apps-ios-json-xcproj"
	xcprojFixtureFileName  = "ios-sample.xcproj"
)

func openXCProjFixture(t *testing.T) XcodeProj {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("internal", "xcproj", "testdata", xcprojFixtureFileName))
	require.NoError(t, err)

	return openProjectFiles(t, map[string][]byte{xcProjFileName: content})
}

func openProjectFiles(t *testing.T, files map[string][]byte) XcodeProj {
	t.Helper()

	projectPath := writeProjectFiles(t, files)
	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	return project
}

func writeProjectFiles(t *testing.T, files map[string][]byte) string {
	t.Helper()

	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(projectPath, name), content, 0644))
	}

	return projectPath
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

func TestXCProj_dependencies(t *testing.T) {
	project := openProjectFiles(t, map[string][]byte{
		xcProjFileName: []byte(`{"targets": [
			{"name": "Lib", "id": "L1"},
			{"name": "Twin", "id": "T1"},
			{"name": "Twin", "id": "T2"},
			{"name": "App", "id": "A1", "dependencies": ["Lib", "Gone", "Twin", "id:T2", "id:X9"]}
		]}`),
	})

	app, ok := project.Target("A1")
	require.True(t, ok)
	assert.Equal(t, []string{"L1", "T2"}, app.dependencyTargetIDs)
}

func TestXCProj_productReferencedByID(t *testing.T) {
	project := openProjectFiles(t, map[string][]byte{
		xcProjFileName: []byte(`{"targets": [{"name": "App", "id": "A1", "product-type": "application", "product": "id:P1"}]}`),
	})

	require.Len(t, project.Targets(), 1)
	app := project.Targets()[0]
	assert.Empty(t, app.productPath)
	assert.False(t, app.IsExecutableProduct())
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

func TestXCProj_TargetDevelopmentTeam(t *testing.T) {
	project := openXCProjFixture(t)

	_, ok := project.TargetDevelopmentTeam(xcprojAppTargetID)
	assert.False(t, ok)
}

func TestXCProj_notSupported(t *testing.T) {
	project := openXCProjFixture(t)

	_, err := project.AppIconSetPaths()
	assert.ErrorIs(t, err, errNotSupportedForXCProj)
	assert.ErrorIs(t, project.ForceCodeSign(ForceCodeSignOptions{}), errNotSupportedForXCProj)
	assert.ErrorIs(t, project.SetBuildSetting(xcprojAppTargetName, "Release", "MARKETING_VERSION", "2.0"), errNotSupportedForXCProj)
	assert.ErrorIs(t, project.Save(), errNotSupportedForXCProj)
}
