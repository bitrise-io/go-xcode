package plistutil

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/v2/fileutil"
)

// newPlistFileMode is v1's mode for a newly created plist file.
const newPlistFileMode = 0644

// FileHandler reads and writes plist files, keeping their format (XML, binary, OpenStep or GNUstep)
// so a file can be written back as it was read.
type FileHandler interface {
	Read(path string) (PlistData, int, error)
	Write(path string, data PlistData, format int) error
}

type fileHandler struct {
	fileManager fileutil.FileManager
}

// NewFileHandler returns a FileHandler that accesses files through fileManager.
func NewFileHandler(fileManager fileutil.FileManager) FileHandler {
	return fileHandler{fileManager: fileManager}
}

// Read returns the plist at path and its format.
func (h fileHandler) Read(path string) (PlistData, int, error) {
	file, err := h.fileManager.Open(path)
	if err != nil {
		return nil, plist.InvalidFormat, err
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, plist.InvalidFormat, err
	}

	var data PlistData
	format, err := plist.Unmarshal(content, &data)
	if err != nil {
		return nil, plist.InvalidFormat, fmt.Errorf("failed to unmarshal %s: %w", path, err)
	}

	return data, format, nil
}

// Write writes data to path in the given format, such as the one Read returned. An existing file
// keeps its mode.
func (h fileHandler) Write(path string, data PlistData, format int) error {
	content, err := plist.Marshal(data, format)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", path, err)
	}

	return h.fileManager.Write(path, string(content), h.fileMode(path))
}

// fileMode returns the existing file's mode, so that Write keeps it as os.WriteFile did in v1;
// FileManager.Write would otherwise chmod the file. Opening follows symlinks, unlike Lstat.
func (h fileHandler) fileMode(path string) os.FileMode {
	file, err := h.fileManager.Open(path)
	if err != nil {
		return newPlistFileMode
	}
	defer func() {
		_ = file.Close()
	}()

	info, err := file.Stat()
	if err != nil {
		return newPlistFileMode
	}
	return info.Mode().Perm()
}
