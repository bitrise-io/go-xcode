package xcodeproj

import (
	"errors"
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/internal/xcproj"
)

var errNotSupportedForXCProj = errors.New("not supported yet for project.xcproj")

// XCProj is a project whose project file is project.xcproj, the JSON format.
type XCProj struct {
	baseProj
}

func (f Factory) newXCProj(content []byte, projectPath string) (*XCProj, error) {
	return parseXCProj(content, f.newBaseProj(projectPath))
}

// TargetDevelopmentTeam implements XcodeProj. project.xcproj keeps the development team only in
// the DEVELOPMENT_TEAM build setting, so it always reports false.
func (p *XCProj) TargetDevelopmentTeam(string) (string, bool) {
	return "", false
}

// AppIconSetPaths implements XcodeProj.
func (p *XCProj) AppIconSetPaths() (map[string][]string, error) {
	return nil, errNotSupportedForXCProj
}

// ForceCodeSign implements XcodeProj.
func (p *XCProj) ForceCodeSign(ForceCodeSignOptions) error {
	return errNotSupportedForXCProj
}

// SetBuildSetting implements XcodeProj.
func (p *XCProj) SetBuildSetting(string, string, string, any) error {
	return errNotSupportedForXCProj
}

// Save implements XcodeProj.
func (p *XCProj) Save() error {
	return errNotSupportedForXCProj
}

func parseXCProj(content []byte, base baseProj) (*XCProj, error) {
	project, err := xcproj.Decode(content)
	if err != nil {
		return nil, err
	}

	targets, err := xcprojTargets(project)
	if err != nil {
		return nil, err
	}

	base.targets = targets
	base.buildConfigurations = xcprojBuildConfigurations(project.Configurations, project.BuildSettings)
	base.defaultConfigurationName = project.DefaultConfiguration

	return &XCProj{baseProj: base}, nil
}

func xcprojTargets(project xcproj.Project) ([]Target, error) {
	targetIDs := make(map[string]string, len(project.Targets))
	for _, target := range project.Targets {
		targetIDs[target.Name] = target.ID
	}

	targets := make([]Target, 0, len(project.Targets))
	for _, target := range project.Targets {
		kind, err := xcprojTargetKind(target.Kind)
		if err != nil {
			return nil, fmt.Errorf("target %s: %w", target.Name, err)
		}

		var dependencyTargetIDs []string
		for _, name := range target.Dependencies {
			id, ok := targetIDs[name]
			if !ok {
				return nil, fmt.Errorf("target %s depends on unknown target %s", target.Name, name)
			}
			dependencyTargetIDs = append(dependencyTargetIDs, id)
		}

		var productPath string
		if len(target.Product) > 0 {
			productPath = target.Product[len(target.Product)-1]
		}

		targets = append(targets, Target{
			ID:                       target.ID,
			Name:                     target.Name,
			BuildConfigurations:      xcprojBuildConfigurations(project.Configurations, target.BuildSettings),
			DefaultConfigurationName: project.DefaultConfiguration,
			kind:                     kind,
			productType:              target.ProductType,
			productPath:              productPath,
			dependencyTargetIDs:      dependencyTargetIDs,
		})
	}

	return targets, nil
}

func xcprojTargetKind(kind string) (targetKind, error) {
	switch kind {
	case xcproj.TargetKindNative:
		return nativeTargetKind, nil
	case xcproj.TargetKindAggregate:
		return aggregateTargetKind, nil
	case xcproj.TargetKindExternalBuildSystem:
		return externalBuildSystemTargetKind, nil
	default:
		return 0, fmt.Errorf("unknown target kind: %s", kind)
	}
}

func xcprojBuildConfigurations(configurations []xcproj.Configuration, settings serialized.Object) []BuildConfiguration {
	buildConfigurations := make([]BuildConfiguration, 0, len(configurations))
	for _, configuration := range configurations {
		buildConfigurations = append(buildConfigurations, BuildConfiguration{
			ID:            configuration.ID,
			Name:          configuration.Name,
			buildSettings: xcproj.SettingsForConfiguration(settings, configuration.Name),
		})
	}
	return buildConfigurations
}
