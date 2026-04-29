package artifacts

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// IPAReader ...
type IPAReader struct {
	zipReader ZipReadCloser
}

// NewIPAReader ...
func NewIPAReader(zipReader ZipReadCloser) IPAReader {
	return IPAReader{zipReader: zipReader}
}

// ProvisioningProfileInfo ...
func (reader IPAReader) ProvisioningProfileInfo() (*profileutil.ProvisioningProfileInfoModel, error) {
	b, err := reader.zipReader.ReadFile("Payload/*.app/embedded.mobileprovision")
	if err != nil {
		return nil, err
	}

	provisioningProfileInfo, err := profileutil.NewProvisioningProfileInfoFromPKCS7Content(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile info: %w", err)
	}

	return &provisioningProfileInfo, nil
}

// AppInfoPlist ...
func (reader IPAReader) AppInfoPlist() (plistutil.PlistData, error) {
	b, err := reader.zipReader.ReadFile("Payload/*.app/Info.plist")
	if err != nil {
		return nil, err
	}

	return plistutil.NewPlistDataFromContent(string(b))
}
