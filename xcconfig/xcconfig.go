package xcconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// Writer ...
type Writer interface {
	Write(input string) (string, error)
}

type writer struct {
	pathProvider pathutil.PathProvider
	fileManager  fileutil.FileManager
}

// NewWriter ...
func NewWriter(pathProvider pathutil.PathProvider, fileManager fileutil.FileManager) Writer {
	return &writer{pathProvider: pathProvider, fileManager: fileManager}
}

func (w writer) Write(content string) (string, error) {
	if w.isPath(content) {
		pathExists, err := pathutil.NewPathChecker().IsPathExists(content)
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
		if !pathExists {
			return "", fmt.Errorf("provided xcconfig file path doesn't exist")
		}
	}

	dir, err := w.pathProvider.CreateTempDir("")
	if err != nil {
		return "", fmt.Errorf("unable to create temp dir for writing XCConfig: %v", err)
	}
	xcconfigPath := filepath.Join(dir, "temp.xcconfig")
	if err = w.fileManager.Write(xcconfigPath, content, 0644); err != nil {
		return "", fmt.Errorf("unable to write XCConfig content into file: %v", err)
	}
	return xcconfigPath, nil
}

func (w writer) isPath(input string) bool {
	return strings.HasSuffix(input, ".xcconfig")
}
