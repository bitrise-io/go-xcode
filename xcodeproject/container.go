// Package xcodeproject is the common view of Xcode projects and workspaces: what a path to either
// gives a caller that does not care which of the two it holds.
package xcodeproject

import (
	"path/filepath"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
)

// Container is an opened Xcode project or workspace, so a caller never needs to know which of the
// two it has.
type Container interface {
	// Scheme returns the named scheme and the path of the bundle holding it: for a workspace, either
	// the workspace itself or one of the projects it references.
	Scheme(name string) (xcscheme.Scheme, string, error)
	// Project returns the project at pth, an absolute .xcodeproj path, such as the one a scheme's
	// build action entry references. A project the container already opened, itself or while
	// searching schemes, is returned as it is; any other is opened and kept. So each project is
	// parsed once.
	Project(pth string) (*xcodeproj.XcodeProj, error)
}

// schemeSource looks up a scheme by name. Both *xcodeproj.XcodeProj and *xcworkspace.Workspace
// implement it.
type schemeSource interface {
	Scheme(name string) (xcscheme.Scheme, string, error)
}

// container is the Container over either a project or a workspace: it keeps what both offer, so no
// method needs to know which of the two it holds.
type container struct {
	schemes  schemeSource
	projects projectSource
}

// Scheme implements Container.
func (c *container) Scheme(name string) (xcscheme.Scheme, string, error) {
	return c.schemes.Scheme(name)
}

// Project implements Container.
func (c *container) Project(pth string) (*xcodeproj.XcodeProj, error) {
	return c.projects.Project(pth)
}

// projectSource opens projects by path, keeping each one it opened. *xcworkspace.Workspace
// implements it for the projects it references.
type projectSource interface {
	Project(pth string) (*xcodeproj.XcodeProj, error)
}

// projectCache is the projectSource of a container opened as a project: the project itself, and any
// other project, such as a sub-project one of its schemes builds, opened on first use and kept.
type projectCache struct {
	factory  xcodeproj.Factory
	projects map[string]*xcodeproj.XcodeProj
}

func newProjectCache(factory xcodeproj.Factory, project *xcodeproj.XcodeProj) *projectCache {
	return &projectCache{
		factory:  factory,
		projects: map[string]*xcodeproj.XcodeProj{filepath.Clean(project.Path): project},
	}
}

// Project implements projectSource.
func (c *projectCache) Project(pth string) (*xcodeproj.XcodeProj, error) {
	if project, ok := c.projects[filepath.Clean(pth)]; ok {
		return project, nil
	}

	project, err := c.factory.Open(pth)
	if err != nil {
		return nil, err
	}

	c.projects[filepath.Clean(pth)] = project
	c.projects[filepath.Clean(project.Path)] = project

	return project, nil
}
