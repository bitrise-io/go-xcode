package xcworkspace

import (
	"fmt"
	"strings"
)

// FileRef is a file the workspace references, such as a project.
type FileRef struct {
	Location string `xml:"location,attr"`
}

// FileRefType says what a location's path is relative to.
type FileRefType string

const (
	// AbsoluteFileRefType is an absolute path, or a relative one that xcodebuild takes relative to
	// the workspace's directory.
	AbsoluteFileRefType FileRefType = "absolute"
	// GroupFileRefType is relative to the enclosing group, or to the workspace's directory at the
	// top level.
	GroupFileRefType FileRefType = "group"
	// ContainerFileRefType is relative to the workspace's directory, whichever group it is in.
	ContainerFileRefType FileRefType = "container"
	// DeveloperFileRefType is relative to the Xcode Developer directory.
	DeveloperFileRefType FileRefType = "developer"
	// SelfFileRefType is the project an embedded project.xcworkspace belongs to.
	SelfFileRefType FileRefType = "self"
)

// TypeAndPath splits the reference's location, such as "group:App.xcodeproj", into its type and path.
func (f FileRef) TypeAndPath() (FileRefType, string, error) {
	return parseLocation(f.Location)
}

func parseLocation(location string) (FileRefType, string, error) {
	refType, pth, found := strings.Cut(location, ":")
	if !found {
		return "", "", fmt.Errorf("unknown file reference location (%s)", location)
	}

	switch FileRefType(refType) {
	case AbsoluteFileRefType, GroupFileRefType, ContainerFileRefType, DeveloperFileRefType, SelfFileRefType:
		return FileRefType(refType), pth, nil
	default:
		return "", "", fmt.Errorf("unknown file reference type: %s, value: %s", refType, pth)
	}
}
