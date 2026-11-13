package appstoreconnect

import (
	"net/http"

	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/time"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

const ProfilesEndpoint = "profiles"

type ListProfilesOptions struct {
	PagingOptions
	FilterProfileState ProfileState `url:"filter[profileState],omitempty"`
	FilterProfileType  ProfileType  `url:"filter[profileType],omitempty"`
	FilterName         string       `url:"filter[name],omitempty"`
	Include            string       `url:"include,omitempty"`
}

type BundleIDPlatform string

const (
	IOS       BundleIDPlatform = "IOS"
	MacOS     BundleIDPlatform = "MAC_OS"
	Universal BundleIDPlatform = "UNIVERSAL"
)

type ProfileState string

const (
	Active  ProfileState = "ACTIVE"
	Invalid ProfileState = "INVALID"
)

type ProfileType string

const (
	IOSAppDevelopment ProfileType = "IOS_APP_DEVELOPMENT"
	IOSAppStore       ProfileType = "IOS_APP_STORE"
	IOSAppAdHoc       ProfileType = "IOS_APP_ADHOC"
	IOSAppInHouse     ProfileType = "IOS_APP_INHOUSE"

	MacAppDevelopment ProfileType = "MAC_APP_DEVELOPMENT"
	MacAppStore       ProfileType = "MAC_APP_STORE"
	MacAppDirect      ProfileType = "MAC_APP_DIRECT"

	TvOSAppDevelopment ProfileType = "TVOS_APP_DEVELOPMENT"
	TvOSAppStore       ProfileType = "TVOS_APP_STORE"
	TvOSAppAdHoc       ProfileType = "TVOS_APP_ADHOC"
	TvOSAppInHouse     ProfileType = "TVOS_APP_INHOUSE"
)

// ReadableString returns the readable version of the ProjectType
// e.g: IOSAppDevelopment => development
func (t ProfileType) ReadableString() string {
	switch t {
	case IOSAppStore, MacAppStore, TvOSAppStore:
		return "app store"
	case IOSAppInHouse, TvOSAppInHouse:
		return "enterprise"
	case IOSAppAdHoc, TvOSAppAdHoc:
		return "ad-hoc"
	case IOSAppDevelopment, MacAppDevelopment, TvOSAppDevelopment:
		return "development"
	case MacAppDirect:
		return "development ID"
	}
	return ""
}

type ProfileAttributes struct {
	Name           string           `json:"name"`
	Platform       BundleIDPlatform `json:"platform"`
	ProfileContent []byte           `json:"profileContent"`
	UUID           string           `json:"uuid"`
	CreatedDate    string           `json:"createdDate"`
	ProfileState   ProfileState     `json:"profileState"`
	ProfileType    ProfileType      `json:"profileType"`
	ExpirationDate time.Time        `json:"expirationDate"`
}

type Profile struct {
	Attributes ProfileAttributes `json:"attributes"`

	Relationships struct {
		BundleID struct {
			Links struct {
				Related string `json:"related"`
				Self    string `json:"self"`
			} `json:"links"`
		} `json:"bundleId"`

		Certificates struct {
			Links struct {
				Related string `json:"related"`
				Self    string `json:"self"`
			} `json:"links"`
		} `json:"certificates"`

		Devices struct {
			Links struct {
				Related string `json:"related"`
				Self    string `json:"self"`
			} `json:"links"`
		} `json:"devices"`
	} `json:"relationships"`

	ID string `json:"id"`
}

type ProfilesResponse struct {
	Data     []Profile `json:"data"`
	Included []struct {
		Type       string            `json:"type"`
		ID         string            `json:"id"`
		Attributes serialized.Object `json:"attributes"`
	} `json:"included"`
	Links PagedDocumentLinks `json:"links,omitempty"`
	Meta  PagingInformation  `json:"meta,omitempty"`
}

func (s ProvisioningService) ListProfiles(opt *ListProfilesOptions) (*ProfilesResponse, error) {
	if err := opt.UpdateCursor(); err != nil {
		return nil, err
	}

	u, err := addOptions(ProfilesEndpoint, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &ProfilesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

type ProfileCreateRequestDataAttributes struct {
	Name        string      `json:"name"`
	ProfileType ProfileType `json:"profileType"`
}

type ProfileCreateRequestDataRelationshipData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ProfileCreateRequestDataRelationshipsBundleID struct {
	Data ProfileCreateRequestDataRelationshipData `json:"data"`
}

type ProfileCreateRequestDataRelationshipsCertificates struct {
	Data []ProfileCreateRequestDataRelationshipData `json:"data"`
}

type ProfileCreateRequestDataRelationshipsDevices struct {
	Data []ProfileCreateRequestDataRelationshipData `json:"data"`
}

type ProfileCreateRequestDataRelationships struct {
	BundleID     ProfileCreateRequestDataRelationshipsBundleID     `json:"bundleId"`
	Certificates ProfileCreateRequestDataRelationshipsCertificates `json:"certificates"`
	Devices      *ProfileCreateRequestDataRelationshipsDevices     `json:"devices,omitempty"`
}

type ProfileCreateRequestData struct {
	Attributes    ProfileCreateRequestDataAttributes    `json:"attributes"`
	Relationships ProfileCreateRequestDataRelationships `json:"relationships"`
	Type          string                                `json:"type"`
}

type ProfileCreateRequest struct {
	Data ProfileCreateRequestData `json:"data"`
}

func NewProfileCreateRequest(profileType ProfileType, name, bundleIDID string, certificateIDs []string, deviceIDs []string) ProfileCreateRequest {
	bundleIDData := ProfileCreateRequestDataRelationshipData{
		ID:   bundleIDID,
		Type: "bundleIds",
	}

	relationships := ProfileCreateRequestDataRelationships{
		BundleID: ProfileCreateRequestDataRelationshipsBundleID{Data: bundleIDData},
	}

	var certData []ProfileCreateRequestDataRelationshipData
	for _, id := range certificateIDs {
		certData = append(certData, ProfileCreateRequestDataRelationshipData{
			ID:   id,
			Type: "certificates",
		})
	}
	relationships.Certificates = ProfileCreateRequestDataRelationshipsCertificates{Data: certData}

	var deviceData []ProfileCreateRequestDataRelationshipData
	for _, id := range deviceIDs {
		deviceData = append(deviceData, ProfileCreateRequestDataRelationshipData{
			ID:   id,
			Type: "devices",
		})
	}
	if len(deviceData) > 0 {
		relationships.Devices = &ProfileCreateRequestDataRelationshipsDevices{Data: deviceData}
	}

	data := ProfileCreateRequestData{
		Attributes: ProfileCreateRequestDataAttributes{
			Name:        name,
			ProfileType: profileType,
		},
		Relationships: relationships,
		Type:          "profiles",
	}

	return ProfileCreateRequest{Data: data}
}

type ProfileResponse struct {
	Data  Profile            `json:"data"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

func (s ProvisioningService) CreateProfile(body ProfileCreateRequest) (*ProfileResponse, error) {
	req, err := s.client.NewRequest(http.MethodPost, ProfilesEndpoint, body)
	if err != nil {
		return nil, err
	}

	r := &ProfileResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s ProvisioningService) DeleteProfile(id string) error {
	req, err := s.client.NewRequest(http.MethodDelete, ProfilesEndpoint+"/"+id, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

func (s ProvisioningService) Profiles(relationshipLink string, opt *PagingOptions) (*ProfilesResponse, error) {
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

	r := &ProfilesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}
