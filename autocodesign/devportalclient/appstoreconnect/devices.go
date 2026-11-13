package appstoreconnect

import (
	"net/http"
)

const DevicesEndpoint = "devices"

type ListDevicesSortOption string

const (
	ListDevicesSortOptionName         ListDevicesSortOption = "name"
	ListDevicesSortOptionNameDesc     ListDevicesSortOption = "-name"
	ListDevicesSortOptionPlatform     ListDevicesSortOption = "platform"
	ListDevicesSortOptionPlatformDesc ListDevicesSortOption = "-platform"
	ListDevicesSortOptionUDID         ListDevicesSortOption = "udid"
	ListDevicesSortOptionUDIDDesc     ListDevicesSortOption = "-udid"
	ListDevicesSortOptionStatus       ListDevicesSortOption = "status"
	ListDevicesSortOptionStatusDesc   ListDevicesSortOption = "-status"
	ListDevicesSortOptionID           ListDevicesSortOption = "id"
	ListDevicesSortOptionIDDesc       ListDevicesSortOption = "-id"
)

type ListDevicesOptions struct {
	PagingOptions
	FilterUDID     string                `url:"filter[udid],omitempty"`
	FilterPlatform DevicePlatform        `url:"filter[platform],omitempty"`
	FilterStatus   Status                `url:"filter[status],omitempty"`
	Sort           ListDevicesSortOption `url:"sort,omitempty"`
}

type DeviceClass string

const (
	AppleWatch DeviceClass = "APPLE_WATCH"
	Ipad       DeviceClass = "IPAD"
	Iphone     DeviceClass = "IPHONE"
	Ipod       DeviceClass = "IPOD"
	AppleTV    DeviceClass = "APPLE_TV"
	Mac        DeviceClass = "MAC"
)

type DevicePlatform string

const (
	IOSDevice   DevicePlatform = "IOS"
	MacOSDevice DevicePlatform = "MAC_OS"
)

type Status string

const (
	Enabled  Status = "ENABLED"
	Disabled Status = "DISABLED"
)

type DeviceAttributes struct {
	DeviceClass DeviceClass      `json:"deviceClass"`
	Model       string           `json:"model"`
	Name        string           `json:"name"`
	Platform    BundleIDPlatform `json:"platform"`
	Status      Status           `json:"status"`
	UDID        string           `json:"udid"`
	AddedDate   string           `json:"addedDate"`
}

type Device struct {
	Type       string           `json:"type"`
	ID         string           `json:"id"`
	Attributes DeviceAttributes `json:"attributes"`
}

type DevicesResponse struct {
	Data  []Device           `json:"data"`
	Links PagedDocumentLinks `json:"links,omitempty"`
	Meta  PagingInformation  `json:"meta,omitempty"`
}

type DeviceResponse struct {
	Data  Device             `json:"data"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

func (s ProvisioningService) ListDevices(opt *ListDevicesOptions) (*DevicesResponse, error) {
	if err := opt.UpdateCursor(); err != nil {
		return nil, err
	}

	u, err := addOptions(DevicesEndpoint, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &DevicesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

type DeviceCreateRequestDataAttributes struct {
	Name     string           `json:"name"`
	Platform BundleIDPlatform `json:"platform"`
	UDID     string           `json:"udid"`
}

type DeviceCreateRequestData struct {
	Attributes DeviceCreateRequestDataAttributes `json:"attributes"`
	Type       string                            `json:"type"`
}

type DeviceCreateRequest struct {
	Data DeviceCreateRequestData `json:"data"`
}

func (s ProvisioningService) RegisterNewDevice(body DeviceCreateRequest) (*DeviceResponse, error) {
	req, err := s.client.NewRequest(http.MethodPost, DevicesEndpoint, body)
	if err != nil {
		return nil, err
	}

	r := &DeviceResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s ProvisioningService) Devices(relationshipLink string, opt *PagingOptions) (*DevicesResponse, error) {
	if err := opt.UpdateCursor(); err != nil {
		return nil, err
	}

	u, err := addOptions(relationshipLink, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequestWithRelationshipURL(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &DevicesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}
