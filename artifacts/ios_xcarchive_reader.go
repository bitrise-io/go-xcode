package artifacts

import (
	"github.com/bitrise-io/go-xcode/v2/plistutil"
)

type IOSXCArchiveReader struct {
	zipReader ZipReadCloser
}

func NewIOSXCArchiveReader(reader ZipReadCloser) IOSXCArchiveReader {
	return IOSXCArchiveReader{zipReader: reader}
}

func (reader IOSXCArchiveReader) AppInfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}
