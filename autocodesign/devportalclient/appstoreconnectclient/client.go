// Package appstoreconnectclient implements autocodesign.DevPortalClient, using an API key as the authentication method.
//
// It depends on appstoreconnect package.
package appstoreconnectclient

import (
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
)

type Client struct {
	*AuthClient
	*CertificateSource
	*DeviceClient
	*ProfileClient
}

func NewAPIDevPortalClient(client *appstoreconnect.Client) autocodesign.DevPortalClient {
	return Client{
		AuthClient:        NewAuthClient(),
		CertificateSource: NewCertificateSource(client),
		DeviceClient:      NewDeviceClient(client),
		ProfileClient:     NewProfileClient(client),
	}
}
