package xcodeproj

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
)

// ProjectModel ...
type ProjectModel struct {
	Pth           string
	Name          string
	SDK           string
	SharedSchemes []SchemeModel
	Targets       []TargetModel
}

// NewProject ...
func NewProject(xcodeprojPth string) (ProjectModel, error) {
	project := ProjectModel{
		Pth:  xcodeprojPth,
		Name: strings.TrimSuffix(filepath.Base(xcodeprojPth), filepath.Ext(xcodeprojPth)),
	}

	// SDK
	pbxprojPth := filepath.Join(xcodeprojPth, "project.pbxproj")

	if exist, err := pathutil.IsPathExists(pbxprojPth); err != nil {
		return ProjectModel{}, err
	} else if !exist {
		return ProjectModel{}, fmt.Errorf("Project descriptor not found at: %s", pbxprojPth)
	}

	sdk, err := GetBuildConfigSDKRoot(pbxprojPth)
	if err != nil {
		return ProjectModel{}, err
	}

	project.SDK = sdk

	// Shared Schemes
	schemes, err := ProjectSharedSchemes(xcodeprojPth)
	if err != nil {
		return ProjectModel{}, err
	}

	project.SharedSchemes = schemes

	// Targets
	targets, err := ProjectTargets(xcodeprojPth)
	if err != nil {
		return ProjectModel{}, err
	}

	project.Targets = targets

	return project, nil
}
