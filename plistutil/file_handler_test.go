package plistutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileHandler_roundTripKeepsFormat(t *testing.T) {
	for name, format := range map[string]int{"xml": plist.XMLFormat, "binary": plist.BinaryFormat} {
		t.Run(name, func(t *testing.T) {
			pth := filepath.Join(t.TempDir(), "Info.plist")
			content, err := plist.Marshal(PlistData{"CFBundleVersion": "1"}, format)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(pth, content, 0600))

			handler := NewFileHandler(fileutil.NewFileManager())

			data, readFormat, err := handler.Read(pth)
			require.NoError(t, err)
			assert.Equal(t, format, readFormat)
			assert.Equal(t, "1", data["CFBundleVersion"])

			data["CFBundleVersion"] = "42"
			require.NoError(t, handler.Write(pth, data, readFormat))

			reread, rereadFormat, err := handler.Read(pth)
			require.NoError(t, err)
			assert.Equal(t, format, rereadFormat)
			assert.Equal(t, "42", reread["CFBundleVersion"])

			info, err := os.Stat(pth)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
		})
	}
}

func TestFileHandler_Read_errors(t *testing.T) {
	handler := NewFileHandler(fileutil.NewFileManager())

	_, _, err := handler.Read(filepath.Join(t.TempDir(), "missing.plist"))
	require.Error(t, err)

	malformed := filepath.Join(t.TempDir(), "malformed.plist")
	require.NoError(t, os.WriteFile(malformed, []byte("<plist"), 0644))
	_, _, err = handler.Read(malformed)
	require.ErrorContains(t, err, "malformed.plist")
}
