package artifacts

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/plistutil"
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

	profilePKCS7, err := profileutil.ProvisioningProfileFromContent(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded.mobilprovision: %w", err)
	}

	provisioningProfileInfo, err := profileutil.NewProvisioningProfileInfo(*profilePKCS7)
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
