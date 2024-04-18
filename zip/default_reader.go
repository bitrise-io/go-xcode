package zip

import (
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/internal/zip"
)

// DefaultReader is a zip reader, that utilises zip.StdlibRead and zip.DittoReader readers.
// If zip.StdlibRead.ReadFile fails it falls back to zip.DittoReader.ReadFile.
type DefaultReader struct {
	logger log.Logger

	stdlibZipReader      *zip.StdlibRead
	dittoZipReader       *zip.DittoReader
	useFallbackZipReader bool
}

// NewDefaultReader ...
func NewDefaultReader(archivePath string, logger log.Logger) (*DefaultReader, error) {
	stdlibReader, err := zip.NewStdlibRead(archivePath, logger)
	if err != nil {
		return nil, err
	}

	var dittoReader *zip.DittoReader
	if zip.IsDittoReaderAvailable() {
		dittoReader = zip.NewDittoReader(archivePath, logger)
	}

	return &DefaultReader{
		logger:               logger,
		stdlibZipReader:      stdlibReader,
		dittoZipReader:       dittoReader,
		useFallbackZipReader: false,
	}, nil
}

// ReadFile ...
func (r *DefaultReader) ReadFile(relPthPattern string) ([]byte, error) {
	if !r.useFallbackZipReader {
		b, err := r.stdlibZipReader.ReadFile(relPthPattern)
		if err != nil {
			if zip.IsErrFormat(err) {
				r.logger.Warnf("stdlib zip reader failed to read %s: %s", relPthPattern, err)
				r.logger.Warnf("Retrying with ditto zip reader...")

				r.useFallbackZipReader = true
				return r.ReadFile(relPthPattern)
			}
			return nil, err
		}

		return b, nil
	} else {
		return r.dittoZipReader.ReadFile(relPthPattern)
	}
}

// Close ...
func (r *DefaultReader) Close() error {
	stdlibZipReaderCloseErr := r.stdlibZipReader.Close()
	if !r.useFallbackZipReader {
		return stdlibZipReaderCloseErr
	}

	dittoZipReaderCloseErr := r.dittoZipReader.Close()
	if dittoZipReaderCloseErr == nil {
		return stdlibZipReaderCloseErr
	}

	// ditto reader's close has failed
	if stdlibZipReaderCloseErr != nil {
		r.logger.Warnf("failed to close stdlib zip reader: %s", stdlibZipReaderCloseErr)
	}

	return dittoZipReaderCloseErr
}
