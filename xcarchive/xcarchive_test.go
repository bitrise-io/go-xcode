package xcarchive

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

const (
	tempDirName  = "__artifacts__"
	DSYMSDirName = "dSYMs"
)

func TestIsMacOS(t *testing.T) {
	tests := []struct {
		name     string
		archPath string
		want     bool
		wantErr  bool
	}{
		{
			name:     "macOS",
			archPath: filepath.Join(sampleRepoPath(t), "archives/macos.xcarchive"),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "iOS",
			archPath: filepath.Join(sampleRepoPath(t), "archives/ios.xcarchive"),
			want:     false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathChecker := pathutil.NewPathChecker()
			logger := log.NewLogger()
			archiveReader := NewArchiveReader(pathChecker, logger)
			got, err := archiveReader.IsMacOS(tt.archPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsMacOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsMacOS() = %v, want %v", got, tt.want)
			}
		})
	}
}
