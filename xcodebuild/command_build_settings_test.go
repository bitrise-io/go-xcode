package xcodebuild

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/stretchr/testify/require"
)

func TestCommandBuildSettings_cmdArgs(t *testing.T) {
	tests := []struct {
		name          string
		buildSettings CommandBuildSettings
		want          []string
	}{
		{
			name: "Creates xcodebuild command args",
			buildSettings: CommandBuildSettings{
				CodeSigningAllowed: pointers.NewBoolPtr(false),
			},
			want: []string{"CODE_SIGNING_ALLOWED=NO"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ElementsMatch(t, tt.want, tt.buildSettings.cmdArgs())
		})
	}
}
