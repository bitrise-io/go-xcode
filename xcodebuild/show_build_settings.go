package xcodebuild

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

const (
	toolName = "xcodebuild"

	showBuildSettingsFlag = "-showBuildSettings"
	jsonFlag              = "-json"
	projectFlag           = "-project"
	workspaceFlag         = "-workspace"
	targetFlag            = "-target"
	schemeFlag            = "-scheme"
	configurationFlag     = "-configuration"

	xcworkspaceExtension = ".xcworkspace"
)

// BuildSettingsProvider returns effective build settings using xcodebuild -showBuildSettings.
type BuildSettingsProvider interface {
	TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error)
	SchemeBuildSettings(projectPath, scheme, configuration string, extraArgs ...string) (serialized.Object, error)
}

type showBuildSettingsProvider struct {
	commandFactory command.Factory
	logger         log.Logger
}

// NewShowBuildSettingsProvider returns a BuildSettingsProvider backed by xcodebuild.
func NewShowBuildSettingsProvider(commandFactory command.Factory, logger log.Logger) BuildSettingsProvider {
	return showBuildSettingsProvider{
		commandFactory: commandFactory,
		logger:         logger,
	}
}

// TargetBuildSettings returns the effective build settings of one target.
func (p showBuildSettingsProvider) TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error) {
	return p.run(showBuildSettingsArgs(projectPath, targetFlag, target, configuration, extraArgs), target)
}

// SchemeBuildSettings returns the effective build settings of one scheme: its first
// target's, which is the scheme's main target.
func (p showBuildSettingsProvider) SchemeBuildSettings(projectPath, scheme, configuration string, extraArgs ...string) (serialized.Object, error) {
	return p.run(showBuildSettingsArgs(projectPath, schemeFlag, scheme, configuration, extraArgs), "")
}

// run reads the settings from -json output on stdout; stderr carries xcodebuild's
// warnings, which would break the JSON. If stdout is not JSON, it falls back to the text
// parser on the whole output.
func (p showBuildSettingsProvider) run(args []string, target string) (serialized.Object, error) {
	var stdout, stderr bytes.Buffer
	cmd := p.commandFactory.Create(toolName, args, &command.Opts{Stdout: &stdout, Stderr: &stderr})

	// Logged at normal level, as in v1, so the command shows up in step logs.
	p.logger.TPrintf("Reading build settings...")
	p.logger.TDonef("$ %s", cmd.PrintableCommandArgs())

	if err := cmd.Run(); err != nil {
		out := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, commandError{
				message: fmt.Sprintf("%s failed with exit status %d: %s", cmd.PrintableCommandArgs(), exitErr.ExitCode(), out),
				err:     err,
			}
		}
		return nil, fmt.Errorf("failed to run %s: %w", cmd.PrintableCommandArgs(), err)
	}

	p.logger.TPrintf("Read target settings.")

	settings, err := parseShowBuildSettingsJSON(stdout.Bytes(), target)
	if err != nil {
		p.logger.Debugf("Reading build settings as JSON failed (%s), reading the text output", err)
		return parseShowBuildSettingsOutput(stdout.String() + "\n" + stderr.String())
	}
	return settings, nil
}

func showBuildSettingsArgs(projectPath, nameFlag, name, configuration string, extraArgs []string) []string {
	var args []string

	if projectPath != "" {
		containerFlag := projectFlag
		if filepath.Ext(projectPath) == xcworkspaceExtension {
			containerFlag = workspaceFlag
		}
		args = append(args, containerFlag, projectPath)
	}

	if name != "" {
		args = append(args, nameFlag, name)
	}

	if configuration != "" {
		args = append(args, configurationFlag, configuration)
	}

	args = append(args, showBuildSettingsFlag, jsonFlag)

	return append(args, extraArgs...)
}

// showBuildSettingsEntry is one element of `xcodebuild -showBuildSettings -json`: one per
// target the command covers, a scheme's main target first.
type showBuildSettingsEntry struct {
	Target        string            `json:"target"`
	BuildSettings map[string]string `json:"buildSettings"`
}

// parseShowBuildSettingsJSON returns the settings of target, or of the first entry when
// target is empty or not listed. Values are trimmed: xcodebuild pads list values with
// spaces (" @executable_path/Frameworks"), which the text output never showed. Quotes
// inside values stay, unlike in the text parser (-framework "SnapKit").
func parseShowBuildSettingsJSON(out []byte, target string) (serialized.Object, error) {
	var entries []showBuildSettingsEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, err
	}

	settings := serialized.Object{}
	if len(entries) == 0 {
		// a Swift package scheme lists no targets, as its text output lists no settings
		return settings, nil
	}

	entry := entries[0]
	for _, e := range entries {
		if target != "" && e.Target == target {
			entry = e
			break
		}
	}

	for key, value := range entry.BuildSettings {
		settings[key] = strings.TrimSpace(value)
	}
	return settings, nil
}

// parseShowBuildSettingsOutput keeps the first occurrence of a repeated key, which for a
// multi-target scheme is the main target (v1 fix 0c84f25). ReadLine is used because values can
// exceed bufio.Scanner's line limit.
func parseShowBuildSettingsOutput(out string) (serialized.Object, error) {
	settings := serialized.Object{}

	reader := bufio.NewReader(strings.NewReader(out))
	var buffer bytes.Buffer

	for {
		fragment, isPrefix, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		buffer.Write(fragment)

		if isPrefix {
			continue
		}

		line := strings.TrimSpace(buffer.String())
		buffer.Reset()

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"`)

		if _, exists := settings[key]; !exists {
			settings[key] = value
		}
	}

	return settings, nil
}

// commandError reports a failed xcodebuild run with its output, which explains the failure better
// than the exit status, while still unwrapping to the underlying error.
type commandError struct {
	message string
	err     error
}

func (e commandError) Error() string { return e.message }

func (e commandError) Unwrap() error { return e.err }
