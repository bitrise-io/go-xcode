package xcodeproj

import (
	"fmt"
	"maps"
	"path/filepath"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

const pbxProjectISA = "PBXProject"

func parsePBXProj(content []byte, base Project) (*Project, error) {
	rawProj, format, err := decodePBXProj(content)
	if err != nil {
		return nil, err
	}

	objects, ok := rawProj.Object("objects")
	if !ok {
		return nil, fmt.Errorf("failed to parse project.pbxproj: no objects")
	}

	projectID, rawProject, err := findProject(objects)
	if err != nil {
		return nil, err
	}

	buildConfigurationListID, ok := rawProject.String("buildConfigurationList")
	if !ok {
		return nil, fmt.Errorf("project %s has no buildConfigurationList", projectID)
	}

	buildConfigurations, defaultConfigurationName, err := parseConfigurationList(buildConfigurationListID, objects)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project build configuration list: %w", err)
	}

	targets, buildPhaseIDs, err := parseTargets(rawProject, objects)
	if err != nil {
		return nil, err
	}
	cloneBuildSettings(buildConfigurations, targets)

	var targetAttributes serialized.Object
	if attributes, ok := rawProject.Object("attributes"); ok {
		targetAttributes, _ = attributes.Object("TargetAttributes")
	}

	base.targets = targets
	base.buildConfigurations = buildConfigurations
	base.defaultConfigurationName = defaultConfigurationName
	base.filePath = filepath.Join(base.path, pbxProjFileName)
	base.file = &pbxprojFile{
		logger:           base.logger,
		path:             base.path,
		rawProj:          rawProj,
		format:           format,
		originalContents: content,
		projectID:        projectID,
		targetAttributes: targetAttributes,
		buildPhaseIDs:    buildPhaseIDs,
	}

	return &base, nil
}

// cloneBuildSettings replaces each configuration's build settings, the raw tree's own map, with an
// in-memory copy. Configurations with the same ID share one copy, as several configuration lists can
// reference one XCBuildConfiguration.
func cloneBuildSettings(projectConfigurations []BuildConfiguration, targets []Target) {
	clones := map[string]serialized.Object{}
	clone := func(configurations []BuildConfiguration) {
		for i := range configurations {
			configuration := &configurations[i]
			if _, ok := clones[configuration.ID]; !ok {
				clones[configuration.ID] = maps.Clone(configuration.buildSettings)
			}
			configuration.buildSettings = clones[configuration.ID]
		}
	}

	clone(projectConfigurations)
	for _, target := range targets {
		clone(target.BuildConfigurations)
	}
}

func decodePBXProj(content []byte) (serialized.Object, int, error) {
	var rawProj serialized.Object

	format, err := plist.UnmarshalWithCustomAnnotation(content, &rawProj)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal project.pbxproj: %w", err)
	}

	return removeCustomInfoObject(rawProj), format, nil
}

func findProject(objects serialized.Object) (string, serialized.Object, error) {
	for id := range objects {
		object, ok := objects.Object(id)
		if !ok {
			continue
		}

		isa, ok := object.String("isa")
		if !ok {
			continue
		}

		if isa == pbxProjectISA {
			return id, object, nil
		}
	}

	return "", nil, fmt.Errorf("failed to find PBXProject in project.pbxproj")
}

// parseTargets returns the project's targets and the build phase IDs of each, by target ID.
func parseTargets(rawProject, objects serialized.Object) ([]Target, map[string][]string, error) {
	targetIDs, ok := rawProject.StringSlice("targets")
	if !ok {
		return nil, nil, fmt.Errorf("project has no targets")
	}

	var targets []Target
	buildPhaseIDs := map[string][]string{}
	for _, targetID := range targetIDs {
		rawTarget, ok := objects.Object(targetID)
		if !ok {
			// The targets list can reference targets that were removed.
			continue
		}

		target, err := parseTarget(targetID, objects)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse target %s: %w", targetID, err)
		}

		targets = append(targets, target)
		if ids, ok := rawTarget.StringSlice("buildPhases"); ok {
			buildPhaseIDs[targetID] = ids
		}
	}

	return targets, buildPhaseIDs, nil
}

// removeCustomInfoObject strips the byte positions UnmarshalWithCustomAnnotation stores under
// customAnnotationKey in every dictionary, so they don't show up as keys or get written back.
// The positions exist for save only because the file on disk should change only where we changed it,
// so perObjectModify decodes them again and replaces only the changed objects.
func removeCustomInfoObject(o serialized.Object) serialized.Object {
	for _, v := range o {
		removeCustomInfo(v)
	}
	return o
}

func removeCustomInfo(o any) any {
	switch container := o.(type) {
	case map[string]any:
		delete(container, customAnnotationKey)
		for _, value := range container {
			removeCustomInfo(value)
		}
		return container
	case []any:
		for _, element := range container {
			removeCustomInfo(element)
		}
		return container
	default:
		return o
	}
}

const (
	nativeTargetISA    = "PBXNativeTarget"
	aggregateTargetISA = "PBXAggregateTarget"
	legacyTargetISA    = "PBXLegacyTarget"
)

func parseTarget(id string, objects serialized.Object) (Target, error) {
	rawTarget, ok := objects.Object(id)
	if !ok {
		return Target{}, fmt.Errorf("target %s not found", id)
	}

	isa, ok := rawTarget.String("isa")
	if !ok {
		return Target{}, fmt.Errorf("target %s has no isa", id)
	}

	var kind targetKind
	switch isa {
	case nativeTargetISA:
		kind = nativeTargetKind
	case aggregateTargetISA:
		kind = aggregateTargetKind
	case legacyTargetISA:
		kind = externalBuildSystemTargetKind
	default:
		return Target{}, fmt.Errorf("unknown target type: %s", isa)
	}

	name, ok := rawTarget.String("name")
	if !ok {
		return Target{}, fmt.Errorf("target %s has no name", id)
	}

	productType, _ := rawTarget.String("productType")

	buildConfigurationListID, ok := rawTarget.String("buildConfigurationList")
	if !ok {
		return Target{}, fmt.Errorf("target %s has no buildConfigurationList", id)
	}

	buildConfigurations, defaultConfigurationName, err := parseConfigurationList(buildConfigurationListID, objects)
	if err != nil {
		return Target{}, fmt.Errorf("target %s: %w", id, err)
	}

	dependencyIDs, _ := rawTarget.StringSlice("dependencies")

	var dependencyTargetIDs []string
	for _, dependencyID := range dependencyIDs {
		// Only direct target dependencies carry a "target" key.
		targetID, ok := targetIDOfDependency(dependencyID, objects)
		if !ok {
			continue
		}
		dependencyTargetIDs = append(dependencyTargetIDs, targetID)
	}

	var productPath string
	if productReferenceID, ok := rawTarget.String("productReference"); ok {
		productReference, ok := objects.Object(productReferenceID)
		if !ok {
			return Target{}, fmt.Errorf("target %s references missing product %s", id, productReferenceID)
		}
		productPath, ok = productReference.String("path")
		if !ok {
			return Target{}, fmt.Errorf("target %s product %s has no path", id, productReferenceID)
		}
	}

	return Target{
		ID:                       id,
		Name:                     name,
		BuildConfigurations:      buildConfigurations,
		DefaultConfigurationName: defaultConfigurationName,
		kind:                     kind,
		productType:              productType,
		productPath:              productPath,
		dependencyTargetIDs:      dependencyTargetIDs,
	}, nil
}

func targetIDOfDependency(dependencyID string, objects serialized.Object) (string, bool) {
	rawDependency, ok := objects.Object(dependencyID)
	if !ok {
		return "", false
	}
	return rawDependency.String("target")
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
		buildSettings: buildSettings,
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
