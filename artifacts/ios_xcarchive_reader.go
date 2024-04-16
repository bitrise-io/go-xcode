package artifacts

import (
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/v2/zip"
)

// IOSXCArchiveReader ...
type IOSXCArchiveReader struct {
	zipReader zip.ReadCloser
}

// NewIOSXCArchiveReader ...
func NewIOSXCArchiveReader(reader zip.ReadCloser) IOSXCArchiveReader {
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
