package xcodeproj

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
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
		plistutil.NewFileHandler(fileutil.NewFileManager()),
		pathutil.NewPathModifier(),
		pathutil.NewPathProvider(),
		NewUserProvider(),
		NewDirCreator(),
	)
}

func openProjectFiles(t *testing.T, files map[string][]byte) *Project {
	t.Helper()

	projectPath := writeProjectFiles(t, files)
	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	return project
}

func writeProjectFiles(t *testing.T, files map[string][]byte) string {
	t.Helper()

	// Real directory: a faked FileManager would have to return a real *os.File anyway.
	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(projectPath, name), content, 0644))
	}

	return projectPath
}

func parseFixture(t *testing.T, name string) *Project {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	project, err := testFactory().newPBXProj(content, "/projects/XcodeProj.xcodeproj")
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

	assert.Equal(t, "XcodeProj", project.Name())
	assert.Equal(t, "/projects/XcodeProj.xcodeproj", project.Path())

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

func TestFactory_Open(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	projectPath := writeProjectFiles(t, map[string][]byte{pbxProjFileName: content})

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	assert.Equal(t, "App", project.Name())
	assert.Equal(t, projectPath, project.Path())
	require.Len(t, project.Targets(), 1)
}

func TestFactory_Open_bothProjectFiles(t *testing.T) {
	pbxProj, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	projectPath := writeProjectFiles(t, map[string][]byte{
		pbxProjFileName: pbxProj,
		xcProjFileName:  []byte("not read"),
	})

	project, err := testFactory().Open(projectPath)
	require.ErrorContains(t, err, "found both project.pbxproj and project.xcproj")
	assert.True(t, project == nil, "got a non-nil %T", project)
}

func TestFactory_Open_missingProject(t *testing.T) {
	_, err := testFactory().Open(filepath.Join(t.TempDir(), "Missing.xcodeproj"))
	require.ErrorContains(t, err, "found neither project.pbxproj nor project.xcproj")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestFactory_Open_unreadableProjectFile(t *testing.T) {
	projectPath := writeProjectFiles(t, map[string][]byte{})
	require.NoError(t, os.Mkdir(filepath.Join(projectPath, pbxProjFileName), 0755))

	project, err := testFactory().Open(projectPath)
	require.ErrorContains(t, err, pbxProjFileName)
	assert.NotErrorIs(t, err, fs.ErrNotExist)
	assert.True(t, project == nil, "got a non-nil %T", project)
}

func TestFactory_Open_invalidProjectReturnsNilProject(t *testing.T) {
	projectPath := writeProjectFiles(t, map[string][]byte{pbxProjFileName: []byte("{}")})

	project, err := testFactory().Open(projectPath)
	require.Error(t, err)
	// Not assert.Nil: it also accepts an interface holding a nil pointer, which is what this guards.
	assert.True(t, project == nil, "got a non-nil %T", project)
}

func TestProject_Target(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.Target(appTargetID)
	require.True(t, ok)
	assert.Equal(t, "XcodeProj", target.Name)

	_, ok = project.Target("NOPE")
	assert.False(t, ok)
}

func TestProject_TargetByName(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.TargetByName("TodayExtension")
	require.True(t, ok)
	assert.Equal(t, todayExtensionTarget, target.ID)

	_, ok = project.TargetByName("Nope")
	assert.False(t, ok)
}

func TestProject_DependentTargetsOfTarget(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	app, ok := project.Target(appTargetID)
	require.True(t, ok)
	assert.Equal(t, []string{"TodayExtension"}, targetNames(project.DependentTargetsOfTarget(app)))

	uiTests, ok := project.Target(uiTestTargetID)
	require.True(t, ok)
	assert.Equal(t, []string{"XcodeProj", "TodayExtension"}, targetNames(project.DependentTargetsOfTarget(uiTests)))

	extension, ok := project.Target(todayExtensionTarget)
	require.True(t, ok)
	assert.Empty(t, project.DependentTargetsOfTarget(extension))
}

func TestProject_DependentTargetsOfTarget_unresolvableDependencyIsSkipped(t *testing.T) {
	project := &Project{
		logger: log.NewLogger(),
		targets: []Target{
			{ID: "A", Name: "A", dependencyTargetIDs: []string{"B", "GONE"}},
			{ID: "B", Name: "B"},
		},
	}

	assert.Equal(t, []string{"B"}, targetNames(project.DependentTargetsOfTarget(project.targets[0])))
}

func TestProject_DependentTargetsOfTarget_cycle(t *testing.T) {
	project := &Project{
		logger: log.NewLogger(),
		targets: []Target{
			{ID: "A", Name: "A", dependencyTargetIDs: []string{"B"}},
			{ID: "B", Name: "B", dependencyTargetIDs: []string{"C"}},
			{ID: "C", Name: "C", dependencyTargetIDs: []string{"A"}},
		},
	}

	assert.Equal(t, []string{"B", "C"}, targetNames(project.DependentTargetsOfTarget(project.targets[0])))
}

func TestProject_DependentTargetsOfTarget_sharedDependency(t *testing.T) {
	project := &Project{
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

func targetNames(targets []Target) []string {
	var names []string
	for _, target := range targets {
		names = append(names, target.Name)
	}
	return names
}

func projectFileOf(p *Project) projectFile {
	return p.file
}
