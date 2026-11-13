package artifacts

type ZipReadCloser interface {
	ReadFile(relPthPattern string) ([]byte, error)
	Close() error
}
