// Package xcodeproj reads and edits Xcode project bundles (.xcodeproj).
//
// A project is opened through a Factory, which holds the collaborators every project needs:
// a logger, a filesystem, and a BuildSettingsProvider that runs xcodebuild. Injecting them keeps
// the package testable — Factory.Parse builds a complete project from bytes, with no filesystem
// and no Xcode installation.
//
// # Declared versus effective build settings
//
// The package exposes build settings two ways, and they are not interchangeable.
// BuildConfiguration.BuildSetting returns the value DECLARED in the project file: cheap to read,
// but not variable-expanded, and an .xcconfig file may override it invisibly.
// XcodeProj.TargetBuildSettings returns the EFFECTIVE value by asking xcodebuild: correct, but it
// runs a subprocess. Prefer the effective value for anything that must match what Xcode builds.
//
// # Editing
//
// Edits are made in memory and are not written until Save is called. Save rewrites project.pbxproj
// in place, preserving the ordering and comments of objects that did not change, which keeps the
// file compatible with other tools that read it.
package xcodeproj

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// XcodeProjExtension is the file extension of an Xcode project bundle.
const XcodeProjExtension = ".xcodeproj"

const pbxProjFileName = "project.pbxproj"

const (
	customAnnotationKey = plist.CustomAnnotationKey
	startKey            = plist.CustomAnnotationStartKey
	endKey              = plist.CustomAnnotationEndKey
)

// IsXcodeProj reports whether the path looks like an Xcode project bundle, by extension. It does
// not check the filesystem.
func IsXcodeProj(pth string) bool {
	return filepath.Ext(pth) == XcodeProjExtension
}

// Factory opens Xcode projects using a fixed set of dependencies. Build one per process and reuse
// it: consumers that walk a workspace or scan a repository open many projects.
type Factory struct {
	logger        log.Logger
	buildSettings BuildSettingsProvider
	fileManager   fileutil.FileManager
	pathModifier  pathutil.PathModifier
	pathProvider  pathutil.PathProvider
	userProvider  UserProvider
}

// NewFactory returns a Factory that injects the given collaborators into every project it opens.
func NewFactory(
	logger log.Logger,
	buildSettings BuildSettingsProvider,
	fileManager fileutil.FileManager,
	pathModifier pathutil.PathModifier,
	pathProvider pathutil.PathProvider,
	userProvider UserProvider,
) Factory {
	return Factory{
		logger:        logger,
		buildSettings: buildSettings,
		fileManager:   fileManager,
		pathModifier:  pathModifier,
		pathProvider:  pathProvider,
		userProvider:  userProvider,
	}
}

// Open reads and parses the project at pth, which must be an .xcodeproj directory.
func (f Factory) Open(pth string) (*XcodeProj, error) {
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

	return f.Parse(content, absPth)
}

// Parse builds a project from project.pbxproj content already in memory.
//
// projectPath is the .xcodeproj path the project is considered to live at. It is used to resolve
// relative build setting paths and is where Save writes; it need not exist on disk unless a method
// that touches the filesystem is called.
func (f Factory) Parse(content []byte, projectPath string) (*XcodeProj, error) {
	p, err := parsePBXProj(content)
	if err != nil {
		return nil, err
	}

	p.logger = f.logger
	p.buildSettings = f.buildSettings
	p.fileManager = f.fileManager
	p.pathModifier = f.pathModifier
	p.pathProvider = f.pathProvider
	p.userProvider = f.userProvider

	p.Path = projectPath
	p.Name = strings.TrimSuffix(filepath.Base(projectPath), filepath.Ext(projectPath))

	return p, nil
}

// XcodeProj is a parsed Xcode project. Obtain one from a Factory.
//
// It is a handle to a mutable document, so it is passed by pointer and is not safe for concurrent
// use while being edited.
type XcodeProj struct {
	// Name is the project's file name without the .xcodeproj extension.
	Name string
	// Path is the absolute path of the .xcodeproj directory.
	Path string

	logger        log.Logger
	buildSettings BuildSettingsProvider
	fileManager   fileutil.FileManager
	pathModifier  pathutil.PathModifier
	pathProvider  pathutil.PathProvider
	userProvider  UserProvider

	// rawProj is the decoded project.pbxproj. It is the source of truth that Save writes out:
	// every edit lands here, and the parsed fields below are derived from it.
	rawProj serialized.Object
	format  int
	// originalContents is the file as read. Save diffs against a re-parse of it to find which
	// objects changed, so it can splice only those byte ranges and leave the rest untouched.
	originalContents []byte

	projectID                string
	targets                  []Target
	buildConfigurations      []BuildConfiguration
	defaultConfigurationName string
	targetAttributes         serialized.Object
}

// Targets returns the project's targets, in project file order.
func (p *XcodeProj) Targets() []Target {
	return p.targets
}

// Target returns the target with the given ID, which is the identifier a scheme's
// BuildableReference uses as its BlueprintIdentifier.
func (p *XcodeProj) Target(id string) (Target, bool) {
	for _, target := range p.targets {
		if target.ID == id {
			return target, true
		}
	}
	return Target{}, false
}

// TargetByName returns the target with the given name.
func (p *XcodeProj) TargetByName(name string) (Target, bool) {
	for _, target := range p.targets {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

// BuildConfigurations returns the project-level build configurations, which targets inherit from.
func (p *XcodeProj) BuildConfigurations() []BuildConfiguration {
	return p.buildConfigurations
}

// DefaultConfigurationName returns the name of the project's default build configuration, or an
// empty string if the project does not name one.
func (p *XcodeProj) DefaultConfigurationName() string {
	return p.defaultConfigurationName
}

// DependentTargetsOfTarget returns every target the given target depends on, including transitive
// dependencies. Each target appears once. Dependencies that cannot be resolved are logged and
// skipped rather than failing the call, because a project can reference targets it does not
// contain.
func (p *XcodeProj) DependentTargetsOfTarget(target Target) []Target {
	// visited guards against a dependency cycle. Xcode does not create one, but a hand-edited or
	// generated project file can, and without the guard the recursion would not terminate.
	visited := map[string]bool{target.ID: true}
	return deduplicateTargets(p.dependentTargetsOfTarget(target, visited))
}

func (p *XcodeProj) dependentTargetsOfTarget(target Target, visited map[string]bool) []Target {
	var dependents []Target

	for _, dependencyTargetID := range target.dependencyTargetIDs {
		child, ok := p.Target(dependencyTargetID)
		if !ok {
			p.logger.Warnf("couldn't find dependency %s of target %s (%s), skipping", dependencyTargetID, target.Name, target.ID)
			continue
		}

		// A target already seen is skipped entirely. That keeps a cycle from recursing forever and
		// keeps the target the caller asked about out of its own result.
		if visited[child.ID] {
			continue
		}
		visited[child.ID] = true

		dependents = append(dependents, child)
		dependents = append(dependents, p.dependentTargetsOfTarget(child, visited)...)
	}

	return dependents
}

// TargetDevelopmentTeam returns the DevelopmentTeam recorded for the target in the project's
// TargetAttributes. ok is false when the project has no attributes for the target or the target has
// no team set, both of which are normal.
//
// This reads what the project file declares. It is not the same question as which team Xcode would
// build with, which also depends on build settings.
func (p *XcodeProj) TargetDevelopmentTeam(targetID string) (string, bool) {
	if p.targetAttributes == nil {
		return "", false
	}

	attributes, ok := p.targetAttributes.Object(targetID)
	if !ok {
		return "", false
	}

	developmentTeam, ok := attributes.String("DevelopmentTeam")
	if !ok || developmentTeam == "" {
		return "", false
	}

	return developmentTeam, true
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
