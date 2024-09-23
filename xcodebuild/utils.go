package xcodebuild

import (
	"bufio"
	"bytes"
	"io"
	"slices"
	"strings"
)

func ParseShowBuildSettingsCommandOutput(out string) (map[string]any, error) {
	settings := map[string]any{}
	var buffer bytes.Buffer
	reader := bufio.NewReader(strings.NewReader(out))

	for {
		b, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		lineFragment := string(b)
		buffer.WriteString(lineFragment)

		// isPrefix is set to false once a full line has been read
		if isPrefix == false {
			line := strings.TrimSpace(buffer.String())

			if split := strings.Split(line, "="); len(split) > 1 {
				key := strings.TrimSpace(split[0])
				value := strings.TrimSpace(strings.Join(split[1:], "="))
				value = strings.Trim(value, `"`)

				settings[key] = value
			}

			buffer.Reset()
		}
	}

	return settings, nil
}

type parseAdditionalArgsResult struct {
	options       map[string]any
	actions       []string
	buildSettings map[string]string
	userDefault   map[string]string
}

func (r parseAdditionalArgsResult) toArgs() []string {
	var args []string
	for k, v := range r.options {
		if v == true {
			args = append(args, k)
		} else {
			args = append(args, k, v.(string))
		}
	}

	args = append(args, r.actions...)

	for k, v := range r.buildSettings {
		args = append(args, k+"="+v)
	}

	for k, v := range r.userDefault {
		args = append(args, k+"="+v)
	}

	return args
}

/*
xcodebuild [-project <projectname>] [[-target <targetname>]...|-alltargets] [-configuration <configurationname>] [-arch <architecture>]... [-sdk [<sdkname>|<sdkpath>]] [-showBuildSettings [-json]] [<buildsetting>=<value>]... [<buildaction>]...
xcodebuild [-project name.xcodeproj] -scheme schemename [[-destination destinationspecifier] ...] [-destination-timeout value] [-configuration configurationname] [-sdk [sdkfullpath | sdkname]] [action ...] [buildsetting=value ...] [-userdefault=value ...]
*/
func parseAdditionalArgs(args []string) parseAdditionalArgsResult {
	options := map[string]any{}
	var actions []string
	buildSettings := map[string]string{}
	userDefault := map[string]string{}

	i := 0
	for i < len(args) {
		arg := args[0]

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

	return parseAdditionalArgsResult{
		options:       options,
		actions:       actions,
		buildSettings: buildSettings,
		userDefault:   userDefault,
	}
}

func mergeAdditionalArgs(args1, args2 parseAdditionalArgsResult) parseAdditionalArgsResult {
	options := map[string]any{}
	for k, v := range args1.options {
		options[k] = v
	}
	for k, v := range args2.options {
		options[k] = v
	}

	actions := append([]string{}, args1.actions...)
	for _, action := range args2.actions {
		if slices.Index(actions, action) == -1 {
			actions = append(actions, action)
		}
	}

	buildSettings := map[string]string{}
	for k, v := range args1.buildSettings {
		buildSettings[k] = v
	}
	for k, v := range args2.buildSettings {
		buildSettings[k] = v
	}

	userDefault := map[string]string{}
	for k, v := range args1.userDefault {
		userDefault[k] = v
	}
	for k, v := range args2.userDefault {
		userDefault[k] = v
	}

	return parseAdditionalArgsResult{
		options:       options,
		actions:       actions,
		buildSettings: buildSettings,
		userDefault:   userDefault,
	}
}

func isOption(arg string) bool {
	return strings.HasPrefix(arg, "-")
}

func isAction(arg string) bool {
	var knownActions = []string{
		"build",
		"build-for-testing",
		"analyze",
		"archive",
		"test",
		"test-without-building",
		"docbuild",
		"installsrc",
		"install",
		"clean",
	}
	return slices.Index(knownActions, arg) != -1
}

func isBuildSetting(arg string) bool {
	return !strings.HasPrefix(arg, "-") && strings.Contains(arg, "=")
}

func isUserDefaults(arg string) bool {
	return strings.HasPrefix(arg, "-") && strings.Contains(arg, "=")
}
