package xcodeproj

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTarget_fromFixture(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.TargetByName("XcodeProj")
	require.True(t, ok)

	assert.Equal(t, appTargetID, target.ID)
	assert.Equal(t, nativeTargetKind, target.kind)
	assert.Equal(t, "com.apple.product-type.application", target.productType)
	assert.Equal(t, "XcodeProj.app", target.productPath)
	assert.Equal(t, []string{todayExtensionTarget}, target.dependencyTargetIDs)
	assert.Equal(t, "Release", target.DefaultConfigurationName)
	require.Len(t, target.BuildConfigurations, 2)
	assert.NotEmpty(t, pbxFile(t, project).buildPhaseIDs[target.ID])
}

func TestParseTarget_unknownTargetType(t *testing.T) {
	objects := map[string]any{
		"T1": map[string]any{"isa": "PBXSomethingElse", "name": "X"},
	}

	_, err := parseTarget("T1", objects)
	require.ErrorContains(t, err, "unknown target type")
}

func TestParseTarget_productReferenceWithoutPath(t *testing.T) {
	objects := map[string]any{
		"T1": map[string]any{
			"isa":                    nativeTargetISA,
			"name":                   "App",
			"buildConfigurationList": "L1",
			"productReference":       "P1",
		},
		"L1": map[string]any{"buildConfigurations": []any{}},
		"P1": map[string]any{"isa": "PBXFileReference"},
	}

	_, err := parseTarget("T1", objects)
	require.ErrorContains(t, err, "has no path")
}

func TestParseBuildConfiguration_fromFixture(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.TargetByName("XcodeProj")
	require.True(t, ok)
	require.Len(t, target.BuildConfigurations, 2)

	debug := target.BuildConfigurations[0]
	assert.Equal(t, "Debug", debug.Name)
	assert.NotEmpty(t, debug.ID)

	bundleID, ok := debug.BuildSetting("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	assert.Equal(t, "com.bitrise.XcodeProj", bundleID)
}

func TestParseBuildConfiguration_withoutBuildSettings(t *testing.T) {
	objects := serialized.Object{
		"C1": map[string]any{"isa": "XCBuildConfiguration", "name": "Debug"},
	}

	_, err := parseBuildConfiguration("C1", objects)
	require.ErrorContains(t, err, "has no buildSettings")
}

func TestParseConfigurationList_missing(t *testing.T) {
	_, _, err := parseConfigurationList("GONE", serialized.Object{})
	require.ErrorContains(t, err, "GONE")
}

func TestParseConfigurationList_defaultNameNeedNotExist(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.Len(t, project.BuildConfigurations(), 1)
	assert.Equal(t, "Debug", project.BuildConfigurations()[0].Name)
	assert.Equal(t, "Release", project.DefaultConfigurationName())
}

// Regression test for 2f529ef from the abandoned xcodeproject-v2 branch.
func TestFactory_Parse_optionalTargetKeys(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	require.NotContains(t, string(content), "dependencies")
	require.NotContains(t, string(content), "buildPhases")

	project, err := testFactory().newPBXProj(content, "/projects/App.xcodeproj")
	require.NoError(t, err)

	require.Len(t, project.Targets(), 1)
	target := project.Targets()[0]
	assert.Equal(t, "App", target.Name)
	assert.Empty(t, target.dependencyTargetIDs)
	assert.Empty(t, pbxFile(t, project).buildPhaseIDs[target.ID])
}

func TestFactory_Parse_invalidContent(t *testing.T) {
	_, err := testFactory().newPBXProj([]byte("not a plist"), "/projects/App.xcodeproj")
	require.Error(t, err)
}

func TestFactory_Parse_noProjectObject(t *testing.T) {
	content := []byte("{\n\tobjects = {\n\t\tAA1 = {\n\t\t\tisa = PBXGroup;\n\t\t};\n\t};\n}\n")

	_, err := testFactory().newPBXProj(content, "/projects/App.xcodeproj")
	require.ErrorContains(t, err, "PBXProject")
}

func TestFactory_Parse_sharedBuildConfiguration(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	// The target's configuration list references the project's XCBuildConfiguration.
	shared := strings.Replace(string(content), "AA00000000000000000000A2 /* Debug */,\n", "AA00000000000000000000A1 /* Debug */,\n", 1)
	require.NotEqual(t, string(content), shared)

	p, err := testFactory().newPBXProj([]byte(shared), "/projects/App.xcodeproj")
	require.NoError(t, err)

	require.NoError(t, p.SetBuildSetting("App", "Debug", "MARKETING_VERSION", "2.0"))

	value, ok := p.BuildConfigurations()[0].BuildSetting("MARKETING_VERSION")
	require.True(t, ok)
	assert.Equal(t, "2.0", value, "a configuration shared by two lists is one configuration")
}
