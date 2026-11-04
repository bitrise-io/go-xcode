package xcodeproj

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	bullsEyeAppTargetID      = "D2A5F1FF1F4A9144005CD714"
	bullsEyeUITestTargetName = "BullsEyeUITests"
)

// projectModel is what both project file formats must read the same from one project.
type projectModel struct {
	name                     string
	targets                  []Target
	targetsByName            map[string]Target
	buildConfigurations      []BuildConfiguration
	defaultConfigurationName string
	dependentTargets         map[string][]string
	appIconSetPaths          map[string][]string
}

// testdata/format-parity holds one project saved by Xcode as project.pbxproj and, converted, as
// project.xcproj.
func TestXcodeProj_formatParity(t *testing.T) {
	pbxProj := openBullsEyeFixture(t, pbxProjFileName)
	xcProj := openBullsEyeFixture(t, xcProjFileName)
	require.IsType(t, &PBXProj{}, pbxProj)
	require.IsType(t, &XCProj{}, xcProj)

	want := readProjectModel(t, pbxProj)
	assert.Equal(t, "BullsEye", want.name)
	assert.Equal(t, []string{"BullsEye", "BullsEyeTests", "BullsEyeSlowTests", bullsEyeUITestTargetName, "BullsEyeFailingTests"}, targetNames(want.targets))
	assert.Equal(t, []string{"BullsEye"}, want.dependentTargets[bullsEyeUITestTargetName])
	assert.Equal(t, map[string][]string{
		bullsEyeAppTargetID: {filepath.Join("BullsEye", "Assets.xcassets", "AppIcon.appiconset")},
	}, want.appIconSetPaths)
	assert.Equal(t, want, readProjectModel(t, xcProj))

	for _, project := range []XcodeProj{pbxProj, xcProj} {
		require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("BullsEye", "Release")))
		require.NoError(t, project.ForceCodeSign(forceCodeSignOptions(bullsEyeUITestTargetName, "Debug")))
	}
	requireBuildSetting(t, pbxProj, "BullsEye", "Release", "CODE_SIGN_STYLE", "Manual")
	requireBuildSetting(t, pbxProj, "BullsEye", "Debug", "CODE_SIGN_STYLE", "Automatic")

	want = readProjectModel(t, pbxProj)
	assert.Equal(t, want, readProjectModel(t, xcProj), "after ForceCodeSign")

	for _, project := range []XcodeProj{pbxProj, xcProj} {
		require.NoError(t, project.Save())
		reopened, err := testFactory().Open(project.Path())
		require.NoError(t, err)
		assert.Equal(t, want, readProjectModel(t, reopened), "%T after Save", project)
	}
}

func openBullsEyeFixture(t *testing.T, projectFileName string) XcodeProj {
	t.Helper()

	fixtureDir := filepath.Join("testdata", "format-parity")
	projectDir := t.TempDir()
	require.NoError(t, os.CopyFS(filepath.Join(projectDir, "BullsEye"), os.DirFS(filepath.Join(fixtureDir, "BullsEye"))))

	content, err := os.ReadFile(filepath.Join(fixtureDir, projectFileName))
	require.NoError(t, err)
	projectPath := filepath.Join(projectDir, "BullsEye.xcodeproj")
	require.NoError(t, os.Mkdir(projectPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, projectFileName), content, 0644))

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	return project
}

func readProjectModel(t *testing.T, project XcodeProj) projectModel {
	t.Helper()

	model := projectModel{
		name:                     project.Name(),
		buildConfigurations:      withoutConfigurationIDs(project.BuildConfigurations()),
		defaultConfigurationName: project.DefaultConfigurationName(),
		targetsByName:            map[string]Target{},
		dependentTargets:         map[string][]string{},
		appIconSetPaths:          map[string][]string{},
	}

	for _, target := range project.Targets() {
		target.BuildConfigurations = withoutConfigurationIDs(target.BuildConfigurations)
		model.targets = append(model.targets, target)

		byName, ok := project.TargetByName(target.Name)
		require.True(t, ok, target.Name)
		byName.BuildConfigurations = withoutConfigurationIDs(byName.BuildConfigurations)
		model.targetsByName[target.Name] = byName

		model.dependentTargets[target.Name] = targetNames(project.DependentTargetsOfTarget(target))
	}

	appIconSetPaths, err := project.AppIconSetPaths()
	require.NoError(t, err)
	projectDir := filepath.Dir(project.Path())
	for targetID, paths := range appIconSetPaths {
		for _, pth := range paths {
			relativePath, err := filepath.Rel(projectDir, pth)
			require.NoError(t, err)
			model.appIconSetPaths[targetID] = append(model.appIconSetPaths[targetID], relativePath)
		}
	}

	return model
}

// withoutConfigurationIDs clears the configuration IDs, which project.xcproj does not always record.
func withoutConfigurationIDs(configurations []BuildConfiguration) []BuildConfiguration {
	configurations = slices.Clone(configurations)
	for i := range configurations {
		configurations[i].ID = ""
	}
	return configurations
}
