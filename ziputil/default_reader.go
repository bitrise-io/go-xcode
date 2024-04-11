package ziputil

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ryanuber/go-glob"
)

type defaultRead struct {
	zipReader *zip.ReadCloser
}

// NewDefaultRead ...
func NewDefaultRead(archivePath string) (ReadCloser, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive %s: %w", archivePath, err)
	}

	return defaultRead{
		zipReader: zipReader,
	}, nil
}

// ReadFile ...
func (readCloser defaultRead) ReadFile(relPthPattern string) (File, error) {
	for _, file := range readCloser.zipReader.File {
		absPthPattern := filepath.Join("*", relPthPattern)
		if glob.Glob(absPthPattern, file.Name) {
			return newZipFile(file), nil
		}
	}
	return nil, nil
}

// Close ...
func (readCloser defaultRead) Close() error {
	return readCloser.zipReader.Close()
}

type zipFile struct {
	file *zip.File
}

func newZipFile(file *zip.File) File {
	return zipFile{file: file}
}

// Name ...
func (file zipFile) Name() string {
	return file.file.Name
}

// Open ...
func (file zipFile) Open() (io.ReadCloser, error) {
	return file.file.Open()
}
