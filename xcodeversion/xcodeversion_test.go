package xcodeversion

import (
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/require"
)

func Test_reader_GetVersion(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		wantedVersion Version
		wantErr       bool
	}{
		{
			name:   "Plain output",
			output: "Xcode 8.2.1\nBuild version 8C1002",
			wantedVersion: Version{
				Version:      "Xcode 8.2.1",
				BuildVersion: "8C1002",
				Major:        8,
				Minor:        2,
			},
		},
		{
			name:   "Plain output with more spaces",
			output: "Xcode  8.2.1\nBuild version 8C1002",
			wantedVersion: Version{
				Version:      "Xcode  8.2.1",
				BuildVersion: "8C1002",
				Major:        8,
				Minor:        2,
			},
		},
		{
			name:   "major version only",
			output: "Xcode 16 beta6\nBuild version 16A5230g",
			wantedVersion: Version{
				Version:      "Xcode 16 beta6",
				BuildVersion: "16A5230g",
				Major:        16,
				Minor:        0,
			},
		},
		{
			name:          "version not found",
			output:        "Xcode unexpected string\nBuild version unexped	ted string",
			wantedVersion: Version{},
			wantErr:       true,
		},
		{
			name:   "build version not found",
			output: "Xcode 16.2.1\nunexpected build version",
			wantedVersion: Version{
				Version:      "Xcode 16.2.1",
				Major:        16,
				Minor:        2,
				BuildVersion: "unknown",
			},
			wantErr: false,
		},
		{
			name: "Warnings in output (Xcode 13.2.1 bug)",
			output: `objc[82434]: Class AMSupportURLConnectionDelegate is implemented in both /usr/lib/libauthinstall.dylib (0x212da2b90) and /Library/Apple/System/Library/PrivateFrameworks/MobileDevice.framework/Versions/A/MobileDevice (0x1046dc2c8). One of the two will be used. Which one is undefined.
objc[82434]: Class AMSupportURLSession is implemented in both /usr/lib/libauthinstall.dylib (0x212da2be0) and /Library/Apple/System/Library/PrivateFrameworks/MobileDevice.framework/Versions/A/MobileDevice (0x1046dc318). One of the two will be used. Which one is undefined.
Xcode 13.2.1
Build version 13C100`,
			wantedVersion: Version{
				Version:      "Xcode 13.2.1",
				BuildVersion: "13C100",
				Major:        13,
				Minor:        2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommand := new(mocks.Command)
			mockCommand.On("RunAndReturnTrimmedCombinedOutput").Return(tt.output, nil)
			mockCommandFactory := new(mocks.CommandFactory)
			mockCommandFactory.On("Create", "xcodebuild", []string{"-version"}, &command.Opts{}).Return(mockCommand)

			xcodeVersionProvider := &reader{
				commandFactory: mockCommandFactory,
			}

			got, err := xcodeVersionProvider.GetVersion()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCommandFactory.AssertExpectations(t)
			mockCommand.AssertExpectations(t)
			require.Equal(t, tt.wantedVersion, got)
		})
	}
}

func TestVersion_IsGreaterThanOrEqualTo(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		major   int64
		minor   int64
		want    bool
	}{
		{
			name:    "greater major",
			version: Version{Major: 8, Minor: 2},
			major:   7,
			minor:   2,
			want:    true,
		},
		{
			name:    "lesser major",
			version: Version{Major: 18, Minor: 2},
			major:   17,
			minor:   2,
			want:    true,
		},
		{
			name:    "equal major, greater minor",
			version: Version{Major: 8, Minor: 2},
			major:   8,
			minor:   1,
			want:    true,
		},
		{
			name:    "equal major, equal minor",
			version: Version{Major: 8, Minor: 2},
			major:   8,
			minor:   2,
			want:    true,
		},
		{
			name:    "equal major, lesser minor",
			version: Version{Major: 8, Minor: 2},
			major:   8,
			minor:   3,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.version.IsGreaterThanOrEqualTo(tt.major, tt.minor))
		})
	}
}
