// Package xcodeproj reads and edits Xcode project bundles (.xcodeproj).
package xcodeproj

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

const (
	// XcodeProjExtension is the file extension of an Xcode project bundle.
	XcodeProjExtension = ".xcodeproj"

	pbxProjFileName = "project.pbxproj"
	xcProjFileName  = "project.xcproj"
)

// Factory opens Xcode projects with a fixed set of dependencies.
type Factory struct {
	logger           log.Logger
	buildSettings    BuildSettingsProvider
	fileManager      fileutil.FileManager
	plistFileHandler plistutil.FileHandler
	pathModifier     pathutil.PathModifier
	pathProvider     pathutil.PathProvider
	userProvider     UserProvider
	dirCreator       DirCreator
	schemeFactory    xcscheme.Factory
}

// projectFile is the format-specific part of a project: how edits reach the project file and where
// asset catalogs come from. Everything else is the same for both formats.
type projectFile interface {
	setBuildSetting(target Target, configuration BuildConfiguration, key string, value any) error
	setManualSigning(target Target, developmentTeam string)
	content() ([]byte, error)
	assetCatalogPaths(target Target) ([]string, error)
}

// Project is a parsed Xcode project, read from its project.pbxproj or project.xcproj.
type Project struct {
	name string
	path string

	file projectFile
	// filePath is where Save writes file: <path>/project.pbxproj or <path>/project.xcproj.
	filePath string

	logger           log.Logger
	buildSettings    BuildSettingsProvider
	fileManager      fileutil.FileManager
	plistFileHandler plistutil.FileHandler
	pathModifier     pathutil.PathModifier
	pathProvider     pathutil.PathProvider
	userProvider     UserProvider
	dirCreator       DirCreator
	schemeFactory    xcscheme.Factory

	targets                  []Target
	buildConfigurations      []BuildConfiguration
	defaultConfigurationName string
}

// NewFactory returns a Factory that injects the given collaborators into every project it opens.
func NewFactory(
	logger log.Logger,
	buildSettings BuildSettingsProvider,
	fileManager fileutil.FileManager,
	plistFileHandler plistutil.FileHandler,
	pathModifier pathutil.PathModifier,
	pathProvider pathutil.PathProvider,
	userProvider UserProvider,
	dirCreator DirCreator,
) Factory {
	return Factory{
		logger:           logger,
		buildSettings:    buildSettings,
		fileManager:      fileManager,
		plistFileHandler: plistFileHandler,
		pathModifier:     pathModifier,
		pathProvider:     pathProvider,
		userProvider:     userProvider,
		dirCreator:       dirCreator,
		schemeFactory:    xcscheme.NewFactory(fileManager),
	}
}

// IsXcodeProj reports whether the path has the .xcodeproj extension.
func IsXcodeProj(pth string) bool {
	return filepath.Ext(pth) == XcodeProjExtension
}

// Open reads and parses the project at pth: its project.pbxproj or its project.xcproj. A bundle
// with both is an error, because xcodebuild refuses to read it.
func (f Factory) Open(pth string) (*Project, error) {
	absPth, err := f.pathModifier.AbsPath(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path (%s): %w", pth, err)
	}

	pbxProjPath := filepath.Join(absPth, pbxProjFileName)
	xcProjPath := filepath.Join(absPth, xcProjFileName)
	hasPBXProj, err := f.projectFileExists(pbxProjPath)
	if err != nil {
		return nil, err
	}
	hasXCProj, err := f.projectFileExists(xcProjPath)
	if err != nil {
		return nil, err
	}

	switch {
	case hasPBXProj && hasXCProj:
		return nil, fmt.Errorf("found both %s and %s in %s: Xcode treats a project with both as invalid, remove the obsolete one", pbxProjFileName, xcProjFileName, absPth)
	case hasPBXProj:
		content, err := f.readProjectFile(pbxProjPath)
		if err != nil {
			return nil, err
		}
		return f.newPBXProj(content, absPth)
	case hasXCProj:
		content, err := f.readProjectFile(xcProjPath)
		if err != nil {
			return nil, err
		}
		return f.newXCProj(content, absPth)
	default:
		return nil, fmt.Errorf("found neither %s nor %s in %s: %w", pbxProjFileName, xcProjFileName, absPth, fs.ErrNotExist)
	}
}

func (f Factory) projectFileExists(pth string) (bool, error) {
	_, err := f.fileManager.Lstat(pth)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check %s: %w", pth, err)
	}
	return true, nil
}

func (f Factory) readProjectFile(pth string) ([]byte, error) {
	file, err := f.fileManager.Open(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", pth, err)
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", pth, err)
	}
	return content, nil
}

func (f Factory) newProject(projectPath string) Project {
	return Project{
		name:             strings.TrimSuffix(filepath.Base(projectPath), filepath.Ext(projectPath)),
		path:             projectPath,
		logger:           f.logger,
		buildSettings:    f.buildSettings,
		fileManager:      f.fileManager,
		plistFileHandler: f.plistFileHandler,
		pathModifier:     f.pathModifier,
		pathProvider:     f.pathProvider,
		userProvider:     f.userProvider,
		dirCreator:       f.dirCreator,
		schemeFactory:    f.schemeFactory,
	}
}

// Name returns the bundle name without the .xcodeproj extension.
func (p *Project) Name() string {
	return p.name
}

// Path returns where the bundle is considered to live: relative build setting paths resolve
// against it and Save writes to it.
func (p *Project) Path() string {
	return p.path
}

// Targets returns the project's targets, in project file order.
func (p *Project) Targets() []Target {
	return p.targets
}

// Target returns the target with the given ID.
func (p *Project) Target(id string) (Target, bool) {
	for _, target := range p.targets {
		if target.ID == id {
			return target, true
		}
	}
	return Target{}, false
}

// TargetByName returns the target with the given name.
func (p *Project) TargetByName(name string) (Target, bool) {
	for _, target := range p.targets {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

// BuildConfigurations returns the project-level build configurations.
func (p *Project) BuildConfigurations() []BuildConfiguration {
	return p.buildConfigurations
}

// DefaultConfigurationName returns the name of the project's default build configuration.
func (p *Project) DefaultConfigurationName() string {
	return p.defaultConfigurationName
}

// DependentTargetsOfTarget returns the direct and transitive dependencies of target, each once.
// Dependencies that cannot be resolved are logged and skipped.
func (p *Project) DependentTargetsOfTarget(target Target) []Target {
	// visited guards against dependency cycles in malformed project files.
	visited := map[string]bool{target.ID: true}
	return deduplicateTargets(p.dependentTargetsOfTarget(target, visited))
}

func (p *Project) dependentTargetsOfTarget(target Target, visited map[string]bool) []Target {
	var dependents []Target

	for _, dependencyTargetID := range target.dependencyTargetIDs {
		child, ok := p.Target(dependencyTargetID)
		if !ok {
			p.logger.Warnf("couldn't find dependency %s of target %s (%s), skipping", dependencyTargetID, target.Name, target.ID)
			continue
		}

		if visited[child.ID] {
			continue
		}
		visited[child.ID] = true

		dependents = append(dependents, child)
		dependents = append(dependents, p.dependentTargetsOfTarget(child, visited)...)
	}

	return dependents
}

func deduplicateTargets(targets []Target) []Target {
	seen := map[string]bool{}
	unique := make([]Target, 0, len(targets))

	for _, target := range targets {
		if seen[target.ID] {
			continue
		}
		seen[target.ID] = true
		unique = append(unique, target)
	}

	return unique
}
