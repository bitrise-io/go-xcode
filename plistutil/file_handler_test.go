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
			require.NoError(t, os.Chmod(pth, 0664))

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
			assert.Equal(t, os.FileMode(0664), info.Mode().Perm(), "an existing file keeps its mode, as in v1")
		})
	}
}

func TestFileHandler_Write_newFileGetsV1Mode(t *testing.T) {
	pth := filepath.Join(t.TempDir(), "New.plist")

	require.NoError(t, NewFileHandler(fileutil.NewFileManager()).Write(pth, PlistData{"A": "1"}, plist.XMLFormat))

	info, err := os.Stat(pth)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

// Writing through a symlink keeps the target's mode and the link itself.
func TestFileHandler_Write_throughSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Info.plist")
	link := filepath.Join(dir, "Link.plist")
	require.NoError(t, os.WriteFile(target, []byte(`<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict/></plist>`), 0600))
	require.NoError(t, os.Chmod(target, 0664))
	require.NoError(t, os.Symlink(target, link))

	require.NoError(t, NewFileHandler(fileutil.NewFileManager()).Write(link, PlistData{"A": "1"}, plist.XMLFormat))

	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0664), info.Mode().Perm())

	linkInfo, err := os.Lstat(link)
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, linkInfo.Mode()&os.ModeSymlink, "the link is still a link")
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
