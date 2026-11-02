package xcodeproj

import (
	"errors"
	"strings"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/internal/xcproj"
)

const xcprojIDReferencePrefix = "id:"

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
}

func (f Factory) newXCProj(content []byte, projectPath string) (*XCProj, error) {
	return parseXCProj(content, f.newBaseProj(projectPath))
}

// TargetDevelopmentTeam implements XcodeProj. project.xcproj records the development team only in
// the DEVELOPMENT_TEAM build setting.
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

	base.targets = xcprojTargets(project, base.logger)
	base.buildConfigurations = xcprojBuildConfigurations(project.Configurations, project.BuildSettings)
	base.defaultConfigurationName = project.DefaultConfiguration

	return &XCProj{baseProj: base}, nil
}

func xcprojTargets(project xcproj.Project, logger log.Logger) []Target {
	knownIDs := make(map[string]bool, len(project.Targets))
	// An empty ID marks a name more than one target has.
	idsByName := make(map[string]string, len(project.Targets))
	for _, target := range project.Targets {
		knownIDs[target.ID] = true
		if _, ok := idsByName[target.Name]; ok {
			idsByName[target.Name] = ""
		} else {
			idsByName[target.Name] = target.ID
		}
	}

	targets := make([]Target, 0, len(project.Targets))
	for _, target := range project.Targets {
		var dependencyTargetIDs []string
		for _, reference := range target.Dependencies {
			id, ok := resolveXCProjDependency(reference, idsByName, knownIDs)
			if !ok {
				logger.Warnf("couldn't resolve dependency %s of target %s (%s) to a single target, skipping", reference, target.Name, target.ID)
				continue
			}
			dependencyTargetIDs = append(dependencyTargetIDs, id)
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

func resolveXCProjDependency(reference string, idsByName map[string]string, knownIDs map[string]bool) (string, bool) {
	if id, ok := idsByName[reference]; ok {
		return id, id != ""
	}
	if id, ok := strings.CutPrefix(reference, xcprojIDReferencePrefix); ok && knownIDs[id] {
		return id, true
	}
	return "", false
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
