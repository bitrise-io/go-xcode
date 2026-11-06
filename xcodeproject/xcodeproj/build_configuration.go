package xcodeproj

import (
	"fmt"
	"maps"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildConfiguration is one named build configuration, such as Debug or Release.
type BuildConfiguration struct {
	ID   string
	Name string

	// buildSettings is the project's in-memory copy, shared by every copy of the Target, so a
	// Target fetched before SetBuildSetting sees the change. Only SetBuildSetting writes it, after
	// the project file accepted the value. It is not the project file's own representation: for
	// project.pbxproj it is a shallow clone, so nested values such as arrays are still shared with
	// the raw tree and must not be modified in place.
	buildSettings serialized.Object
}

// BuildSetting returns the value of key as declared in the project file: not variable-expanded
// and without .xcconfig overrides. For the effective value, use XcodeProj.TargetBuildSettings.
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
		return BuildConfiguration{}, fmt.Errorf("build configuration %s has no buildSettings", id)
	}

	return BuildConfiguration{
		ID:            id,
		Name:          name,
		buildSettings: maps.Clone(buildSettings),
	}, nil
}

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
