package buildsettings

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodecommand"
)

// Query selects the settings to read: a target or a scheme of a project or workspace.
type Query struct {
	ProjectPath       string // .xcodeproj or .xcworkspace; a workspace needs a Scheme
	Target            string
	Scheme            string
	Configuration     string
	AdditionalOptions []string // build settings and package flags that change the result
}

// Reader reads effective build settings.
type Reader interface {
	Read(query Query) (List, error)
}

type reader struct {
	commandFactory command.Factory
	logger         log.Logger
}

// NewReader returns a Reader that runs xcodebuild.
func NewReader(commandFactory command.Factory, logger log.Logger) Reader {
	return reader{commandFactory: commandFactory, logger: logger}
}

// Read runs `xcodebuild -showBuildSettings -json` and parses its stdout. xcodebuild's
// warnings go to stderr, which is logged at debug level and included in the error of a
// failed run.
func (r reader) Read(query Query) (List, error) {
	cmd, err := xcodecommand.ShowBuildSettings(xcodecommand.ShowBuildSettingsParams{
		ProjectPath:       query.ProjectPath,
		Target:            query.Target,
		Scheme:            query.Scheme,
		Configuration:     query.Configuration,
		AdditionalOptions: query.AdditionalOptions,
	})
	if err != nil {
		return nil, err
	}
	// The step reports its xcodebuild_options on its main command.
	for _, d := range cmd.Diagnostics() {
		r.logger.Debugf("build settings query: %s", d)
	}

	var stdout, stderr bytes.Buffer
	run := cmd.Create(r.commandFactory, &command.Opts{Stdout: &stdout, Stderr: &stderr})

	r.logger.TPrintf("Reading build settings...")
	r.logger.TDonef("$ %s", run.PrintableCommandArgs())

	runErr := run.Run()
	if warnings := strings.TrimSpace(stderr.String()); warnings != "" {
		r.logger.Debugf("xcodebuild stderr:\n%s", warnings)
	}
	if runErr != nil {
		return nil, fmt.Errorf("%s: %w\n%s", run.PrintableCommandArgs(), runErr, strings.TrimSpace(stderr.String()+"\n"+stdout.String()))
	}

	list, err := Parse(stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s printed no build settings JSON: %w", run.PrintableCommandArgs(), err)
	}
	r.logger.TPrintf("Read build settings of %d target(s).", len(list))
	return list, nil
}
