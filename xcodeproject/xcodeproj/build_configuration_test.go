package xcodeproj

import (
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfiguration_BuildSetting(t *testing.T) {
	configuration := BuildConfiguration{
		ID:   "C1",
		Name: "Debug",
		buildSettings: serialized.Object{
			"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.App",
			"INFOPLIST_FILE":            "$(SRCROOT)/App/Info.plist",
			"SUPPORTED_PLATFORMS":       []any{"iphoneos", "iphonesimulator"},
		},
	}

	t.Run("string value", func(t *testing.T) {
		value, ok := configuration.BuildSetting("PRODUCT_BUNDLE_IDENTIFIER")
		require.True(t, ok)
		assert.Equal(t, "io.bitrise.App", value)
	})

	t.Run("declared values are not variable-expanded", func(t *testing.T) {
		value, ok := configuration.BuildSetting("INFOPLIST_FILE")
		require.True(t, ok)
		assert.Equal(t, "$(SRCROOT)/App/Info.plist", value)
	})

	t.Run("absent key", func(t *testing.T) {
		_, ok := configuration.BuildSetting("NOPE")
		assert.False(t, ok)
	})

	t.Run("value that is not a string", func(t *testing.T) {
		_, ok := configuration.BuildSetting("SUPPORTED_PLATFORMS")
		assert.False(t, ok)
	})

	t.Run("configuration with no build settings", func(t *testing.T) {
		_, ok := BuildConfiguration{}.BuildSetting("ANY")
		assert.False(t, ok)
	})
}

func TestParseBuildConfiguration_fromFixture(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.TargetByName("XcodeProj")
	require.True(t, ok)
	require.Len(t, target.BuildConfigurations, 2)

	debug := target.BuildConfigurations[0]
	assert.Equal(t, "Debug", debug.Name)
	assert.NotEmpty(t, debug.ID)

	bundleID, ok := debug.BuildSetting("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	assert.Equal(t, "com.bitrise.XcodeProj", bundleID)
}

func TestParseConfigurationList_missing(t *testing.T) {
	_, _, err := parseConfigurationList("GONE", serialized.Object{})
	require.ErrorContains(t, err, "GONE")
}

func TestParseConfigurationList_defaultNameNeedNotExist(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.Len(t, project.BuildConfigurations(), 1)
	assert.Equal(t, "Debug", project.BuildConfigurations()[0].Name)
	assert.Equal(t, "Release", project.DefaultConfigurationName())
}
