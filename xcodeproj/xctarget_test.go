package xcodeproj

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GivenOnlyApp_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTarget := TargetModel{
		Name:       "MyApp",
		ID:         "95B215AF326162D28887FCB1",
		HasXCTest:  false,
		HasAppClip: false,
	}

	// Given
	templateDir := givenGeneratedProject(t, "onlyapp")
	defer clearDir(t, templateDir)

	// When
	targets, err := ProjectTargets(templateDir + "/MyApp.xcodeproj")

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Contains(t, targets, expectedTarget)
}

func Test_GivenAppWithAppClip_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTargets := []TargetModel{
		{
			Name:       "MyApp",
			ID:         "95B215AF326162D28887FCB1",
			HasXCTest:  false,
			HasAppClip: true,
		},
		{
			Name:       "MyAppClip",
			ID:         "6FAA6B396C7410901BE6FA94",
			HasXCTest:  false,
			HasAppClip: false,
		},
	}

	templateDir := givenGeneratedProject(t, "appWithAppClip")
	defer clearDir(t, templateDir)

	// When
	targets, err := ProjectTargets(templateDir + "/MyApp.xcodeproj")

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 2)
	for _, expectedTarget := range expectedTargets {
		assert.Contains(t, targets, expectedTarget)
	}
}

func Test_GivenAppWithTest_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTargets := []TargetModel{
		{
			Name:       "MyApp",
			ID:         "95B215AF326162D28887FCB1",
			HasXCTest:  true,
			HasAppClip: false,
		},
	}

	templateDir := givenGeneratedProject(t, "appWithTest")
	defer clearDir(t, templateDir)

	// When
	targets, err := ProjectTargets(templateDir + "/MyApp.xcodeproj")

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	for _, expectedTarget := range expectedTargets {
		assert.Contains(t, targets, expectedTarget)
	}
}

func Test_GivenAppWithUITest_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTargets := []TargetModel{
		{
			Name:       "MyApp",
			ID:         "95B215AF326162D28887FCB1",
			HasXCTest:  true,
			HasAppClip: false,
		},
	}

	templateDir := givenGeneratedProject(t, "appWithUITest")
	defer clearDir(t, templateDir)

	// When
	targets, err := ProjectTargets(templateDir + "/MyApp.xcodeproj")

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	for _, expectedTarget := range expectedTargets {
		assert.Contains(t, targets, expectedTarget)
	}
}

func Test_GivenAppWithTestAndAppClipAndWidget_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTargets := []TargetModel{
		{
			Name:       "MyApp",
			ID:         "95B215AF326162D28887FCB1",
			HasXCTest:  true,
			HasAppClip: true,
		},
		{
			Name:       "MyAppClip",
			ID:         "6FAA6B396C7410901BE6FA94",
			HasXCTest:  false,
			HasAppClip: false,
		},
		{
			Name:       "MyAppWidget",
			ID:         "499A64C35FF149965A13A41C",
			HasXCTest:  false,
			HasAppClip: false,
		},
	}

	templateDir := givenGeneratedProject(t, "appWithAppClipAndTestAndWidget")
	defer clearDir(t, templateDir)

	// When
	targets, err := ProjectTargets(templateDir + "/MyApp.xcodeproj")

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 3)
	for _, expectedTarget := range expectedTargets {
		assert.Contains(t, targets, expectedTarget)
	}
}

func givenGeneratedProject(t *testing.T, template string) string {
	templateDir, err := filepath.Abs("_test/template/" + template)
	require.NoError(t, err)

	temporaryDir, err := pathutil.NormalizedOSTempDirPath("")
	require.NoError(t, err)

	err = command.CopyDir(templateDir, temporaryDir, false)
	require.NoError(t, err)

	destDir := filepath.Join(temporaryDir, template)

	f := command.NewFactory(env.NewRepository())
	cmd := f.Create("tuist", []string{"generate", "--path", destDir, "--project-only"}, nil)

	output, err := cmd.RunAndReturnTrimmedCombinedOutput()
	fmt.Println(output)
	require.NoError(t, err)

	return destDir
}

func clearDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	require.NoError(t, err)
}
