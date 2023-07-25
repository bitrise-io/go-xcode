package destination

import (
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func Test_isRuntimeSupportedByXcode(t *testing.T) {
	tests := []struct {
		name            string
		runtimePlatform string
		runtimeVersion  *version.Version
		xcodeVersion    xcodeversion.Version
		want            bool
	}{
		{
			name:            "iOS 16 on Xcode 15",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("16.4")),
			xcodeVersion:    xcodeversion.Version{MajorVersion: 15},
			want:            true,
		},
		{
			name:            "iOS 16 on unknown Xcode version",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("16.4")),
			xcodeVersion:    xcodeversion.Version{MajorVersion: 3}, // unknown version
			want:            true,
		},
		{
			name:            "tvOS 17 on Xcode 14",
			runtimePlatform: "tvOS",
			runtimeVersion:  version.Must(version.NewVersion("17")),
			xcodeVersion:    xcodeversion.Version{MajorVersion: 14},
			want:            false,
		},
		{
			name:            "unknown platform",
			runtimePlatform: "walletOS",
			runtimeVersion:  version.Must(version.NewVersion("1")),
			xcodeVersion:    xcodeversion.Version{MajorVersion: 15},
			want:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRuntimeSupportedByXcode(tt.runtimePlatform, tt.runtimeVersion, tt.xcodeVersion)
			require.Equal(t, tt.want, got)
		})
	}
}
