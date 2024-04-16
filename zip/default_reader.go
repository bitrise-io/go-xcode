package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"sort"

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
	var files []*zip.File
	for _, f := range r.zipReader.File {
		if glob.Glob(relPthPattern, f.Name) {
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no file found with pattern: %s", relPthPattern)
	}

	sort.Slice(files, func(i, j int) bool {
		return len(files[i].Name) < len(files[j].Name)
	})

	file := files[0]
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

func IsErrFormat(err error) bool {
	return errors.Is(err, zip.ErrFormat)
}
