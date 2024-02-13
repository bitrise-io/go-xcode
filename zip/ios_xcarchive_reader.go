package zip

import "github.com/bitrise-io/go-xcode/plistutil"

type IOSXcarchiveReader struct {
	zipReader Reader
}

func NewIOSXcarchiveReader(reader Reader) IOSXcarchiveReader {
	return IOSXcarchiveReader{zipReader: reader}
}

func (reader IOSXcarchiveReader) AppInfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}
