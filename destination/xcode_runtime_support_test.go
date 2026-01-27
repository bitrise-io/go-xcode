package destination

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func Test_isRuntimeSupportedByXcode(t *testing.T) {
	tests := []struct {
		name            string
		runtimePlatform string
		runtimeVersion  *version.Version
		xcodeVersion    *version.Version
		want            bool
	}{
		{
			name:            "iOS 26.1 on Xcode 26.1",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("26.1")),
			xcodeVersion:    version.Must(version.NewVersion("26.1")),
			want:            true,
		},
		{
			name:            "iOS 26.2 on Xcode 26.1",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("26.2")),
			xcodeVersion:    version.Must(version.NewVersion("26.1")),
			want:            false,
		},
		{
			name:            "iOS 18 on Xcode 26",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("18")),
			xcodeVersion:    version.Must(version.NewVersion("26.1")),
			want:            true,
		},
		{
			name:            "iOS 27.0 on Xcode 27 beta",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("27.0")),
			xcodeVersion:    version.Must(version.NewVersion("27.0beta1")),
			want:            true,
		},
		{
			name:            "iOS 18 on Xcode 16",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("18.2")),
			xcodeVersion:    version.Must(version.NewVersion("16")),
			want:            true,
		},
		{
			name:            "iOS 26 on unknown Xcode version",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("16.4")),
			xcodeVersion:    version.Must(version.NewVersion("0")), // unknown version
			want:            true,
		},
		{
			name:            "tvOS 17 on Xcode 14",
			runtimePlatform: "tvOS",
			runtimeVersion:  version.Must(version.NewVersion("17")),
			xcodeVersion:    version.Must(version.NewVersion("14")),
			want:            false,
		},
		{
			name:            "unknown platform",
			runtimePlatform: "walletOS",
			runtimeVersion:  version.Must(version.NewVersion("1")),
			xcodeVersion:    version.Must(version.NewVersion("15")),
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
