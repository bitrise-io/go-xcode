package buildsettings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	out, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return out
}

// The fixtures are reduced from real Xcode 26.5 output: a scheme with its test target
// (sample-apps-ios-simple-objc) and a CocoaPods workspace (sample-apps-ios-workspace-swift).
func TestParse(t *testing.T) {
	list, err := Parse(readFixture(t, "scheme_with_test_target.json"))
	require.NoError(t, err)
	require.Len(t, list, 2)

	main, ok := list.Main()
	require.True(t, ok)
	require.Equal(t, "ios-simple-objc", main.Target)
	require.Equal(t, "build", main.Action)
	bundleID, ok := main.Value("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	require.Equal(t, "Bitrise.ios-simple-objc", bundleID)
	require.Equal(t, "@executable_path/Frameworks", main.Values["LD_RUNPATH_SEARCH_PATHS"], "xcodebuild's padding is trimmed")
	_, ok = main.Value("BUNDLE_LOADER")
	require.False(t, ok, "the test target's own settings stay with the test target")

	tests, ok := list.Target("ios-simple-objcTests")
	require.True(t, ok)
	require.Equal(t, "Bitrise.ios-simple-objcTests", tests.Values["PRODUCT_BUNDLE_IDENTIFIER"])
	_, ok = list.Target("Nope")
	require.False(t, ok)

	list, err = Parse(readFixture(t, "quoted_value.json"))
	require.NoError(t, err)
	require.Equal(t, `-framework "SnapKit"`, list[0].Values["OTHER_LDFLAGS"], "quotes inside a value stay")
}

func TestParse_noTargets(t *testing.T) {
	// A Swift package scheme lists no targets.
	list, err := Parse([]byte("[\n\n]"))
	require.NoError(t, err)
	_, ok := list.Main()
	require.False(t, ok)
}

func TestParse_notJSON(t *testing.T) {
	_, err := Parse([]byte("Build settings for action build and target App:\n    SDKROOT = iphoneos"))
	require.Error(t, err)
}
