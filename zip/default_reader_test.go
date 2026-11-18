package zip

import (
	stdzip "archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

func TestDefaultReader_ReadFile_stdlibReaderOnly(t *testing.T) {
	const (
		entryName    = "Payload/Test.app/Info.plist"
		entryContent = "content"
	)

	tests := []struct {
		name          string
		invalidHeader bool
		pattern       string
		want          []byte
		wantErr       bool
		wantErrFormat bool
	}{
		{
			name:    "reads the file matching the pattern",
			pattern: "Payload/*.app/Info.plist",
			want:    []byte(entryContent),
		},
		{
			name:    "returns an error when no file matches the pattern",
			pattern: "Payload/*.app/embedded.mobileprovision",
			wantErr: true,
		},
		{
			name:          "returns the format error of an unreadable file",
			invalidHeader: true,
			pattern:       "Payload/*.app/Info.plist",
			wantErr:       true,
			wantErrFormat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATH", "")

			archivePath := createArchive(t, entryName, entryContent, tt.invalidHeader)

			reader, err := NewDefaultReader(archivePath, log.NewLogger())
			require.NoError(t, err)

			for i := 0; i < 2; i++ {
				got, err := reader.ReadFile(tt.pattern)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				if tt.wantErrFormat {
					require.ErrorIs(t, err, stdzip.ErrFormat)
				}
				require.Equal(t, tt.want, got)
			}

			require.NoError(t, reader.Close())
		})
	}
}

func createArchive(t *testing.T, entryName, entryContent string, invalidHeader bool) string {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), "test.ipa")

	f, err := os.Create(archivePath)
	require.NoError(t, err)

	w := stdzip.NewWriter(f)
	entry, err := w.Create(entryName)
	require.NoError(t, err)
	_, err = entry.Write([]byte(entryContent))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	if invalidHeader {
		_, err = f.WriteAt([]byte{0}, 0)
		require.NoError(t, err)
	}
	require.NoError(t, f.Close())

	return archivePath
}
