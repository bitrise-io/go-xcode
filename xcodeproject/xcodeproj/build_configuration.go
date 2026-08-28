package xcodeproj

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildConfiguration is one named build configuration, such as Debug or Release.
type BuildConfiguration struct {
	ID   string
	Name string

	// buildSettings is the configuration's XCBuildConfiguration.buildSettings node. It is
	// intentionally unexported: build settings are written through XcodeProj.SetBuildSetting so
	// that the change reaches the raw project tree that Save writes out.
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

	buildSettings, ok := raw.Object("buildSettings")
	if !ok {
		// A configuration without any build settings is unusual but valid; treat it as empty
		// rather than failing the whole parse.
		buildSettings = serialized.Object{}
	}

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
