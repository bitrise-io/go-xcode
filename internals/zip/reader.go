package zip

import (
	"archive/zip"
	"fmt"
)

type ReadCloser interface {
	Files() []File
	Close() error
}

type defaultReadCloser struct {
	zipReader *zip.ReadCloser
}

func NewDefaultReadCloser(archivePath string) (ReadCloser, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive %s: %w", archivePath, err)
	}

	return defaultReadCloser{
		zipReader: zipReader,
	}, nil
}

func (readCloser defaultReadCloser) Files() []File {
	var files []File
	for _, file := range readCloser.zipReader.File {
		files = append(files, newDefaultFile(file))
	}
	return files
}

func (readCloser defaultReadCloser) Close() error {
	return readCloser.zipReader.Close()
}
