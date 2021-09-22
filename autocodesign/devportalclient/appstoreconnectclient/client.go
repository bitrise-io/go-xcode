package appstoreconnectclient

import (
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/devportal"
)

// NewAPIDevportalClient ...
func NewAPIDevportalClient(client *appstoreconnect.Client) devportal.Client {
	return devportal.Client{
		CertificateSource: NewAPICertificateSource(client),
		DeviceClient:      NewAPIDeviceClient(client),
		ProfileClient:     NewAPIProfileClient(client),
	}
}
