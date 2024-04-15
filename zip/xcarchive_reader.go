package zip

import (
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
	b, err := reader.zipReader.ReadFile("*.xcarchive/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}

// IsMacOS ...
func (reader XCArchiveReader) IsMacOS() bool {
	// TODO: distingush unzip and not found errors
	_, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Contents/*")
	return err == nil
}
