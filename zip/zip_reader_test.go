package zip

import (
	"io"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/internals/zip"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/require"
)

func TestReader_ReadFile(t *testing.T) {
	tests := []struct {
		name           string
		zipReader      zip.ReadCloser
		targetPathGlob string
		want           []byte
		wantErr        string
	}{
		{
			name: "Exact match",
			zipReader: createMockZipReadCloser([]string{
				"Payload/watch-test.app/Base.lproj/LaunchScreen.storyboardc/Info.plist",
				"Payload/watch-test.app/Base.lproj/Main.storyboardc/Info.plist",
				"Payload/watch-test.app/Info.plist",
				"Payload/watch-test.app/Watch/watch-test WatchKit App.app/Info.plist",
				"Payload/watch-test.app/Watch/watch-test WatchKit App.app/PlugIns/watch-test WatchKit Extension.appex/Info.plist",
			},
				"Payload/watch-test.app/Info.plist",
				"expected"),
			targetPathGlob: "Payload/watch-test.app/Info.plist",
			want:           []byte("expected"),
		},
		{
			name: "Glob match - it returns the first match",
			zipReader: createMockZipReadCloser([]string{
				"Payload/watch-test.app/Base.lproj/LaunchScreen.storyboardc/Info.plist",
				"Payload/watch-test.app/Base.lproj/Main.storyboardc/Info.plist",
				"Payload/watch-test.app/Info.plist",
				"Payload/watch-test.app/Watch/watch-test WatchKit App.app/Info.plist",
				"Payload/watch-test.app/Watch/watch-test WatchKit App.app/PlugIns/watch-test WatchKit Extension.appex/Info.plist",
			},
				"Payload/watch-test.app/Info.plist",
				"expected"),
			targetPathGlob: "Payload/*.app/Info.plist",
			want:           []byte("expected"),
		},
		{
			name: "No match",
			zipReader: createMockZipReadCloser([]string{
				"Payload/watch-test.app/Base.lproj/LaunchScreen.storyboardc/Info.plist",
				"Payload/watch-test.app/Base.lproj/Main.storyboardc/Info.plist",
				"Payload/watch-test.app/Info.plist",
				"Payload/watch-test.app/Watch/watch-test WatchKit App.app/Info.plist",
				"Payload/watch-test.app/Watch/watch-test WatchKit App.app/PlugIns/watch-test WatchKit Extension.appex/Info.plist",
			},
				"Payload/watch-test.app/Info.plist",
				"expected"),
			targetPathGlob: "watch-test.app/Info.plist",
			wantErr:        "no file found with pattern: watch-test.app/Info.plist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := Reader{
				zipReader: tt.zipReader,
				logger:    log.NewLogger(),
			}
			got, err := reader.ReadFile(tt.targetPathGlob)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func createMockZipReadCloser(paths []string, expectedPath, expectedContent string) zip.ReadCloser {
	var files []zip.File
	for _, pth := range paths {
		file := new(mocks.ZipFile)
		file.On("Name").Return(pth)
		if pth == expectedPath {
			file.On("Open").Return(io.NopCloser(strings.NewReader(expectedContent)), nil)
		} else {
			file.On("Open").Return(io.NopCloser(strings.NewReader("")), nil)
		}

		files = append(files, file)
	}

	readCloser := new(mocks.ZipReadCloser)
	readCloser.On("Close").Return(nil)
	readCloser.On("Files").Return(files)
	return readCloser
}
