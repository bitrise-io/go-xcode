package appstoreconnectclient

import (
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
)

// Client ...
type Client struct {
	*CertificateSource
	*DeviceClient
	*ProfileClient
}

// NewAPIDevportalClient ...
func NewAPIDevportalClient(client *appstoreconnect.Client) autocodesign.DevPortalClient {
	return Client{
		CertificateSource: NewCertificateSource(client),
		DeviceClient:      NewDeviceClient(client),
		ProfileClient:     NewProfileClient(client),
	}
}
