package v2

import (
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/zipreader"
)

type IOSXcarchiveZipReader struct {
	zipReader zipreader.ZipReader
}

func NewIOSXcarchiveZipReader(reader zipreader.ZipReader) IOSXcarchiveZipReader {
	return IOSXcarchiveZipReader{zipReader: reader}
}

func (reader IOSXcarchiveZipReader) AppInfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Products/Applications/*.app/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}
