package devportalservice

import (
	"io"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/v2/fileutil"
)

type MockFileReader struct {
	contents string
}

func NewMockFileReader(contents string) fileutil.FileManager {
	return &MockFileReader{
		contents: contents,
	}
}

func (r *MockFileReader) Open(path string) (*os.File, error) {
	panic("not implemented")
}

func (r *MockFileReader) OpenReaderIfExists(path string) (io.Reader, error) {
	return io.NopCloser(strings.NewReader(r.contents)), nil
}

func (r *MockFileReader) ReadDirEntryNames(path string) ([]string, error) {
	panic("not implemented")
}

func (r *MockFileReader) Remove(path string) error {
	panic("not implemented")
}

func (r *MockFileReader) RemoveAll(path string) error {
	panic("not implemented")
}

func (r *MockFileReader) Write(path string, value string, perm os.FileMode) error {
	panic("not implemented")
}

func (r *MockFileReader) WriteBytes(path string, value []byte) error {
	panic("not implemented")
}

func (r *MockFileReader) FileSizeInBytes(pth string) (int64, error) {
	panic("not implemented")
}
