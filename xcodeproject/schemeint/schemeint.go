package schemeint

import (
	"github.com/bitrise-io/go-xcode/v2/xcodebuild"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
)

// HasScheme represents a struct that implements Scheme.
type HasScheme interface {
	Scheme(string) (*xcscheme.Scheme, string, error)
}

// Scheme returns the project or workspace scheme by name.
func Scheme(pth string, name string, xcodebuildFactory xcodebuild.Factory) (*xcscheme.Scheme, string, error) {
	var p HasScheme
	var err error
	if xcodeproj.IsXcodeProj(pth) {
		var proj xcodeproj.XcodeProj
		proj, err = xcodeproj.NewFromFile(pth, xcodebuildFactory)
		p = &proj
	} else {
		p, err = xcworkspace.NewFromFile(pth, xcodebuildFactory)
	}
	if err != nil {
		return nil, "", err
	}
	return p.Scheme(name)
}
