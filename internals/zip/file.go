package zip

import (
	"archive/zip"
	"io"
)

type File interface {
	Name() string
	Open() (io.ReadCloser, error)
}

type defaultFile struct {
	file *zip.File
}

func newDefaultFile(file *zip.File) File {
	return defaultFile{file: file}
}

func (file defaultFile) Name() string {
	return file.file.Name
}

func (file defaultFile) Open() (io.ReadCloser, error) {
	return file.file.Open()
}
