package xcodeproj

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
)

// XcodeProject ...
type XcodeProject interface {
	Path() string
	SharedSchemes() []SchemeModel
	Targets() []TargetModel
}

// ProjectModel ...
type ProjectModel struct {
	pth  string
	name string

	sharedSchemes []SchemeModel
	targets       []TargetModel

	SDKs []string
}

// NewProject ...
func NewProject(xcodeprojPth string) (ProjectModel, error) {
	project := ProjectModel{
		pth:  xcodeprojPth,
		name: strings.TrimSuffix(filepath.Base(xcodeprojPth), filepath.Ext(xcodeprojPth)),
	}

	// SDK
	pbxprojPth := filepath.Join(xcodeprojPth, "project.pbxproj")

	if exist, err := pathutil.IsPathExists(pbxprojPth); err != nil {
		return ProjectModel{}, err
	} else if !exist {
		return ProjectModel{}, fmt.Errorf("Project descriptor not found at: %s", pbxprojPth)
	}

	sdks, err := GetBuildConfigSDKs(pbxprojPth)
	if err != nil {
		return ProjectModel{}, err
	}

	project.SDKs = sdks

	// Shared Schemes
	schemes, err := ProjectSharedSchemes(xcodeprojPth)
	if err != nil {
		return ProjectModel{}, err
	}

	project.sharedSchemes = schemes

	// Targets
	targets, err := ProjectTargets(xcodeprojPth)
	if err != nil {
		return ProjectModel{}, err
	}

	project.targets = targets

	return project, nil
}

// Path ...
func (p ProjectModel) Path() string {
	return p.pth
}

// SharedSchemes ...
func (p ProjectModel) SharedSchemes() []SchemeModel {
	return p.sharedSchemes
}

// Targets ...
func (p ProjectModel) Targets() []TargetModel {
	return p.targets
}

// ContainsSDK ...
func (p ProjectModel) ContainsSDK(sdk string) bool {
	for _, s := range p.SDKs {
		if s == sdk {
			return true
		}
	}
	return false
}
