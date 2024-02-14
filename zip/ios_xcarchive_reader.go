package zip

import "github.com/bitrise-io/go-xcode/plistutil"

// IOSXCArchiveReader ...
type IOSXCArchiveReader struct {
	zipReader Reader
}

// NewIOSXCArchiveReader ...
func NewIOSXCArchiveReader(reader Reader) IOSXCArchiveReader {
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
