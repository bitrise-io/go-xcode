package zip

import "github.com/bitrise-io/go-xcode/plistutil"

// XcarchiveReader ...
type XcarchiveReader struct {
	zipReader Reader
}

// NewXcarchiveReader ...
func NewXcarchiveReader(reader Reader) XcarchiveReader {
	return XcarchiveReader{zipReader: reader}
}

// InfoPlist ...
func (reader XcarchiveReader) InfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}

// IsMacOS ...
func (reader XcarchiveReader) IsMacOS() bool {
	return reader.zipReader.IsFileOrDirExistsInZipArchive("*.xcarchive/Products/Applications/*.app/Contents")
}
