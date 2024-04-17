package zip

import (
	"fmt"

	"github.com/bitrise-io/go-utils/v2/log"
)

// defaultReader is a zip ReadCloser, that utilises multiple zip readers.
// If one fails it falls back to the next reader.
type defaultReader struct {
	logger           log.Logger
	zipReaders       []ReadCloser
	currentReaderIdx int
}

func NewDefaultReader(archivePath string, logger log.Logger) (ReadCloser, error) {
	var zipReaders []ReadCloser

	stdlibReader, err := NewStdlibRead(archivePath, logger)
	if err != nil {
		return nil, err
	}
	zipReaders = append(zipReaders, stdlibReader)

	if IsDittoReaderAvailable() {
		dittoReader := NewDittoReader(archivePath, logger)
		zipReaders = append(zipReaders, dittoReader)
	}

	return &defaultReader{
		logger:           logger,
		zipReaders:       zipReaders,
		currentReaderIdx: 0,
	}, nil
}

func (r *defaultReader) ReadFile(relPthPattern string) ([]byte, error) {
	zipReader := r.zipReaders[r.currentReaderIdx]
	b, err := zipReader.ReadFile(relPthPattern)
	if err != nil && r.currentReaderIdx < len(r.zipReaders)-1 {
		r.logger.Warnf("zip reader #%d failed to read %s: %s", r.currentReaderIdx+1, relPthPattern, err)
		r.logger.Warnf("Retrying with the next zip reader...")

		r.currentReaderIdx++
		return r.ReadFile(relPthPattern)
	}
	return b, err
}

func (r *defaultReader) Close() error {
	var closeErrs []map[int]error
	for i := 0; i <= r.currentReaderIdx; i++ {
		zipReader := r.zipReaders[r.currentReaderIdx]
		if err := zipReader.Close(); err != nil {
			closeErrs = append(closeErrs, map[int]error{
				i: err,
			})
		}
	}

	return handleCloseErrs(closeErrs, r.logger)
}

func handleCloseErrs(closeErrs []map[int]error, logger log.Logger) error {
	if len(closeErrs) == 0 {
		return nil
	}

	// The last error is returned, the rest printed.
	if len(closeErrs) > 1 {
		for i := 0; i < len(closeErrs)-1; i++ {
			for idx, err := range closeErrs[i] {
				logger.Warnf("failed to close zip reader #%d: %s", idx, err)
			}
		}
	}

	for idx, err := range closeErrs[len(closeErrs)-1] {
		return fmt.Errorf("failed to close zip reader #%d: %s", idx, err)
	}

	return nil
}
