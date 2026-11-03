package xcodeproj

import (
	"errors"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/internal/xcproj"
)

var (
	errNotSupportedForXCProj = errors.New("not supported yet for project.xcproj")

	xcprojTargetKinds = map[string]targetKind{
		xcproj.TargetKindNative:              nativeTargetKind,
		xcproj.TargetKindAggregate:           aggregateTargetKind,
		xcproj.TargetKindExternalBuildSystem: externalBuildSystemTargetKind,
	}
)

// XCProj is a project whose project file is project.xcproj, the JSON format.
type XCProj struct {
	baseProj

	developmentTeams map[string]string
}

func (f Factory) newXCProj(content []byte, projectPath string) (*XCProj, error) {
	return parseXCProj(content, f.newBaseProj(projectPath))
}

// TargetDevelopmentTeam implements XcodeProj. It reads the target's legacy-team-id.
func (p *XCProj) TargetDevelopmentTeam(targetID string) (string, bool) {
	developmentTeam := p.developmentTeams[targetID]
	return developmentTeam, developmentTeam != ""
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

	base.targets = xcprojTargets(project, base.logger)
	base.buildConfigurations = xcprojBuildConfigurations(project.Configurations, project.BuildSettings)
	base.defaultConfigurationName = project.DefaultConfiguration

	return &XCProj{baseProj: base, developmentTeams: xcprojDevelopmentTeams(project.Targets)}, nil
}

func xcprojTargets(project xcproj.Project, logger log.Logger) []Target {
	idsByName := make(map[string]string, len(project.Targets))
	for _, target := range project.Targets {
		idsByName[target.Name] = target.ID
	}

	targets := make([]Target, 0, len(project.Targets))
	for _, target := range project.Targets {
		var dependencyTargetIDs []string
		for _, name := range target.Dependencies {
			dependencyTargetIDs = append(dependencyTargetIDs, idsByName[name])
		}

		var productPath string
		if len(target.Product) > 0 {
			productPath = target.Product[len(target.Product)-1]
		} else if target.ProductID != "" {
			logger.Warnf("product of target %s (%s) is referenced by ID %s, its path is unknown", target.Name, target.ID, target.ProductID)
		}

		targets = append(targets, Target{
			ID:                       target.ID,
			Name:                     target.Name,
			BuildConfigurations:      xcprojBuildConfigurations(project.Configurations, target.BuildSettings),
			DefaultConfigurationName: project.DefaultConfiguration,
			kind:                     xcprojTargetKinds[target.Kind],
			productType:              target.ProductType,
			productPath:              productPath,
			dependencyTargetIDs:      dependencyTargetIDs,
		})
	}

	return targets
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

func xcprojDevelopmentTeams(targets []xcproj.Target) map[string]string {
	developmentTeams := make(map[string]string, len(targets))
	for _, target := range targets {
		developmentTeams[target.ID] = target.LegacyTeamID
	}
	return developmentTeams
}
