package filewriter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWriteKeepingMode_newFile(t *testing.T) {
	pth := filepath.Join(t.TempDir(), "missing", "dir", "file")

	require.NoError(t, WriteKeepingMode(fileutil.NewFileManager(), pth, []byte("new"), 0644))

	content, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))

	info, err := os.Stat(pth)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestWriteKeepingMode_existingFileKeepsItsMode(t *testing.T) {
	pth := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(pth, []byte("old"), 0600))
	require.NoError(t, os.Chmod(pth, 0664))

	require.NoError(t, WriteKeepingMode(fileutil.NewFileManager(), pth, []byte("new"), 0644))

	content, err := os.ReadFile(pth)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))

	info, err := os.Stat(pth)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0664), info.Mode().Perm())
}

// Overwriting must not chmod, so an existing file is written with WriteBytes, not Write.
func TestWriteKeepingMode_existingFileIsNotChmodded(t *testing.T) {
	fileManager := mocks.NewFileManager(t)
	fileManager.On("Lstat", "file").Return(nil, nil)
	fileManager.On("WriteBytes", "file", []byte("new")).Return(nil)

	require.NoError(t, WriteKeepingMode(fileManager, "file", []byte("new"), 0644))
	fileManager.AssertNotCalled(t, "Write", mock.Anything, mock.Anything, mock.Anything)
}
