package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GivenOnlyApp_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTarget := TargetModel{
		Name:       "test",
		ID:         "64E1835F2588FD3C00D666BF",
		HasXCTest:  false,
		HasAppClip: false,
	}

	projectPath := givenPBXProjWithContent(t, onlyApp)

	// When
	targets, err := ProjectTargets(projectPath)

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Contains(t, targets, expectedTarget)
}

func Test_GivenAppWithAppClip_WhenProjectTargetCalled_ThenExpectSingleTarget(t *testing.T) {
	// Given
	expectedTargets := []TargetModel{
		{
			Name:       "test",
			ID:         "64E1835F2588FD3C00D666BF",
			HasXCTest:  false,
			HasAppClip: true,
		},
		{
			Name:       "clip",
			ID:         "64E1839B2588FD5E00D666BF",
			HasXCTest:  false,
			HasAppClip: false,
		},
	}

	projectPath := givenPBXProjWithContent(t, appWithAppClip)

	// When
	targets, err := ProjectTargets(projectPath)

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
			Name:       "test",
			ID:         "64E1835F2588FD3C00D666BF",
			HasXCTest:  true,
			HasAppClip: false,
		},
	}

	projectPath := givenPBXProjWithContent(t, appWithTest)

	// When
	targets, err := ProjectTargets(projectPath)

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
			Name:       "test",
			ID:         "64E1835F2588FD3C00D666BF",
			HasXCTest:  true,
			HasAppClip: true,
		},
		{
			Name:       "clip",
			ID:         "64E1839B2588FD5E00D666BF",
			HasXCTest:  false,
			HasAppClip: false,
		},
		{
			Name:       "widgetExtension",
			ID:         "64E183DC2588FD9D00D666BF",
			HasXCTest:  false,
			HasAppClip: false,
		},
	}

	projectPath := givenPBXProjWithContent(t, appWithTestAndAppClipAndWidget)

	// When
	targets, err := ProjectTargets(projectPath)

	// Then
	require.NoError(t, err)
	assert.Len(t, targets, 3)
	for _, expectedTarget := range expectedTargets {
		assert.Contains(t, targets, expectedTarget)
	}
}

func givenPBXProjWithContent(t *testing.T, content string) string {
	tempDir, err := pathutil.NormalizedOSTempDirPath("__bitrise_init__")
	require.NoError(t, err)

	// Create xcodeproj
	xcodeproj := filepath.Join(tempDir, "test.xcodeproj")
	err = os.MkdirAll(xcodeproj, os.ModePerm)
	require.NoError(t, err)

	// Create pbxproj
	pbxProjPth := filepath.Join(xcodeproj, "project.pbxproj")
	err = fileutil.WriteStringToFile(pbxProjPth, content)
	require.NoError(t, err)

	return xcodeproj
}
