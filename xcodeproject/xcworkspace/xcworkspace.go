// Package xcworkspace reads Xcode workspace bundles (.xcworkspace).
package xcworkspace

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

// XCWorkspaceExtension is the file extension of an Xcode workspace bundle.
const XCWorkspaceExtension = ".xcworkspace"

const contentsFileName = "contents.xcworkspacedata"

// IsWorkspace reports whether the path has the .xcworkspace extension.
func IsWorkspace(pth string) bool {
	return filepath.Ext(pth) == XCWorkspaceExtension
}

// BuildSettingsProvider resolves the effective build settings of a scheme.
type BuildSettingsProvider interface {
	SchemeBuildSettings(workspacePath, scheme, configuration string, extraArgs ...string) (serialized.Object, error)
}

// Factory opens Xcode workspaces with a fixed set of dependencies.
type Factory struct {
	logger           log.Logger
	buildSettings    BuildSettingsProvider
	fileManager      fileutil.FileManager
	plistFileHandler plistutil.FileHandler
	pathModifier     pathutil.PathModifier
	pathChecker      pathutil.PathChecker
	userProvider     xcodeproj.UserProvider
	projectFactory   xcodeproj.Factory
	schemeFactory    xcscheme.Factory
}

// NewFactory returns a Factory that injects the given collaborators into every workspace it opens.
// projectFactory opens the projects a workspace references.
func NewFactory(
	logger log.Logger,
	buildSettings BuildSettingsProvider,
	fileManager fileutil.FileManager,
	plistFileHandler plistutil.FileHandler,
	pathModifier pathutil.PathModifier,
	pathChecker pathutil.PathChecker,
	userProvider xcodeproj.UserProvider,
	projectFactory xcodeproj.Factory,
) Factory {
	return Factory{
		logger:           logger,
		buildSettings:    buildSettings,
		fileManager:      fileManager,
		plistFileHandler: plistFileHandler,
		pathModifier:     pathModifier,
		pathChecker:      pathChecker,
		userProvider:     userProvider,
		projectFactory:   projectFactory,
		schemeFactory:    xcscheme.NewFactory(fileManager),
	}
}

// Open reads the workspace at pth.
func (f Factory) Open(pth string) (*Workspace, error) {
	absPth, err := f.pathModifier.AbsPath(pth)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path (%s): %w", pth, err)
	}

	contentsPth := filepath.Join(absPth, contentsFileName)

	file, err := f.fileManager.Open(contentsPth)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", contentsPth, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var contents contents
	if err := xml.NewDecoder(file).Decode(&contents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", contentsPth, err)
	}

	return &Workspace{
		Name:             strings.TrimSuffix(filepath.Base(absPth), filepath.Ext(absPth)),
		Path:             absPth,
		fileRefs:         contents.FileRefs,
		groups:           contents.Groups,
		logger:           f.logger,
		buildSettings:    f.buildSettings,
		fileManager:      f.fileManager,
		plistFileHandler: f.plistFileHandler,
		pathModifier:     f.pathModifier,
		pathChecker:      f.pathChecker,
		userProvider:     f.userProvider,
		projectFactory:   f.projectFactory,
		schemeFactory:    f.schemeFactory,
	}, nil
}

// Workspace is a parsed Xcode workspace.
type Workspace struct {
	Name string
	Path string

	fileRefs []FileRef
	groups   []Group

	logger           log.Logger
	buildSettings    BuildSettingsProvider
	fileManager      fileutil.FileManager
	plistFileHandler plistutil.FileHandler
	pathModifier     pathutil.PathModifier
	pathChecker      pathutil.PathChecker
	userProvider     xcodeproj.UserProvider
	projectFactory   xcodeproj.Factory
	schemeFactory    xcscheme.Factory
}

// FileLocations returns the absolute paths of the files the workspace references: the top-level
// ones first, then each group's, where a group's own files precede those of its subgroups.
func (w *Workspace) FileLocations() ([]string, error) {
	return w.fileLocations(w.fileRefs, w.groups, filepath.Dir(w.Path))
}

func (w *Workspace) fileLocations(fileRefs []FileRef, groups []Group, groupDir string) ([]string, error) {
	var locations []string

	for _, fileRef := range fileRefs {
		pth, err := w.absPath(fileRef.Location, groupDir)
		if err != nil {
			return nil, err
		}

		locations = append(locations, pth)
	}

	for _, group := range groups {
		subgroupDir, err := w.absPath(group.Location, groupDir)
		if err != nil {
			return nil, err
		}

		groupLocations, err := w.fileLocations(group.FileRefs, group.Groups, subgroupDir)
		if err != nil {
			return nil, err
		}

		locations = append(locations, groupLocations...)
	}

	return locations, nil
}

// absPath resolves a file reference or group location found in the group at groupDir.
func (w *Workspace) absPath(location, groupDir string) (string, error) {
	refType, pth, err := parseLocation(location)
	if err != nil {
		return "", err
	}

	switch refType {
	case GroupFileRefType:
		pth = filepath.Join(groupDir, pth)
	case ContainerFileRefType:
		pth = filepath.Join(filepath.Dir(w.Path), pth)
	case AbsoluteFileRefType:
	}

	return w.pathModifier.AbsPath(pth)
}

// ProjectFileLocations returns the absolute paths of the projects the workspace references, in
// FileLocations order.
func (w *Workspace) ProjectFileLocations() ([]string, error) {
	locations, err := w.FileLocations()
	if err != nil {
		return nil, err
	}

	var projectLocations []string
	for _, location := range locations {
		if xcodeproj.IsXcodeProj(location) {
			projectLocations = append(projectLocations, location)
		}
	}

	return projectLocations, nil
}

// Projects opens the projects the workspace references, in ProjectFileLocations order. A referenced
// project missing from disk is skipped, as Xcode opens the workspace without it.
func (w *Workspace) Projects() ([]*xcodeproj.XcodeProj, error) {
	locations, err := w.ProjectFileLocations()
	if err != nil {
		return nil, fmt.Errorf("failed to get project locations from workspace: %w", err)
	}

	var projects []*xcodeproj.XcodeProj
	for _, location := range locations {
		exists, err := w.pathChecker.IsPathExists(location)
		if err != nil {
			return nil, fmt.Errorf("failed to check if project (%s) exists: %w", location, err)
		}
		if !exists {
			w.logger.TDebugf("Skipping missing project: %s", location)
			continue
		}

		project, err := w.projectFactory.Open(location)
		if err != nil {
			return nil, fmt.Errorf("failed to open project (%s): %w", location, err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// SchemeBuildSettings returns the scheme's effective build settings, resolved by xcodebuild within
// the workspace.
func (w *Workspace) SchemeBuildSettings(scheme, configuration string, extraArgs ...string) (serialized.Object, error) {
	return w.buildSettings.SchemeBuildSettings(w.Path, scheme, configuration, extraArgs...)
}
