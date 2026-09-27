package buildsettings_tests

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/buildsettings"
	"github.com/stretchr/testify/require"
)

// Runs the real xcodebuild, which the unit tests only mock.
func TestReader_simpleObjc(t *testing.T) {
	projectPath := filepath.Join(_integration_tests.GetRepository(t, "https://github.com/bitrise-io/sample-apps-ios-simple-objc.git", "master"), "ios-simple-objc", "ios-simple-objc.xcodeproj")
	reader := buildsettings.NewReader(command.NewFactory(env.NewRepository()), log.NewLogger())

	list, err := reader.Read(buildsettings.Query{ProjectPath: projectPath, Target: "ios-simple-objc", Configuration: "Release"})
	require.NoError(t, err)
	main, ok := list.Main()
	require.True(t, ok)
	require.Equal(t, "ios-simple-objc", main.Target)
	require.Equal(t, "Bitrise.ios-simple-objc", main.Values["PRODUCT_BUNDLE_IDENTIFIER"])

	list, err = reader.Read(buildsettings.Query{ProjectPath: projectPath, Scheme: "ios-simple-objc"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list), 2, "the scheme builds its test target too")
	main, _ = list.Main()
	require.Equal(t, "ios-simple-objc", main.Target, "the scheme's main target comes first")
	_, ok = list.Target("ios-simple-objcTests")
	require.True(t, ok)
}
