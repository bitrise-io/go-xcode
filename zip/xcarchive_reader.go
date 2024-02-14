package zip

import "github.com/bitrise-io/go-xcode/plistutil"

// XCArchiveReader ...
type XCArchiveReader struct {
	zipReader Reader
}

// NewXCArchiveReader ...
func NewXCArchiveReader(reader Reader) XCArchiveReader {
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
	return reader.zipReader.IsFileOrDirExistsInZipArchive("*.xcarchive/Products/Applications/*.app/Contents/*")
}
