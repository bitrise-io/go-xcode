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

const (
	// XCWorkspaceExtension ...
	XCWorkspaceExtension = ".xcworkspace"
)

const contentsFileName = "contents.xcworkspacedata"

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

// Workspace represents an Xcode workspace
type Workspace struct {
	FileRefs []FileRef `xml:"FileRef"`
	Groups   []Group   `xml:"Group"`

	Name string
	Path string

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
		if closeErr := file.Close(); closeErr != nil {
			f.logger.Warnf("Failed to close %s: %s", contentsPth, closeErr)
		}
	}()

	var workspace Workspace
	if err := xml.NewDecoder(file).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", contentsPth, err)
	}

	workspace.Name = strings.TrimSuffix(filepath.Base(absPth), filepath.Ext(absPth))
	workspace.Path = absPth
	workspace.logger = f.logger
	workspace.buildSettings = f.buildSettings
	workspace.fileManager = f.fileManager
	workspace.plistFileHandler = f.plistFileHandler
	workspace.pathModifier = f.pathModifier
	workspace.pathChecker = f.pathChecker
	workspace.userProvider = f.userProvider
	workspace.projectFactory = f.projectFactory
	workspace.schemeFactory = f.schemeFactory

	return &workspace, nil
}

// SchemeBuildSettings returns the scheme's effective build settings, resolved by xcodebuild within
// the workspace.
func (w *Workspace) SchemeBuildSettings(scheme, configuration string, customOptions ...string) (serialized.Object, error) {
	return w.buildSettings.SchemeBuildSettings(w.Path, scheme, configuration, customOptions...)
}

// FileLocations returns the absolute paths of the files the workspace references: the top-level
// ones first, then each group's, where a group's own files precede those of its subgroups. Files
// outside the workspace's own tree, developer: and self: locations, are left out, as xcodebuild
// ignores them when listing schemes.
func (w *Workspace) FileLocations() ([]string, error) {
	return w.fileLocations(w.FileRefs, w.Groups, filepath.Dir(w.Path))
}

func (w *Workspace) fileLocations(fileRefs []FileRef, groups []Group, groupDir string) ([]string, error) {
	var locations []string

	for _, fileRef := range fileRefs {
		pth, ok, err := w.absPath(fileRef.Location, groupDir)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		locations = append(locations, pth)
	}

	for _, group := range groups {
		subgroupDir, ok, err := w.absPath(group.Location, groupDir)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		groupLocations, err := w.fileLocations(group.FileRefs, group.Groups, subgroupDir)
		if err != nil {
			return nil, err
		}

		locations = append(locations, groupLocations...)
	}

	return locations, nil
}

// absPath resolves a file reference or group location found in the group at groupDir. ok is false
// for a developer: or self: location, which lies outside the workspace's own tree.
func (w *Workspace) absPath(location, groupDir string) (pth string, ok bool, err error) {
	refType, pth, err := parseLocation(location)
	if err != nil {
		return "", false, err
	}

	switch refType {
	case GroupFileRefType:
		pth = filepath.Join(groupDir, pth)
	case ContainerFileRefType, AbsoluteFileRefType:
		// An absolute: location can hold a relative path, which xcodebuild resolves against the
		// workspace directory, like a container: one.
		if !filepath.IsAbs(pth) {
			pth = filepath.Join(filepath.Dir(w.Path), pth)
		}
	case DeveloperFileRefType, SelfFileRefType:
		w.logger.TDebugf("Skipping %s location outside the workspace: %s", refType, location)
		return "", false, nil
	}

	pth, err = w.pathModifier.AbsPath(pth)
	if err != nil {
		return "", false, err
	}

	return pth, true, nil
}

// ProjectFileLocations ...
func (w *Workspace) ProjectFileLocations() ([]string, error) {
	var projectLocations []string
	fileLocations, err := w.FileLocations()
	if err != nil {
		return nil, err
	}
	for _, fileLocation := range fileLocations {
		if xcodeproj.IsXcodeProj(fileLocation) {
			projectLocations = append(projectLocations, fileLocation)
		}
	}
	return projectLocations, nil
}

// Projects opens the projects the workspace references, in ProjectFileLocations order. A referenced
// project missing from disk is skipped, as xcodebuild -list skips it and its schemes.
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
