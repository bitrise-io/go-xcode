package xcodebuild

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAdditionalArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want CommandArgs
	}{
		{
			name: "Empty args parsed into CommandArgs with maps initialized",
			args: []string{},
			want: CommandArgs{
				Options:       map[string]any{},
				Actions:       nil,
				BuildSettings: map[string]string{},
				UserDefault:   map[string]string{},
			},
		},
		{
			name: "Command options, actions, build settings, and user defaults parsed into CommandArgs",
			args: []string{"-project", "name.xcodeproj", "-target", "targetname", "-configuration", "configurationname", "build", "buildsetting=value", "-userdefault=value"},
			want: CommandArgs{
				Options:       map[string]any{"-project": "name.xcodeproj", "-target": "targetname", "-configuration": "configurationname"},
				Actions:       []string{"build"},
				BuildSettings: map[string]string{"buildsetting": "value"},
				UserDefault:   map[string]string{"-userdefault": "value"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAdditionalArgs(tt.args)
			require.Equal(t, tt.want, got)

			gotArgs := got.ToArgs()
			require.ElementsMatch(t, tt.args, gotArgs)
		})
	}
}

func TestMergeAdditionalArgs(t *testing.T) {
	tests := []struct {
		name  string
		args1 CommandArgs
		args2 CommandArgs
		want  CommandArgs
	}{
		{
			name:  "Merges empty arguments",
			args1: CommandArgs{},
			args2: CommandArgs{},
			want: CommandArgs{
				Options:       map[string]any{},
				Actions:       nil,
				BuildSettings: map[string]string{},
				UserDefault:   map[string]string{},
			},
		},
		{
			name: "Merges two sets of additional arguments",
			args1: CommandArgs{
				Options:       map[string]any{"-project": "name.xcodeproj", "-target": "targetname"},
				Actions:       []string{"build"},
				BuildSettings: map[string]string{"buildsetting": "value"},
				UserDefault:   map[string]string{"-userdefault": "value"},
			},
			args2: CommandArgs{
				Options:       map[string]any{"-target": "newtargetname", "-configuration": "configurationname"},
				Actions:       []string{"build", "test"},
				BuildSettings: map[string]string{"buildsetting": "newvalue", "buildsetting1": "value"},
				UserDefault:   map[string]string{"-userdefault": "newvalue", "-userdefault1": "value"},
			},
			want: CommandArgs{
				Options:       map[string]any{"-project": "name.xcodeproj", "-target": "newtargetname", "-configuration": "configurationname"},
				Actions:       []string{"build", "test"},
				BuildSettings: map[string]string{"buildsetting": "newvalue", "buildsetting1": "value"},
				UserDefault:   map[string]string{"-userdefault": "newvalue", "-userdefault1": "value"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeAdditionalArgs(tt.args1, tt.args2)
			require.Equal(t, tt.want, got)
		})
	}
}
