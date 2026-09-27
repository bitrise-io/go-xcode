// Package buildsettings reads a project's effective build settings with
// `xcodebuild -showBuildSettings -json`.
package buildsettings

import (
	"encoding/json"
	"strings"
)

// Settings is one target's effective build settings, by setting name.
type Settings map[string]string

// Parse reads the JSON that `xcodebuild -showBuildSettings -json` prints on stdout: the
// settings of every target the query covers, by target name. Values are trimmed; xcodebuild
// pads list values with spaces (" @executable_path/Frameworks").
func Parse(out []byte) (map[string]Settings, error) {
	var entries []struct {
		Target        string            `json:"target"`
		BuildSettings map[string]string `json:"buildSettings"`
	}
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, err
	}

	targets := make(map[string]Settings, len(entries))
	for _, e := range entries {
		settings := make(Settings, len(e.BuildSettings))
		for key, value := range e.BuildSettings {
			settings[key] = strings.TrimSpace(value)
		}
		targets[e.Target] = settings
	}
	return targets, nil
}
