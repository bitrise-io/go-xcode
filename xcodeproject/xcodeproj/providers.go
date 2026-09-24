package xcodeproj

import (
	"os/user"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildSettingsProvider resolves the effective build settings of a target, as Xcode would
// evaluate them: variables expanded and .xcconfig files applied.
//
// Lookup is by target only. Scheme-based lookup was left out on purpose: this package works per
// target, and targets such as app extensions usually have no scheme to look up.
//
// This is deliberately narrow: it describes what this package needs, not what the xcodebuild
// package offers. Implementations run `xcodebuild -showBuildSettings`, which costs seconds, and
// this package does not memoize — a caller that resolves several things for one target will pay
// for each. Memoizing is left to whoever knows how long an answer stays valid.
type BuildSettingsProvider interface {
	TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error)
}

// UserProvider supplies the name of the OS user Xcode runs as. The name appears in the
// user-scheme path (xcuserdata/<username>.xcuserdatad/xcschemes), so injecting it keeps scheme
// discovery reproducible across machines and CI.
type UserProvider interface {
	CurrentUserName() (string, error)
}

type osUserProvider struct{}

// NewUserProvider returns a UserProvider backed by the os/user package.
func NewUserProvider() UserProvider {
	return osUserProvider{}
}

// CurrentUserName returns the name of the user the process runs as.
func (osUserProvider) CurrentUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Username, nil
}
