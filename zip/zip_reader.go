package zip

import (
	"fmt"
	"io"
	"sort"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/internals/zip"
	"github.com/ryanuber/go-glob"
)

// Reader ...
type Reader struct {
	zipReader zip.ReadCloser
	logger    log.Logger
}

// NewReader ...
func NewReader(archivePath string, logger log.Logger) (*Reader, error) {
	zipReader, err := zip.NewDefaultReadCloser(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive %s: %w", archivePath, err)
	}

	return &Reader{
		zipReader: zipReader,
		logger:    logger,
	}, nil
}

// Close ...
func (reader Reader) Close() error {
	return reader.zipReader.Close()
}

// ReadFile ...
func (reader Reader) ReadFile(targetPathGlob string) ([]byte, error) {
	var files []zip.File
	for _, file := range reader.zipReader.Files() {
		name := file.Name()
		if glob.Glob(targetPathGlob, name) {
			files = append(files, file)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return len(files[i].Name()) < len(files[j].Name())
	})

	if len(files) == 0 {
		return nil, fmt.Errorf("no file found with pattern: %s", targetPathGlob)
	}

	file := files[0]
	r, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open archive file %s: %w", file.Name(), err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			reader.logger.Warnf("failed to close archive file %s: %s", file.Name(), err)
		}
	}()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive file %s: %w", file.Name(), err)
	}

	return b, nil
}

// IsFileOrDirExistsInZipArchive ...
func (reader Reader) IsFileOrDirExistsInZipArchive(targetPathGlob string) bool {
	for _, file := range reader.zipReader.Files() {
		if glob.Glob(targetPathGlob, file.Name()) {
			return true
		}
	}
	return false
}
