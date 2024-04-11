package ziputil

import "io"

// ReadCloser ...
type ReadCloser interface {
	ReadFile(relPthPattern string) (File, error)
	Close() error
}

// File ...
type File interface {
	Name() string
	Open() (io.ReadCloser, error)
}
