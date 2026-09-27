package buildsettings

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodecommand"
)

// Reader reads effective build settings. additionalOptions are build settings and package
// flags that change the result (from the step's xcodebuild_options).
type Reader interface {
	// Target returns a project target's settings: `-project <projectPath> -target <target>`.
	Target(projectPath, target, configuration string, additionalOptions ...string) (Settings, error)
	// SchemeTarget returns the settings of target as scheme builds it:
	// `-workspace|-project <projectPath> -scheme <scheme>`, then target's entry.
	SchemeTarget(projectPath, scheme, target, configuration string, additionalOptions ...string) (Settings, error)
	// Scheme returns the settings of every target scheme builds, by target name.
	Scheme(projectPath, scheme, configuration string, additionalOptions ...string) (map[string]Settings, error)
}

type reader struct {
	commandFactory command.Factory
	logger         log.Logger
}

// NewReader returns a Reader that runs xcodebuild.
func NewReader(commandFactory command.Factory, logger log.Logger) Reader {
	return reader{commandFactory: commandFactory, logger: logger}
}

func (r reader) Target(projectPath, target, configuration string, additionalOptions ...string) (Settings, error) {
	if filepath.Ext(projectPath) != ".xcodeproj" {
		return nil, fmt.Errorf("reading a target's build settings needs an .xcodeproj, got %s", projectPath)
	}
	targets, err := r.read(xcodecommand.ShowBuildSettingsParams{ProjectPath: projectPath, Target: target, Configuration: configuration, AdditionalOptions: additionalOptions})
	if err != nil {
		return nil, err
	}
	return pick(targets, target, "project "+projectPath)
}

func (r reader) SchemeTarget(projectPath, scheme, target, configuration string, additionalOptions ...string) (Settings, error) {
	targets, err := r.Scheme(projectPath, scheme, configuration, additionalOptions...)
	if err != nil {
		return nil, err
	}
	return pick(targets, target, "scheme "+scheme)
}

func (r reader) Scheme(projectPath, scheme, configuration string, additionalOptions ...string) (map[string]Settings, error) {
	if ext := filepath.Ext(projectPath); ext != ".xcodeproj" && ext != ".xcworkspace" {
		// A Swift package has no build settings to read: xcodebuild prints none for its schemes.
		return nil, fmt.Errorf("reading build settings needs an .xcodeproj or .xcworkspace, got %s", projectPath)
	}
	return r.read(xcodecommand.ShowBuildSettingsParams{ProjectPath: projectPath, Scheme: scheme, Configuration: configuration, AdditionalOptions: additionalOptions})
}

func pick(targets map[string]Settings, target, where string) (Settings, error) {
	if settings, ok := targets[target]; ok {
		return settings, nil
	}
	names := make([]string, 0, len(targets))
	for name := range targets {
		names = append(names, name)
	}
	slices.Sort(names)
	return nil, fmt.Errorf("no build settings for target %s in %s; it has settings for: %s", target, where, strings.Join(names, ", "))
}

// read runs `xcodebuild -showBuildSettings -json` and parses its stdout. xcodebuild's
// warnings go to stderr, which is logged at debug level and included in the error of a
// failed run.
func (r reader) read(params xcodecommand.ShowBuildSettingsParams) (map[string]Settings, error) {
	cmd, err := xcodecommand.ShowBuildSettings(params)
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

	targets, err := Parse(stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s printed no build settings JSON: %w", run.PrintableCommandArgs(), err)
	}
	r.logger.TPrintf("Read build settings of %d target(s).", len(targets))
	return targets, nil
}
