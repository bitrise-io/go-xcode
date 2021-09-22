package appstoreconnectclient

import (
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
)

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
