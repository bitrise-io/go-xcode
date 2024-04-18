package artifacts

// ZipReadCloser ...
type ZipReadCloser interface {
	ReadFile(relPthPattern string) ([]byte, error)
	Close() error
}
