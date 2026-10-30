// Package xcproj decodes project.xcproj, the JSON project file Xcode 27 writes inside an
// .xcodeproj bundle in place of project.pbxproj.
package xcproj

import (
	"encoding/json"
	"fmt"

	"github.com/tailscale/hujson"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// The target kinds. A target written without a kind is native.
const (
	TargetKindNative              = "native"
	TargetKindAggregate           = "aggregate"
	TargetKindExternalBuildSystem = "external-build-system"
)

const localTargetDependencyKind = "localTarget"

// Project is the part of project.xcproj that go-xcode reads.
type Project struct {
	DefaultConfiguration string
	Configurations       []Configuration

	// BuildSettings holds every configuration in one map: a key carries its conditions, as in
	// CODE_SIGN_STYLE[config=Release].
	BuildSettings serialized.Object

	Targets []Target
}

// Configuration is one build configuration of the project.
type Configuration struct {
	// ID is usually empty: Xcode writes one only when the name alone would be ambiguous.
	ID   string
	Name string
}

// Target is one target of the project.
type Target struct {
	ID   string
	Name string
	Kind string

	// ProductType is the pbxproj product type without its com.apple.product-type. prefix.
	ProductType string
	Product     string

	// TestHostTarget is the name of the app target a test target runs in.
	TestHostTarget string

	// Dependencies are the names of the project's own targets this target depends on.
	// Dependencies on other projects and on Swift packages are left out.
	Dependencies []string

	// BuildSettings is keyed like Project.BuildSettings.
	BuildSettings serialized.Object
}

type rawProject struct {
	DefaultConfiguration string             `json:"default-configuration"`
	Configurations       []rawConfiguration `json:"configurations"`
	BuildSettings        serialized.Object  `json:"build-settings"`
	Targets              []rawTarget        `json:"targets"`
}

// rawConfiguration is written as a bare name unless it has an ID or an xcconfig file.
type rawConfiguration struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type rawTarget struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Kind           string            `json:"kind"`
	ProductType    string            `json:"product-type"`
	Product        string            `json:"product"`
	TestHostTarget string            `json:"test-host-target"`
	Dependencies   []rawDependency   `json:"dependencies"`
	BuildSettings  serialized.Object `json:"build-settings"`
}

// rawDependency is written as a bare target name unless it carries platform filters or points
// outside the project.
type rawDependency struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
}

// Decode parses the content of a project.xcproj file.
func Decode(content []byte) (Project, error) {
	standard, err := hujson.Standardize(content)
	if err != nil {
		return Project{}, fmt.Errorf("failed to parse project.xcproj (Xcode writes JSON with comments and trailing commas; other JSON5 syntax, such as unquoted keys or single-quoted strings, is not supported): %w", err)
	}

	var raw rawProject
	if err := json.Unmarshal(standard, &raw); err != nil {
		return Project{}, fmt.Errorf("failed to decode project.xcproj: %w", err)
	}

	return raw.project()
}

// UnmarshalJSON accepts a bare name or an object.
func (c *rawConfiguration) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.Name); err == nil {
		return nil
	}

	type configurationObject rawConfiguration
	return json.Unmarshal(data, (*configurationObject)(c))
}

// UnmarshalJSON accepts a bare target name or an object.
func (d *rawDependency) UnmarshalJSON(data []byte) error {
	d.Kind = localTargetDependencyKind
	if err := json.Unmarshal(data, &d.Target); err == nil {
		return nil
	}

	type dependencyObject rawDependency
	return json.Unmarshal(data, (*dependencyObject)(d))
}

func (r rawProject) project() (Project, error) {
	configurations := make([]Configuration, 0, len(r.Configurations))
	for i, rawConfig := range r.Configurations {
		if rawConfig.Name == "" {
			return Project{}, fmt.Errorf("configuration %d has no name", i)
		}
		configurations = append(configurations, Configuration(rawConfig))
	}

	targets := make([]Target, 0, len(r.Targets))
	for i, rawTarget := range r.Targets {
		target, err := rawTarget.target()
		if err != nil {
			return Project{}, fmt.Errorf("target %d: %w", i, err)
		}
		targets = append(targets, target)
	}

	return Project{
		DefaultConfiguration: r.DefaultConfiguration,
		Configurations:       configurations,
		BuildSettings:        nonNil(r.BuildSettings),
		Targets:              targets,
	}, nil
}

func (r rawTarget) target() (Target, error) {
	if r.Name == "" {
		return Target{}, fmt.Errorf("target has no name")
	}
	if r.ID == "" {
		return Target{}, fmt.Errorf("target %s has no id", r.Name)
	}

	kind := r.Kind
	switch kind {
	case "":
		kind = TargetKindNative
	case TargetKindNative, TargetKindAggregate, TargetKindExternalBuildSystem:
	default:
		return Target{}, fmt.Errorf("target %s has unknown kind: %s", r.Name, kind)
	}

	var dependencies []string
	for _, dependency := range r.Dependencies {
		if dependency.Kind != localTargetDependencyKind {
			continue
		}
		if dependency.Target == "" {
			return Target{}, fmt.Errorf("target %s has a dependency with no target", r.Name)
		}
		dependencies = append(dependencies, dependency.Target)
	}

	return Target{
		ID:             r.ID,
		Name:           r.Name,
		Kind:           kind,
		ProductType:    r.ProductType,
		Product:        r.Product,
		TestHostTarget: r.TestHostTarget,
		Dependencies:   dependencies,
		BuildSettings:  nonNil(r.BuildSettings),
	}, nil
}

// nonNil keeps an absent settings map from reading differently than an empty one.
func nonNil(settings serialized.Object) serialized.Object {
	if settings == nil {
		return serialized.Object{}
	}
	return settings
}
