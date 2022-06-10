package xcscheme

import (
	"encoding/xml"
	"testing"

	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/stretchr/testify/require"
)

func TestOpenScheme(t *testing.T) {
	pth := testhelper.CreateTmpFile(t, "ios-simple-objc.xcscheme", schemeContent)
	scheme, err := Open(pth)
	require.NoError(t, err)

	require.Equal(t, "ios-simple-objc", scheme.Name)
	require.Equal(t, pth, scheme.Path)

	require.Equal(t, "Release", scheme.ArchiveAction.BuildConfiguration)
	require.Equal(t, 2, len(scheme.BuildAction.BuildActionEntries))

	{
		entry := scheme.BuildAction.BuildActionEntries[0]
		require.Equal(t, "YES", entry.BuildForArchiving)
		require.Equal(t, "YES", entry.BuildForTesting)
		require.Equal(t, "BA3CBE7419F7A93800CED4D5", entry.BuildableReference.BlueprintIdentifier)

		pth, err := entry.BuildableReference.ReferencedContainerAbsPath("/project.xcodeproj")
		require.NoError(t, err)
		require.Equal(t, "/project.xcodeproj/ios-simple-objc.xcodeproj", pth)
	}

	{
		entry := scheme.BuildAction.BuildActionEntries[1]
		require.Equal(t, "NO", entry.BuildForArchiving)
		require.Equal(t, "YES", entry.BuildForTesting)
		require.Equal(t, "BA3CBE9019F7A93900CED4D5", entry.BuildableReference.BlueprintIdentifier)
	}
}

func TestAppBuildActionEntry(t *testing.T) {
	var scheme Scheme
	require.NoError(t, xml.Unmarshal([]byte(schemeContent), &scheme))

	entry, ok := scheme.AppBuildActionEntry()
	require.True(t, ok)

	require.Equal(t, "YES", entry.BuildForArchiving)
	require.Equal(t, "YES", entry.BuildForTesting)
	require.Equal(t, "BA3CBE7419F7A93800CED4D5", entry.BuildableReference.BlueprintIdentifier)
	require.Equal(t, "ios-simple-objc.app", entry.BuildableReference.BuildableName)
	require.Equal(t, "ios-simple-objc", entry.BuildableReference.BlueprintName)
	require.Equal(t, "container:ios-simple-objc.xcodeproj", entry.BuildableReference.ReferencedContainer)

	require.True(t, entry.BuildableReference.IsAppReference())
}

func TestAppTestActionEntry(t *testing.T) {
	var scheme Scheme
	require.NoError(t, xml.Unmarshal([]byte(schemeContent), &scheme))

	require.Equal(t, "Debug", scheme.TestAction.BuildConfiguration)
	require.Equal(t, 2, len(scheme.TestAction.Testables))
	require.Equal(t, "NO", scheme.TestAction.Testables[0].Skipped)
	require.Equal(t, "YES", scheme.TestAction.Testables[1].Skipped)
	require.Equal(t, "BA3CBE9019F7A93900CED4D5", scheme.TestAction.Testables[0].BuildableReference.BlueprintIdentifier)

	require.False(t, scheme.TestAction.Testables[0].BuildableReference.IsAppReference())
	require.False(t, scheme.TestAction.Testables[1].BuildableReference.IsAppReference())
}

func TestScheme_Marshal(t *testing.T) {
	pth := testhelper.CreateTmpFile(t, "ios-simple-objc.xcscheme", schemeContent)
	scheme, err := Open(pth)
	require.NoError(t, err)

	content, err := scheme.Marshal()
	require.NoError(t, err)

	require.Equal(t, schemeContent, string(content))
}

func TestGivenSchemeWithTestPlans_WhenOpen_ThenTestPlanPropertiesAreParsed(t *testing.T) {
	// Given
	pth := testhelper.CreateTmpFile(t, "BullsEye.xcscheme", schemeWithTestPlanContent)

	// When
	scheme, err := Open(pth)

	// Then
	require.NoError(t, err)

	testPlan := scheme.DefaultTestPlan()
	require.NotNil(t, testPlan)
	require.Equal(t, "FullTests", testPlan.Name())
}

func TestGivenSchemeWithoutTestPlans_WhenOpen_ThenTestPlanPropertiesAreEmpty(t *testing.T) {
	// Given
	pth := testhelper.CreateTmpFile(t, "ios-simple-objc.xcscheme", schemeContent)

	// When
	scheme, err := Open(pth)

	// Then
	require.NoError(t, err)

	require.Nil(t, scheme.TestAction.TestPlans)
	testPlan := scheme.DefaultTestPlan()
	require.Nil(t, testPlan)
}
