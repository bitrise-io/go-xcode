package artifacts

import (
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/v2/zip"
)

// XCArchiveReader ...
type XCArchiveReader struct {
	zipReader zip.ReadCloser
}

// NewXCArchiveReader ...
func NewXCArchiveReader(reader zip.ReadCloser) XCArchiveReader {
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
