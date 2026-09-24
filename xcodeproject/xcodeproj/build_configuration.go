package xcodeproj

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildConfiguration is one named build configuration, such as Debug or Release.
type BuildConfiguration struct {
	ID   string
	Name string

	// buildSettings is the configuration's XCBuildConfiguration.buildSettings node: the same map
	// that lives in the raw project tree, not a copy, exactly as in v1. Writes through
	// XcodeProj.SetBuildSetting and XcodeProj.ForceCodeSign land in it, which is how they reach
	// Save. It is nil when the project file declares no buildSettings for the configuration.
	buildSettings serialized.Object
}

// BuildSetting returns the value of key as it is DECLARED in the project file.
//
// This is not the effective value Xcode would build with: it is not variable-expanded, so it may
// contain $(VAR) references, and an .xcconfig file may override it without that being visible
// here. Reading it costs nothing, which is why it is worth having. For the effective value, use
// XcodeProj.TargetBuildSettings, which asks xcodebuild.
//
// ok is false if the key is absent or its value is not a string.
func (c BuildConfiguration) BuildSetting(key string) (string, bool) {
	return c.buildSettings.String(key)
}

func parseBuildConfiguration(id string, objects serialized.Object) (BuildConfiguration, error) {
	raw, ok := objects.Object(id)
	if !ok {
		return BuildConfiguration{}, fmt.Errorf("build configuration %s not found", id)
	}

	name, ok := raw.String("name")
	if !ok {
		return BuildConfiguration{}, fmt.Errorf("build configuration %s has no name", id)
	}

	// A configuration without build settings is unusual but still readable, so it does not fail the
	// parse. The map is left nil rather than replaced by an empty one: an empty map would not be
	// part of the raw tree, and a write into it would silently never reach Save.
	buildSettings, _ := raw.Object("buildSettings")

	return BuildConfiguration{
		ID:            id,
		Name:          name,
		buildSettings: buildSettings,
	}, nil
}

// parseConfigurationList resolves an XCConfigurationList into its configurations and the name of
// the default one. The default name is optional and is empty when the list does not declare it.
func parseConfigurationList(id string, objects serialized.Object) (configurations []BuildConfiguration, defaultName string, err error) {
	raw, ok := objects.Object(id)
	if !ok {
		return nil, "", fmt.Errorf("build configuration list %s not found", id)
	}

	ids, ok := raw.StringSlice("buildConfigurations")
	if !ok {
		return nil, "", fmt.Errorf("build configuration list %s has no buildConfigurations", id)
	}

	for _, configurationID := range ids {
		configuration, err := parseBuildConfiguration(configurationID, objects)
		if err != nil {
			return nil, "", err
		}
		configurations = append(configurations, configuration)
	}

	defaultName, _ = raw.String("defaultConfigurationName")

	return configurations, defaultName, nil
}
