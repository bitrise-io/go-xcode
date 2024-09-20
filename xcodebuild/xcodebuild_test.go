package xcodebuild

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandOptions_toCommandOptions(t *testing.T) {
	tests := []struct {
		name    string
		options CommandOptions
		want    []string
	}{
		{
			name: "Creates xcodebuild options",
			options: CommandOptions{
				Project: "project",
				Scheme:  "scheme",
			},
			want: []string{"-project", "project", "-scheme", "scheme"},
		},
		{
			name: "Adds options for boolean true values",
			options: CommandOptions{
				Project:                  "project",
				AllowProvisioningUpdates: true,
			},
			want: []string{"-project", "project", "-allowProvisioningUpdates"},
		},
		{
			name: "Doesn't add options for boolean false values",
			options: CommandOptions{
				Project:                  "project",
				AllowProvisioningUpdates: false,
			},
			want: []string{"-project", "project"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.toCommandOptions()
			require.ElementsMatch(t, tt.want, got)
		})
	}
}
