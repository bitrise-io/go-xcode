package v2

import (
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/zipreader"
)

type XcarchiveZipReader struct {
	zipReader zipreader.ZipReader
}

func NewXcarchiveZipReader(reader zipreader.ZipReader) XcarchiveZipReader {
	return XcarchiveZipReader{zipReader: reader}
}

func (reader XcarchiveZipReader) InfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("*.xcarchive/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}

// IsMacOS determines if the xcarchive belongs to a macOS app, by searching for the 'Contents' dir in the '<app_name>.app' directory.
func (reader XcarchiveZipReader) IsMacOS() bool {
	return reader.zipReader.IsFileOrDirExistsInZipArchive("*.xcarchive/Products/Applications/*.app/Contents")
}
