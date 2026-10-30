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
		_ = file.Close()
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

// FileLocations ...
func (w *Workspace) FileLocations() ([]string, error) {
	var fileLocations []string

	for _, fileRef := range w.FileRefs {
		pth, err := fileRef.AbsPath(filepath.Dir(w.Path))
		if err != nil {
			return nil, err
		}

		fileLocations = append(fileLocations, pth)
	}

	for _, group := range w.Groups {
		groupFileLocations, err := group.FileLocations(filepath.Dir(w.Path))
		if err != nil {
			return nil, err
		}

		fileLocations = append(fileLocations, groupFileLocations...)
	}

	return fileLocations, nil
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
