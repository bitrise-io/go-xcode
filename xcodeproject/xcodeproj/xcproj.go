package xcodeproj

import (
	"fmt"
	"maps"
	"path"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/internal/xcproj"
)

var (
	xcprojTargetKinds = map[string]targetKind{
		xcproj.TargetKindNative:              nativeTargetKind,
		xcproj.TargetKindAggregate:           aggregateTargetKind,
		xcproj.TargetKindExternalBuildSystem: externalBuildSystemTargetKind,
	}
)

// xcprojFile is project.xcproj, the JSON format.
type xcprojFile struct {
	fileManager fileutil.FileManager
	// path is the project bundle's path, which <PROJECT>/ file paths resolve against.
	path string

	original           []byte
	configurationNames []string
	// editor is created by the first edit; until then content returns original.
	editor *xcproj.Editor
	files  []xcproj.Reference
}

func (f Factory) newXCProj(content []byte, projectPath string) (*xcodeProject, error) {
	return parseXCProj(content, f.newXcodeProject(projectPath))
}

// setBuildSetting supports only string values.
//
// TODO: a key with a [config=...] condition is rejected here but accepted by
// pbxprojFile.setBuildSetting; make the two formats agree.
func (p *xcprojFile) setBuildSetting(target Target, configuration BuildConfiguration, key string, value any) error {
	stringValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to set build setting %s: a %T value is not supported for project.xcproj, only a string", key, value)
	}

	if p.editor == nil {
		editor, err := xcproj.NewEditor(p.original, p.configurationNames)
		if err != nil {
			return err
		}
		p.editor = editor
	}

	return p.editor.SetTargetBuildSetting(target.Name, configuration.Name, key, stringValue)
}

// setManualSigning leaves legacy-team-id and legacy-provisioning-style alone, as Xcode ignores
// them.
func (p *xcprojFile) setManualSigning(Target, string) {}

func (p *xcprojFile) content() ([]byte, error) {
	if p.editor == nil {
		return p.original, nil
	}
	return p.editor.Content(), nil
}

func parseXCProj(content []byte, base xcodeProject) (*xcodeProject, error) {
	project, err := xcproj.Decode(content)
	if err != nil {
		return nil, err
	}

	base.targets = xcprojTargets(project, base.logger)
	base.buildConfigurations = xcprojBuildConfigurations(project.Configurations, project.BuildSettings)
	base.defaultConfigurationName = project.DefaultConfiguration

	configurationNames := make([]string, 0, len(base.buildConfigurations))
	for _, configuration := range base.buildConfigurations {
		configurationNames = append(configurationNames, configuration.Name)
	}

	base.filePath = filepath.Join(base.path, xcProjFileName)
	base.file = &xcprojFile{
		fileManager:        base.fileManager,
		path:               base.path,
		original:           content,
		configurationNames: configurationNames,
		files:              project.Files,
	}

	return &base, nil
}

func xcprojTargets(project xcproj.Project, logger log.Logger) []Target {
	idsByName := make(map[string]string, len(project.Targets))
	for _, target := range project.Targets {
		idsByName[target.Name] = target.ID
	}
	filePathsByID := xcprojFilePathsByID(project.Files)

	targets := make([]Target, 0, len(project.Targets))
	for _, target := range project.Targets {
		var dependencyTargetIDs []string
		for _, name := range target.Dependencies {
			dependencyTargetIDs = append(dependencyTargetIDs, idsByName[name])
		}

		var productPath string
		if len(target.Product) > 0 {
			// The last element is the file name: "product": "Products/App.app" decodes to [Products App.app].
			productPath = xcprojProductName(target.Name, target.Product[len(target.Product)-1])
		} else if filePath, ok := filePathsByID[target.ProductID]; ok {
			productPath = xcprojProductName(target.Name, path.Base(filePath))
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

func xcprojFilePathsByID(references []xcproj.Reference) map[string]string {
	pathsByID := map[string]string{}
	for _, reference := range references {
		if reference.ID != "" {
			pathsByID[reference.ID] = reference.Path
		}
		maps.Copy(pathsByID, xcprojFilePathsByID(reference.Children))
	}
	return pathsByID
}

// xcprojProductName falls back to the target name, PRODUCT_NAME's default, for a product file with
// no name, such as <PRODUCTS>/.app.
func xcprojProductName(targetName, fileName string) string {
	if extension := path.Ext(fileName); extension != "" && fileName == extension {
		return targetName + extension
	}
	return fileName
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
