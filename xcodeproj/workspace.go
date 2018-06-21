package xcodeproj

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// WorkspaceModel ...
type WorkspaceModel struct {
	pth  string
	name string

	Projects       []ProjectModel
	IsPodWorkspace bool
}

// NewPodWorkspace ...
func NewPodWorkspace(pt, name string) WorkspaceModel {
	return WorkspaceModel{
		pth:            pt,
		name:           name,
		IsPodWorkspace: true,
	}
}

// NewWorkspace ...
func NewWorkspace(pth string, includeProjects ...string) (WorkspaceModel, error) {
	workspace := WorkspaceModel{
		pth:  pth,
		name: strings.TrimSuffix(filepath.Base(pth), filepath.Ext(pth)),
	}

	projects, err := WorkspaceProjectReferences(pth)
	if err != nil {
		return WorkspaceModel{}, err
	}

	if len(includeProjects) > 0 {
		filteredProjects := []string{}
		for _, project := range projects {
			for _, includeProject := range includeProjects {
				if project == includeProject {
					filteredProjects = append(filteredProjects, project)
				}
			}
		}
		projects = filteredProjects
	}

	for _, xcodeprojPth := range projects {
		if exist, err := pathutil.IsPathExists(xcodeprojPth); err != nil {
			return WorkspaceModel{}, err
		} else if !exist {
			return WorkspaceModel{}, fmt.Errorf("referred project (%s) not found", xcodeprojPth)
		}

		project, err := NewProject(xcodeprojPth)
		if err != nil {
			return WorkspaceModel{}, err
		}

		workspace.Projects = append(workspace.Projects, project)
	}

	return workspace, nil
}

// Path ...
func (w WorkspaceModel) Path() string {
	return w.pth
}

// SharedSchemes ...
func (w WorkspaceModel) SharedSchemes() []SchemeModel {
	schemes := []SchemeModel{}
	for _, project := range w.Projects {
		schemes = append(schemes, project.SharedSchemes()...)
	}
	return schemes
}

// Targets ...
func (w WorkspaceModel) Targets() []TargetModel {
	targets := []TargetModel{}
	for _, project := range w.Projects {
		targets = append(targets, project.Targets()...)
	}
	return targets
}

// WorkspaceProjectReferences ...
func WorkspaceProjectReferences(workspace string) ([]string, error) {
	projects := []string{}

	workspaceDir := filepath.Dir(workspace)

	xcworkspacedataPth := path.Join(workspace, "contents.xcworkspacedata")
	if exist, err := pathutil.IsPathExists(xcworkspacedataPth); err != nil {
		return []string{}, err
	} else if !exist {
		return []string{}, fmt.Errorf("contents.xcworkspacedata does not exist at: %s", xcworkspacedataPth)
	}

	xcworkspacedataStr, err := fileutil.ReadStringFromFile(xcworkspacedataPth)
	if err != nil {
		return []string{}, err
	}

	xcworkspacedataLines := strings.Split(xcworkspacedataStr, "\n")
	fileRefStart := false
	regexp := regexp.MustCompile(`location = "(.+):(.+).xcodeproj"`)

	for _, line := range xcworkspacedataLines {
		if strings.Contains(line, "<FileRef") {
			fileRefStart = true
			continue
		}

		if fileRefStart {
			fileRefStart = false
			matches := regexp.FindStringSubmatch(line)
			if len(matches) == 3 {
				projectName := matches[2]
				project := filepath.Join(workspaceDir, projectName+".xcodeproj")
				projects = append(projects, project)
			}
		}
	}

	sort.Strings(projects)

	return projects, nil
}
