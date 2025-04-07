package xcodeversion

import (
	"fmt"

	"github.com/bitrise-io/go-utils/v2/command"
)

// Version ...
type Version struct {
	Version      string
	BuildVersion string
	MajorVersion int
	Minor        int
}

// Reader ...
type Reader interface {
	GetVersion() (Version, error)
}

type reader struct {
	commandFactory command.Factory
}

// NewXcodeVersionProvider ...
func NewXcodeVersionProvider(commandFactory command.Factory) Reader {
	return &reader{
		commandFactory: commandFactory,
	}
}

// GetVersion ...
func (b *reader) GetVersion() (Version, error) {
	cmd := b.commandFactory.Create("xcodebuild", []string{"-version"}, &command.Opts{})

	outStr, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return Version{}, fmt.Errorf("xcodebuild -version failed: %s, output: %s", err, outStr)
	}

	return getXcodeVersionFromXcodebuildOutput(outStr)
}

func (v Version) IsGreaterThanOrEqualTo(major, minor int) bool {
	if v.MajorVersion > major {
		return true
	}
	if v.MajorVersion == major && v.Minor >= minor {
		return true
	}
	return false
}
