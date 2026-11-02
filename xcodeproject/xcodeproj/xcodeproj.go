// Package xcodeproj reads and edits Xcode project bundles (.xcodeproj).
package xcodeproj

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

// XcodeProjExtension is the file extension of an Xcode project bundle.
const XcodeProjExtension = ".xcodeproj"

const pbxProjFileName = "project.pbxproj"

// IsXcodeProj reports whether the path has the .xcodeproj extension.
func IsXcodeProj(pth string) bool {
	return filepath.Ext(pth) == XcodeProjExtension
}

// XcodeProj is a parsed Xcode project.
type XcodeProj interface {
	// Name returns the bundle name without the .xcodeproj extension.
	Name() string
	// Path returns where the bundle is considered to live: relative build setting paths resolve
	// against it and Save writes to it.
	Path() string

	// Targets returns the project's targets, in project file order.
	Targets() []Target
	// Target returns the target with the given ID.
	Target(id string) (Target, bool)
	// TargetByName returns the target with the given name.
	TargetByName(name string) (Target, bool)
	// BuildConfigurations returns the project-level build configurations.
	BuildConfigurations() []BuildConfiguration
	// DefaultConfigurationName returns the name of the project's default build configuration.
	DefaultConfigurationName() string
	// DependentTargetsOfTarget returns the direct and transitive dependencies of target, each once.
	// Dependencies that cannot be resolved are logged and skipped.
	DependentTargetsOfTarget(target Target) []Target
	// TargetDevelopmentTeam returns the development team the project records for the target
	// outside its build settings.
	TargetDevelopmentTeam(targetID string) (string, bool)

	// TargetBuildSettings returns the target's effective build settings, looked up by target (not
	// scheme) with xcodebuild. Destination-style extraArgs have no effect in target mode.
	TargetBuildSettings(target, configuration string, extraArgs ...string) (serialized.Object, error)
	// TargetBundleID returns the resolved bundle ID, falling back to the Info.plist's
	// CFBundleIdentifier.
	TargetBundleID(target, configuration string) (string, error)
	// TargetCodeSignEntitlements returns the target's entitlements, or
	// ErrCodeSignEntitlementsNotFound if the target declares none.
	TargetCodeSignEntitlements(target, configuration string) (serialized.Object, error)
	// TargetInfoplistPath returns the absolute path of the target's Info.plist, or
	// ErrInfoPlistNotFound if the target declares none.
	TargetInfoplistPath(target, configuration string) (string, error)
	// AppIconSetPaths maps target IDs to the absolute paths of the targets' .appiconset directories.
	AppIconSetPaths() (map[string][]string, error)

	// Schemes returns the project's shared and user schemes. If there are none, it generates
	// Xcode's default schemes, unless the "Autocreate schemes" option is off, in which case it
	// returns an error.
	Schemes() ([]xcscheme.Scheme, error)
	// SchemesWithAutocreateOverride is Schemes with the "Autocreate schemes" option set by the
	// caller, for use within a workspace. With no schemes and autocreate off it returns no schemes
	// and no error.
	SchemesWithAutocreateOverride(autocreateEnabled bool) ([]xcscheme.Scheme, error)
	// Scheme returns the scheme with the given name and the path of its project. Names are compared
	// after Unicode normalisation.
	Scheme(name string) (xcscheme.Scheme, string, error)
	// RecreateSchemes returns the schemes Xcode would create for the project, one per native,
	// non-test target. It only builds them in memory; use SaveSharedScheme to write them.
	RecreateSchemes() []xcscheme.Scheme
	// SaveSharedScheme writes the scheme to <Path>/xcshareddata/xcschemes/<scheme name>.xcscheme,
	// overwriting an existing file. Missing directories are created.
	SaveSharedScheme(scheme xcscheme.Scheme) error

	// ForceCodeSign switches a target configuration to manual code signing. Call Save to persist
	// it.
	ForceCodeSign(opts ForceCodeSignOptions) error
	// SetBuildSetting sets a build setting of a target configuration, such as
	// CURRENT_PROJECT_VERSION or PRODUCT_BUNDLE_IDENTIFIER, in the project file. Several settings
	// can be set before one Save persists them. Targets fetched earlier see the change.
	SetBuildSetting(targetName, configurationName, key string, value any) error
	// Save writes the project file, changing only what the edits changed.
	Save() error
}

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

// Open reads and parses the project at pth.
func (f Factory) Open(pth string) (XcodeProj, error) {
	absPth, err := f.pathModifier.AbsPath(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path (%s): %w", pth, err)
	}

	pbxProjPth := filepath.Join(absPth, pbxProjFileName)

	file, err := f.fileManager.Open(pbxProjPth)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", pbxProjPth, err)
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", pbxProjPth, err)
	}

	// The buffer was read here and nobody else holds it, so it needs no copy.
	project, err := f.newPBXProj(content, absPth)
	if err != nil {
		// Returning the nil *PBXProj would make a non-nil XcodeProj.
		return nil, err
	}
	return project, nil
}

func (f Factory) newBaseProj(projectPath string) baseProj {
	return baseProj{
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

// baseProj is what's common in the PBXProj and XCProj implementations.
type baseProj struct {
	name string
	path string

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

// Name implements XcodeProj.
func (p *baseProj) Name() string {
	return p.name
}

// Path implements XcodeProj.
func (p *baseProj) Path() string {
	return p.path
}

// Targets implements XcodeProj.
func (p *baseProj) Targets() []Target {
	return p.targets
}

// Target implements XcodeProj.
func (p *baseProj) Target(id string) (Target, bool) {
	for _, target := range p.targets {
		if target.ID == id {
			return target, true
		}
	}
	return Target{}, false
}

// TargetByName implements XcodeProj.
func (p *baseProj) TargetByName(name string) (Target, bool) {
	for _, target := range p.targets {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

// BuildConfigurations implements XcodeProj.
func (p *baseProj) BuildConfigurations() []BuildConfiguration {
	return p.buildConfigurations
}

// DefaultConfigurationName implements XcodeProj.
func (p *baseProj) DefaultConfigurationName() string {
	return p.defaultConfigurationName
}

// DependentTargetsOfTarget implements XcodeProj.
func (p *baseProj) DependentTargetsOfTarget(target Target) []Target {
	// visited guards against dependency cycles in malformed project files.
	visited := map[string]bool{target.ID: true}
	return deduplicateTargets(p.dependentTargetsOfTarget(target, visited))
}

func (p *baseProj) dependentTargetsOfTarget(target Target, visited map[string]bool) []Target {
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
