// Package buildsettings reads a project's effective build settings with
// `xcodebuild -showBuildSettings -json`.
package buildsettings

import (
	"encoding/json"
	"strings"
)

// Settings is one target's effective build settings.
type Settings struct {
	Target string
	Action string
	Values map[string]string
}

// Value returns the setting named key.
func (s Settings) Value(key string) (string, bool) {
	value, ok := s.Values[key]
	return value, ok
}

// List holds the settings of every target a query covers. For a scheme the main target
// comes first, then the others it builds (its test targets, for example).
type List []Settings

// Main returns the first target's settings: the queried target, or a scheme's main target.
// A Swift package scheme lists no targets.
func (l List) Main() (Settings, bool) {
	if len(l) == 0 {
		return Settings{}, false
	}
	return l[0], true
}

// Target returns the settings of the target named name.
func (l List) Target(name string) (Settings, bool) {
	for _, s := range l {
		if s.Target == name {
			return s, true
		}
	}
	return Settings{}, false
}

// Parse reads the JSON that `xcodebuild -showBuildSettings -json` prints on stdout.
// Values are trimmed: xcodebuild pads list values with spaces (" @executable_path/Frameworks").
func Parse(out []byte) (List, error) {
	var entries []struct {
		Target        string            `json:"target"`
		Action        string            `json:"action"`
		BuildSettings map[string]string `json:"buildSettings"`
	}
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, err
	}

	list := make(List, 0, len(entries))
	for _, e := range entries {
		values := make(map[string]string, len(e.BuildSettings))
		for key, value := range e.BuildSettings {
			values[key] = strings.TrimSpace(value)
		}
		list = append(list, Settings{Target: e.Target, Action: e.Action, Values: values})
	}
	return list, nil
}
