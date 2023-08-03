package xcodecommand

import (
	"github.com/hashicorp/go-version"
)

type Output struct {
	RawOut   []byte
	ExitCode int
}

type Runner interface {
	CheckInstall() (*version.Version, error)
	Run(workDir string, xcodebuildOpts []string, logFormatterOpts []string) (Output, error)
}
