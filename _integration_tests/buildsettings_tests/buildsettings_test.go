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

	settings, err := reader.Target(projectPath, "ios-simple-objc", "Release")
	require.NoError(t, err)
	require.Equal(t, "Bitrise.ios-simple-objc", settings["PRODUCT_BUNDLE_IDENTIFIER"])

	settings, err = reader.SchemeTarget(projectPath, "ios-simple-objc", "ios-simple-objcTests", "Debug")
	require.NoError(t, err)
	require.Equal(t, "Bitrise.ios-simple-objcTests", settings["PRODUCT_BUNDLE_IDENTIFIER"])

	targets, err := reader.Scheme(projectPath, "ios-simple-objc", "Debug")
	require.NoError(t, err)
	require.Contains(t, targets, "ios-simple-objc")
	require.Contains(t, targets, "ios-simple-objcTests", "the scheme builds its test target too")
}
