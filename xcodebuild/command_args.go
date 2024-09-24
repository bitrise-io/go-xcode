package xcodebuild

import (
	"slices"
	"strings"
)

const (
	BuildAction               = "build"
	BuildForTest              = "build-for-testing"
	AnalyzeAction             = "analyze"
	ArchiveAction             = "archive"
	TestAction                = "test"
	TestWithoutBuildingAction = "test-without-building"
	DocBuildAction            = "docbuild"
	InstallSrcAction          = "installsrc"
	InstallAction             = "install"
	CleanAction               = "clean"
)

var knownActions = []string{
	BuildAction,
	BuildForTest,
	AnalyzeAction,
	ArchiveAction,
	TestAction,
	TestWithoutBuildingAction,
	DocBuildAction,
	InstallSrcAction,
	InstallAction,
	CleanAction,
}

// CommandArgs represents additional arguments for an xcodebuild command.
type CommandArgs struct {
	Options       map[string]any
	Actions       []string
	BuildSettings map[string]string
	UserDefault   map[string]string
}

// ToArgs converts CommandArgs to a slice of string command arguments.
func (r CommandArgs) ToArgs() []string {
	var args []string
	for k, v := range r.Options {
		if v == true {
			args = append(args, k)
		} else {
			args = append(args, k, v.(string))
		}
	}

	args = append(args, r.Actions...)

	for k, v := range r.BuildSettings {
		args = append(args, k+"="+v)
	}

	for k, v := range r.UserDefault {
		args = append(args, k+"="+v)
	}

	return args
}

// ParseAdditionalArgs parses additional arguments for an xcodebuild command.
// The additional arguments are expected to be in the following format:
// -<option> <value> -<bool_option> <action> <buildsetting=value> <-userdefault=value>
func ParseAdditionalArgs(args []string) CommandArgs {
	options := map[string]any{}
	var actions []string
	buildSettings := map[string]string{}
	userDefault := map[string]string{}

	i := 0
	for i < len(args) {
		arg := args[i]

		if isOption(arg) {
			if i+1 < len(args) {
				nextArg := args[i+1]
				if !isOption(nextArg) && !isAction(nextArg) && !isBuildSetting(nextArg) && !isUserDefaults(nextArg) {
					options[arg] = nextArg
					i++
				} else {
					options[arg] = true

				}
			} else {
				options[arg] = true
			}
		} else if isAction(arg) {
			actions = append(actions, arg)
		} else if isBuildSetting(arg) {
			split := strings.Split(arg, "=")
			if len(split) > 1 {
				buildSettings[split[0]] = strings.Join(split[1:], "=")
			} else {
				// TODO: handle error
			}
		} else if isUserDefaults(arg) {
			split := strings.Split(arg, "=")
			if len(split) > 1 {
				userDefault[split[0]] = strings.Join(split[1:], "=")
			} else {
				// TODO: handle error
			}
		}

		i++
	}

	return CommandArgs{
		Options:       options,
		Actions:       actions,
		BuildSettings: buildSettings,
		UserDefault:   userDefault,
	}
}

// MergeAdditionalArgs merges two sets of additional arguments.
// The arguments are merged in the following order:
// 1. Options: args2 options overwrite args1 options
// 2. Actions: args2 actions are appended to args1 actions
// 3. Build settings: args2 build settings overwrite args1 build settings
// 4. User defaults: args2 user defaults overwrite args1 user defaults
func MergeAdditionalArgs(args1, args2 CommandArgs) CommandArgs {
	options := map[string]any{}
	for k, v := range args1.Options {
		options[k] = v
	}
	for k, v := range args2.Options {
		options[k] = v
	}

	var actions []string
	if len(args1.Actions) > 0 {
		actions = append(actions, args1.Actions...)
	}
	for _, action := range args2.Actions {
		if slices.Index(actions, action) == -1 {
			actions = append(actions, action)
		}
	}

	buildSettings := map[string]string{}
	for k, v := range args1.BuildSettings {
		buildSettings[k] = v
	}
	for k, v := range args2.BuildSettings {
		buildSettings[k] = v
	}

	userDefault := map[string]string{}
	for k, v := range args1.UserDefault {
		userDefault[k] = v
	}
	for k, v := range args2.UserDefault {
		userDefault[k] = v
	}

	return CommandArgs{
		Options:       options,
		Actions:       actions,
		BuildSettings: buildSettings,
		UserDefault:   userDefault,
	}
}

func isOption(arg string) bool {
	return strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=")
}

func isAction(arg string) bool {
	return slices.Index(knownActions, arg) != -1
}

func isBuildSetting(arg string) bool {
	return !strings.HasPrefix(arg, "-") && strings.Contains(arg, "=")
}

func isUserDefaults(arg string) bool {
	return strings.HasPrefix(arg, "-") && strings.Contains(arg, "=")
}
