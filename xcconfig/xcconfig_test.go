package xcconfig

import (
	mockfileutil "github.com/bitrise-io/go-utils/fileutil/mocks"
	mockpathutil "github.com/bitrise-io/go-utils/pathutil/mocks"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"path/filepath"
	"testing"
)

func Test_WhenWritingXCConfigContent_ThenItShouldReturnFilePath(t *testing.T) {
	// Given
	testContent := "TEST"
	testTempDir := "temp_dir"
	expectedPath := filepath.Join(testTempDir, "temp.xcconfig")
	mockPathProvider := new(mockpathutil.PathProvider)
	mockPathProvider.On("CreateTempDir", "").Return(testTempDir, nil)
	mockFileManager := new(mockfileutil.FileManager)
	mockFileManager.On("Write", expectedPath, testContent, fs.FileMode(0644)).Return(nil)
	xcconfigWriter := NewWriter(mockPathProvider, mockFileManager)

	// When
	path, err := xcconfigWriter.Write(testContent)

	// Then
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPath, path)
	}
}
