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
			name:            "iOS 26.1 on Xcode 26.1",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("26.1")),
			xcodeVersion:    xcodeversion.Version{Major: 26, Minor: 1},
			want:            true,
		},
		{
			name:            "iOS 26.2 on Xcode 26.1",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("26.2")),
			xcodeVersion:    xcodeversion.Version{Major: 26, Minor: 1},
			want:            false,
		},
		{
			name:            "iOS 18 on Xcode 26",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("18")),
			xcodeVersion:    xcodeversion.Version{Major: 26, Minor: 1},
			want:            true,
		},
		{
			name:            "iOS 18 on Xcode 16",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("18.2")),
			xcodeVersion:    xcodeversion.Version{Major: 16},
			want:            true,
		},
		{
			name:            "iOS 26 on unknown Xcode version",
			runtimePlatform: "iOS",
			runtimeVersion:  version.Must(version.NewVersion("16.4")),
			xcodeVersion:    xcodeversion.Version{Major: 0}, // unknown version
			want:            true,
		},
		{
			name:            "tvOS 17 on Xcode 14",
			runtimePlatform: "tvOS",
			runtimeVersion:  version.Must(version.NewVersion("17")),
			xcodeVersion:    xcodeversion.Version{Major: 14},
			want:            false,
		},
		{
			name:            "unknown platform",
			runtimePlatform: "walletOS",
			runtimeVersion:  version.Must(version.NewVersion("1")),
			xcodeVersion:    xcodeversion.Version{Major: 15},
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
