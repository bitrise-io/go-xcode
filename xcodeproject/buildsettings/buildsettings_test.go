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
	targets, err := Parse(readFixture(t, "scheme_with_test_target.json"))
	require.NoError(t, err)
	require.Len(t, targets, 2)

	app := targets["ios-simple-objc"]
	require.Equal(t, "Bitrise.ios-simple-objc", app["PRODUCT_BUNDLE_IDENTIFIER"])
	require.Equal(t, "@executable_path/Frameworks", app["LD_RUNPATH_SEARCH_PATHS"], "xcodebuild's padding is trimmed")
	require.NotContains(t, app, "BUNDLE_LOADER", "the test target's own settings stay with the test target")
	require.Equal(t, "Bitrise.ios-simple-objcTests", targets["ios-simple-objcTests"]["PRODUCT_BUNDLE_IDENTIFIER"])

	targets, err = Parse(readFixture(t, "quoted_value.json"))
	require.NoError(t, err)
	require.Equal(t, `-framework "SnapKit"`, targets["sample-apps-ios-workspace-swift"]["OTHER_LDFLAGS"], "quotes inside a value stay")

	targets, err = Parse([]byte("[\n\n]"))
	require.NoError(t, err)
	require.Empty(t, targets)

	_, err = Parse([]byte("Build settings for action build and target App:\n    SDKROOT = iphoneos"))
	require.Error(t, err)
}
