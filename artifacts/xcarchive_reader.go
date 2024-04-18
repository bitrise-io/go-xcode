package artifacts

import (
	"github.com/bitrise-io/go-xcode/plistutil"
)

// XCArchiveReader ...
type XCArchiveReader struct {
	zipReader ZipReadCloser
}

// NewXCArchiveReader ...
func NewXCArchiveReader(reader ZipReadCloser) XCArchiveReader {
	return XCArchiveReader{zipReader: reader}
}

// InfoPlist ...
func (reader XCArchiveReader) InfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}

// IsMacOS ...
func (reader XCArchiveReader) IsMacOS() bool {
	_, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Contents/Info.plist")
	return err == nil
}
