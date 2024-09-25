package xcscheme

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GivenScheme_WhenMarshal_ThenContentRemain(t *testing.T) {
	// Given
	schemePth := "testdata/ios-simple-objc.xcscheme"

	f, err := os.Open(schemePth)
	require.NoError(t, err)

	scheme, err := parse(f)
	require.NoError(t, err)

	// When
	content, err := scheme.Marshal()

	// Then
	require.NoError(t, err)

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	schemeContent, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, string(schemeContent), string(content))
}

func Test_GivenSchemeWithTestPlan_WhenOpen_ThenDefaultTestPlanSet(t *testing.T) {
	// Given
	schemePth := "testdata/BullsEye.xcscheme"

	// When
	scheme, err := Open(schemePth)

	// Then
	require.NoError(t, err)

	testPlan := scheme.DefaultTestPlan()
	require.NotNil(t, testPlan)
	require.Equal(t, "FullTests", testPlan.Name())
}

func Test_GivenSimpleScheme_WhenOpen(t *testing.T) {
	// Given
	schemePth := "testdata/ios-simple-objc.xcscheme"

	// When
	scheme, err := Open(schemePth)

	// Then
	require.NoError(t, err)

	require.Equal(t, "ios-simple-objc", scheme.Name)
	require.Equal(t, schemePth, scheme.Path)

	require.Equal(t, "Release", scheme.ArchiveAction.BuildConfiguration)
	require.Equal(t, 2, len(scheme.BuildAction.BuildActionEntries))

	assertIosSimpleObjCSchemeAppBuildActionEntry(scheme, t)
	assertIosSimpleObjCSchemeTestActionEntry(scheme, t)
}

func assertIosSimpleObjCSchemeAppBuildActionEntry(scheme Scheme, t *testing.T) {
	entry, ok := scheme.AppBuildActionEntry()
	require.True(t, ok)

	require.Equal(t, "YES", entry.BuildForArchiving)
	require.Equal(t, "YES", entry.BuildForTesting)
	require.Equal(t, "BA3CBE7419F7A93800CED4D5", entry.BuildableReference.BlueprintIdentifier)
	require.Equal(t, "ios-simple-objc.app", entry.BuildableReference.BuildableName)
	require.Equal(t, "ios-simple-objc", entry.BuildableReference.BlueprintName)
	require.Equal(t, "container:ios-simple-objc.xcodeproj", entry.BuildableReference.ReferencedContainer)

	pth, err := entry.BuildableReference.ReferencedContainerAbsPath("/project.xcodeproj")
	require.NoError(t, err)
	require.Equal(t, "/project.xcodeproj/ios-simple-objc.xcodeproj", pth)

	require.True(t, entry.BuildableReference.IsAppReference())
}

func assertIosSimpleObjCSchemeTestActionEntry(scheme Scheme, t *testing.T) {
	require.Equal(t, "Debug", scheme.TestAction.BuildConfiguration)
	require.Equal(t, 1, len(scheme.TestAction.Testables))
	require.Equal(t, "NO", scheme.TestAction.Testables[0].Skipped)
	require.Equal(t, "BA3CBE9019F7A93900CED4D5", scheme.TestAction.Testables[0].BuildableReference.BlueprintIdentifier)

	require.False(t, scheme.TestAction.Testables[0].BuildableReference.IsAppReference())

	require.Nil(t, scheme.TestAction.TestPlans)
	testPlan := scheme.DefaultTestPlan()
	require.Nil(t, testPlan)
}
