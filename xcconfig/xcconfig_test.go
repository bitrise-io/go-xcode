package xcconfig

import (
	"github.com/bitrise-io/go-xcode/v2/xcconfig/mocks"
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
	mockPathProvider := new(mocks.PathProvider)
	mockPathProvider.On("CreateTempDir", "").Return(testTempDir, nil)
	mockFileManager := new(mocks.FileManager)
	mockFileManager.On("Write", expectedPath, testContent, fs.FileMode(0644)).Return(nil)
	xcconfigWriter := NewWriter(mockPathProvider, mockFileManager)

	// When
	path, err := xcconfigWriter.Write(testContent)

	// Then
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPath, path)
	}
}
