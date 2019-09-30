package cache

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
)

func Test_xcodeProjectHash(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "normal xcodeproj path",
			path:    "/Users/bitrise/git/sample-swiftpm.xcodeproj",
			want:    "dsvyrfhmubmjkdguolhekiuetuie",
			wantErr: false,
		},
		{
			name:    "normal xcworkspace path",
			path:    "/Users/bitrise/Develop/samples/sample-apps-ios-swiftpm/sample-swiftpm.xcworkspace",
			want:    "domyjojidpnjraaljgmxofiwqhps",
			wantErr: false,
		},
		{
			name:    "Unicode composite character in path",
			path:    "/Users/bitrise/Develop/samples/Gdańsk/sample-swiftpm.xcodeproj",
			want:    "djfhdbzbhhqfklgywrqyqnyflvnl",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := xcodeDerivedDataHash(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("xcodeProjectHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("xcodeProjectHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_xcodeProjectDerivedDataPath(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		want        string
		wantErr     bool
	}{
		{
			name:        "normal xcodeproj path",
			projectPath: "/Users/lpusok/Develop/samples/sample-apps-ios-swiftpm/sample-swiftpm.xcodeproj",
			want:        filepath.Join(pathutil.UserHomeDir(), "Library/Developer/Xcode/DerivedData/sample-swiftpm-atfutdtkzefhykgeccaarxqthpih"),
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := xcodeProjectDerivedDataPath(tt.projectPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("xcodeProjectDerivedDataPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("xcodeProjectDerivedDataPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
