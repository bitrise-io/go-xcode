package zip

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

const testEntryName = "Payload/Test.app/Info.plist"

func TestStdlibRead_ReadFile_limitsFileSize(t *testing.T) {
	t.Run("reads a file of the maximum size", func(t *testing.T) {
		reader, err := NewStdlibRead(createArchive(t, testEntryName, MaxFileSize), log.NewLogger())
		require.NoError(t, err)
		defer func() { require.NoError(t, reader.Close()) }()

		b, err := reader.ReadFile("Payload/*.app/Info.plist")

		require.NoError(t, err)
		require.Len(t, b, MaxFileSize)
	})

	t.Run("rejects a file over the maximum size", func(t *testing.T) {
		reader, err := NewStdlibRead(createArchive(t, testEntryName, MaxFileSize+1), log.NewLogger())
		require.NoError(t, err)
		defer func() { require.NoError(t, reader.Close()) }()

		_, err = reader.ReadFile("Payload/*.app/Info.plist")

		require.ErrorIs(t, err, ErrFileTooLarge)
		require.False(t, IsErrFormat(err), "a format error makes zip.DefaultReader retry with ditto")
	})
}

func createArchive(t *testing.T, entryName string, size int) string {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), "test.ipa")
	f, err := os.Create(archivePath)
	require.NoError(t, err)

	w := zip.NewWriter(f)
	entry, err := w.Create(entryName)
	require.NoError(t, err)
	_, err = entry.Write(make([]byte, size))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	return archivePath
}
