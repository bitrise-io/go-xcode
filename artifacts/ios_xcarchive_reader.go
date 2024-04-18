package artifacts

import (
	"github.com/bitrise-io/go-xcode/plistutil"
)

// IOSXCArchiveReader ...
type IOSXCArchiveReader struct {
	zipReader ZipReadCloser
}

// NewIOSXCArchiveReader ...
func NewIOSXCArchiveReader(reader ZipReadCloser) IOSXCArchiveReader {
	return IOSXCArchiveReader{zipReader: reader}
}

// AppInfoPlist ...
func (reader IOSXCArchiveReader) AppInfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}
