package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Target and configuration identifiers from testdata/ios-sample.pbxproj.
const (
	appTargetID          = "7D5B35FB20E28EE80022BAE6"
	uiTestTargetID       = "7D0342F020F4BA280050B6A6"
	todayExtensionTarget = "7D03430C20F4BB070050B6A6"
)

func testFactory() Factory {
	return NewFactory(
		log.NewLogger(),
		nil, // build settings are not needed by the parse-and-read surface
		fileutil.NewFileManager(),
		pathutil.NewPathModifier(),
		nil, // path provider is only needed for app icon lookup
		NewUserProvider(),
	)
}

func parseFixture(t *testing.T, name string) *XcodeProj {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	project, err := testFactory().Parse(content, "/projects/XcodeProj.xcodeproj")
	require.NoError(t, err)

	return project
}

func TestIsXcodeProj(t *testing.T) {
	assert.True(t, IsXcodeProj("/a/b/App.xcodeproj"))
	assert.False(t, IsXcodeProj("/a/b/App.xcworkspace"))
	assert.False(t, IsXcodeProj("/a/b/App"))
}

func TestFactory_Parse(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	assert.Equal(t, "XcodeProj", project.Name)
	assert.Equal(t, "/projects/XcodeProj.xcodeproj", project.Path)

	var targetNames []string
	for _, target := range project.Targets() {
		targetNames = append(targetNames, target.Name)
	}
	assert.Equal(t, []string{"XcodeProj", "XcodeProjUITests", "TodayExtension"}, targetNames)

	require.Len(t, project.BuildConfigurations(), 2)
	assert.Equal(t, "Debug", project.BuildConfigurations()[0].Name)
	assert.Equal(t, "Release", project.BuildConfigurations()[1].Name)
	assert.Equal(t, "Release", project.DefaultConfigurationName())
}

// A target that declares neither dependencies nor buildPhases must parse: both keys are omitted
// entirely by Xcode when empty. Regression test for the fix carried over from the abandoned
// xcodeproject-v2 branch (2f529ef).
func TestFactory_Parse_optionalTargetKeys(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	require.NotContains(t, string(content), "dependencies")
	require.NotContains(t, string(content), "buildPhases")

	project, err := testFactory().Parse(content, "/projects/App.xcodeproj")
	require.NoError(t, err)

	require.Len(t, project.Targets(), 1)
	target := project.Targets()[0]
	assert.Equal(t, "App", target.Name)
	assert.Empty(t, target.dependencyTargetIDs)
	assert.Empty(t, target.buildPhaseIDs)
}

func TestFactory_Parse_invalidContent(t *testing.T) {
	_, err := testFactory().Parse([]byte("not a plist"), "/projects/App.xcodeproj")
	require.Error(t, err)
}

func TestFactory_Parse_noProjectObject(t *testing.T) {
	// A structurally valid plist with an objects dictionary that contains no PBXProject.
	content := []byte("{\n\tobjects = {\n\t\tAA1 = {\n\t\t\tisa = PBXGroup;\n\t\t};\n\t};\n}\n")

	_, err := testFactory().Parse(content, "/projects/App.xcodeproj")
	require.ErrorContains(t, err, "PBXProject")
}

func TestFactory_Open(t *testing.T) {
	// Open is a thin wrapper over path resolution and file reading, so it is exercised against a
	// real directory: a faked FileManager would have to produce a real *os.File anyway.
	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "project.pbxproj"), content, 0644))

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	assert.Equal(t, "App", project.Name)
	assert.Equal(t, projectPath, project.Path)
	require.Len(t, project.Targets(), 1)
}

func TestFactory_Open_missingProject(t *testing.T) {
	_, err := testFactory().Open(filepath.Join(t.TempDir(), "Missing.xcodeproj"))
	require.Error(t, err)
}

func TestXcodeProj_Target(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.Target(appTargetID)
	require.True(t, ok)
	assert.Equal(t, "XcodeProj", target.Name)

	_, ok = project.Target("NOPE")
	assert.False(t, ok)
}

func TestXcodeProj_TargetByName(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.TargetByName("TodayExtension")
	require.True(t, ok)
	assert.Equal(t, todayExtensionTarget, target.ID)

	_, ok = project.TargetByName("Nope")
	assert.False(t, ok)
}

func TestXcodeProj_DependentTargetsOfTarget(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	app, ok := project.Target(appTargetID)
	require.True(t, ok)
	assert.Equal(t, []string{"TodayExtension"}, targetNames(project.DependentTargetsOfTarget(app)))

	// The UI test target depends on the app, which in turn depends on the extension: the
	// traversal is transitive.
	uiTests, ok := project.Target(uiTestTargetID)
	require.True(t, ok)
	assert.Equal(t, []string{"XcodeProj", "TodayExtension"}, targetNames(project.DependentTargetsOfTarget(uiTests)))

	extension, ok := project.Target(todayExtensionTarget)
	require.True(t, ok)
	assert.Empty(t, project.DependentTargetsOfTarget(extension))
}

func TestXcodeProj_DependentTargetsOfTarget_unresolvableDependencyIsSkipped(t *testing.T) {
	project := &XcodeProj{
		logger: log.NewLogger(),
		targets: []Target{
			{ID: "A", Name: "A", dependencyTargetIDs: []string{"B", "GONE"}},
			{ID: "B", Name: "B"},
		},
	}

	assert.Equal(t, []string{"B"}, targetNames(project.DependentTargetsOfTarget(project.targets[0])))
}

// A dependency cycle must terminate rather than recurse forever. Xcode does not create one, but a
// generated or hand-edited project file can.
func TestXcodeProj_DependentTargetsOfTarget_cycle(t *testing.T) {
	project := &XcodeProj{
		logger: log.NewLogger(),
		targets: []Target{
			{ID: "A", Name: "A", dependencyTargetIDs: []string{"B"}},
			{ID: "B", Name: "B", dependencyTargetIDs: []string{"C"}},
			{ID: "C", Name: "C", dependencyTargetIDs: []string{"A"}},
		},
	}

	// A is the target asked about, so it must not appear as its own dependency.
	assert.Equal(t, []string{"B", "C"}, targetNames(project.DependentTargetsOfTarget(project.targets[0])))
}

// A dependency reachable through two paths appears once, where it is first reached.
func TestXcodeProj_DependentTargetsOfTarget_sharedDependency(t *testing.T) {
	project := &XcodeProj{
		logger: log.NewLogger(),
		targets: []Target{
			{ID: "A", Name: "A", dependencyTargetIDs: []string{"B", "C"}},
			{ID: "B", Name: "B", dependencyTargetIDs: []string{"D"}},
			{ID: "C", Name: "C", dependencyTargetIDs: []string{"D"}},
			{ID: "D", Name: "D", dependencyTargetIDs: []string{"E"}},
			{ID: "E", Name: "E"},
		},
	}

	assert.Equal(t, []string{"B", "D", "E", "C"}, targetNames(project.DependentTargetsOfTarget(project.targets[0])))
}

func TestXcodeProj_TargetDevelopmentTeam(t *testing.T) {
	t.Run("target with a development team", func(t *testing.T) {
		project := parseFixture(t, "minimal.pbxproj")

		team, ok := project.TargetDevelopmentTeam("AA00000000000000000000E1")
		require.True(t, ok)
		assert.Equal(t, "TEAM123456", team)
	})

	t.Run("target has attributes but no development team", func(t *testing.T) {
		project := parseFixture(t, "ios-sample.pbxproj")

		_, ok := project.TargetDevelopmentTeam(uiTestTargetID)
		assert.False(t, ok)
	})

	t.Run("target absent from TargetAttributes", func(t *testing.T) {
		project := parseFixture(t, "without-target-attributes.pbxproj")

		_, ok := project.TargetDevelopmentTeam("13BD6332256BE7BF00F72361")
		assert.False(t, ok)
	})

	t.Run("project without TargetAttributes at all", func(t *testing.T) {
		project := parseFixture(t, "minimal.pbxproj")
		project.targetAttributes = nil

		_, ok := project.TargetDevelopmentTeam("AA00000000000000000000E1")
		assert.False(t, ok)
	})
}

func targetNames(targets []Target) []string {
	var names []string
	for _, target := range targets {
		names = append(names, target.Name)
	}
	return names
}
