package zip

// ReadCloser ...
type ReadCloser interface {
	ReadFile(relPthPattern string) ([]byte, error)
	Close() error
}
