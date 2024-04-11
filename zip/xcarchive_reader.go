package zip

import (
	"io"

	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/v2/ziputil"
)

// XCArchiveReader ...
type XCArchiveReader struct {
	zipReader ziputil.ReadCloser
}

// NewXCArchiveReader ...
func NewXCArchiveReader(reader ziputil.ReadCloser) XCArchiveReader {
	return XCArchiveReader{zipReader: reader}
}

// InfoPlist ...
func (reader XCArchiveReader) InfoPlist() (plistutil.PlistData, error) {
	f, err := reader.zipReader.ReadFile("*.xcarchive/Info.plist")
	if err != nil {
		return nil, err
	}

	r, err := f.Open()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}

// IsMacOS ...
func (reader XCArchiveReader) IsMacOS() bool {
	_, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Contents/*")
	return err == nil
}
