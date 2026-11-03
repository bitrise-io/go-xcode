package xcodeproj

import (
	"os"
	"os/user"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// BuildSettingsProvider resolves effective build settings: those of one target, and those of a
// scheme as xcodebuild reports them for it. xcodebuild prints one block per target the scheme
// builds, first build action entry first; the provider keeps, for each key, the first block that
// has it, so a key the first target lacks comes from a later one. Selecting one target of a scheme
// needs the -json output (go-xcode #386). projectPath is the .xcodeproj bundle; the path's
// extension tells an implementation which xcodebuild flag to pass, as xcodebuild itself accepts
// only an .xcodeproj after -project.
type BuildSettingsProvider interface {
	TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error)
	SchemeBuildSettings(projectPath, scheme, configuration string, extraArgs ...string) (serialized.Object, error)
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

// DirCreator creates directories, including missing parents.
type DirCreator interface {
	MkdirAll(path string, perm os.FileMode) error
}

type osDirCreator struct{}

// NewDirCreator returns a DirCreator backed by os.MkdirAll.
func NewDirCreator() DirCreator {
	return osDirCreator{}
}

// MkdirAll creates path and its missing parents with perm, before umask.
func (osDirCreator) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
