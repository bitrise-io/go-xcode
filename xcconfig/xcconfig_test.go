package xcconfig

import (
	"errors"
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
	mockPathChecker := new(mocks.PathChecker)
	mockFileManager.On("Write", expectedPath, testContent, fs.FileMode(0644)).Return(nil)
	xcconfigWriter := NewWriter(mockPathProvider, mockFileManager, mockPathChecker)

	// When
	path, err := xcconfigWriter.Write(testContent)

	// Then
	if assert.NoError(t, err) {
		assert.Equal(t, expectedPath, path)
	}
}

func Test_XCConfigInput_NonExistentPathErrors(t *testing.T) {
	// Given
	testContent := "TEST.xcconfig"
	testTempDir := "temp_dir"
	expectedPath := filepath.Join(testTempDir, "temp.xcconfig")
	mockPathProvider := new(mocks.PathProvider)
	mockPathProvider.On("CreateTempDir", "").Return(testTempDir, nil)
	mockFileManager := new(mocks.FileManager)
	mockFileManager.On("Write", expectedPath, testContent, fs.FileMode(0644)).Return(nil)
	mockPathChecker := new(mocks.PathChecker)
	mockPathChecker.On("IsPathExists", testContent).Return(false, errors.New("path does not exist"))
	xcconfigWriter := NewWriter(mockPathProvider, mockFileManager, mockPathChecker)

	// When
	path, err := xcconfigWriter.Write(testContent)

	// Then
	assert.Error(t, err)
	assert.Equal(t, path, "")
}

func Test_XCConfigInput_CorrectInputPathReturnSamePath(t *testing.T) {
	// Given
	input := "TEST.xcconfig"
	mockPathChecker := new(mocks.PathChecker)
	mockPathChecker.On("IsPathExists", input).Return(true, nil)
	mockPathProvider := new(mocks.PathProvider)
	mockFileManager := new(mocks.FileManager)
	xcconfigWriter := NewWriter(mockPathProvider, mockFileManager, mockPathChecker)

	// When
	path, err := xcconfigWriter.Write(input)

	// Then
	if assert.NoError(t, err) {
		assert.Equal(t, path, input)
	}
}
