package artifacts

import (
	"github.com/bitrise-io/go-xcode/v2/plistutil"
)

type XCArchiveReader struct {
	zipReader ZipReadCloser
}

func NewXCArchiveReader(reader ZipReadCloser) XCArchiveReader {
	return XCArchiveReader{zipReader: reader}
}

func (reader XCArchiveReader) InfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}

func (reader XCArchiveReader) IsMacOS() bool {
	_, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Contents/Info.plist")
	return err == nil
}
