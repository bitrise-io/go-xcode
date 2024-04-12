package ziputil

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/ryanuber/go-glob"
)

type defaultRead struct {
	zipReader *zip.ReadCloser
	logger    log.Logger
}

// NewDefaultRead ...
func NewDefaultRead(archivePath string, logger log.Logger) (ReadCloser, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive %s: %w", archivePath, err)
	}

	return defaultRead{
		zipReader: zipReader,
		logger:    logger,
	}, nil
}

// ReadFile ...
func (r defaultRead) ReadFile(relPthPattern string) ([]byte, error) {
	absPthPattern := filepath.Join("*", relPthPattern)

	var file *zip.File
	for _, f := range r.zipReader.File {
		if glob.Glob(absPthPattern, f.Name) {
			file = f
			break
		}
	}

	if file == nil {
		return nil, fmt.Errorf("no file found with pattern: %s", absPthPattern)
	}

	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", file.Name, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			r.logger.Warnf("Failed to close %s: %s", file.Name, err)
		}
	}()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", file.Name, err)
	}

	return b, nil
}

// Close ...
func (r defaultRead) Close() error {
	return r.zipReader.Close()
}
