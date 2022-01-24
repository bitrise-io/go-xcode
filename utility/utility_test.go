package utility

import (
	"testing"

	"github.com/bitrise-io/go-xcode/models"
	"github.com/stretchr/testify/require"
)

func TestGetXcodeVersionFromXcodebuildOutput(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		wantedVersion models.XcodebuildVersionModel
	}{
		{
			name:   "Plain output",
			output: "Xcode 8.2.1\nBuild version 8C1002",
			wantedVersion: models.XcodebuildVersionModel{
				Version:      "Xcode 8.2.1",
				BuildVersion: "Build version 8C1002",
				MajorVersion: 8,
			},
		},
		{
			name: "Warnings in output (Xcode 13.2.1 bug)",
			output: `objc[82434]: Class AMSupportURLConnectionDelegate is implemented in both /usr/lib/libauthinstall.dylib (0x212da2b90) and /Library/Apple/System/Library/PrivateFrameworks/MobileDevice.framework/Versions/A/MobileDevice (0x1046dc2c8). One of the two will be used. Which one is undefined.
objc[82434]: Class AMSupportURLSession is implemented in both /usr/lib/libauthinstall.dylib (0x212da2be0) and /Library/Apple/System/Library/PrivateFrameworks/MobileDevice.framework/Versions/A/MobileDevice (0x1046dc318). One of the two will be used. Which one is undefined.
Xcode 13.2.1
Build version 13C100`,
			wantedVersion: models.XcodebuildVersionModel{
				Version:      "Xcode 13.2.1",
				BuildVersion: "Build version 13C100",
				MajorVersion: 13,
			},
		},
	}

	for _, tt := range tests {
		currentVersion, err := getXcodeVersionFromXcodebuildOutput(tt.output)
		require.NoError(t, err)
		require.Equal(t, tt.wantedVersion, currentVersion)
	}
}
