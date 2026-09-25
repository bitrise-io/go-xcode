package xcodeproj

import (
	"os/user"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildSettingsProvider resolves the effective build settings of a target.
type BuildSettingsProvider interface {
	TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error)
}

// UserProvider supplies the current OS user name, which appears in user scheme paths.
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
