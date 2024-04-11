package zip

import (
	"io"

	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/v2/ziputil"
)

// IOSXCArchiveReader ...
type IOSXCArchiveReader struct {
	zipReader ziputil.ReadCloser
}

// NewIOSXCArchiveReader ...
func NewIOSXCArchiveReader(reader ziputil.ReadCloser) IOSXCArchiveReader {
	return IOSXCArchiveReader{zipReader: reader}
}

// AppInfoPlist ...
func (reader IOSXCArchiveReader) AppInfoPlist() (plistutil.PlistData, error) {
	f, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Info.plist")
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
