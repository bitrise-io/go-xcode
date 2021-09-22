package appstoreconnectclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/devportal"
)

// APIProfile ...
type APIProfile struct {
	profile *appstoreconnect.Profile
	client  *appstoreconnect.Client
}

// NewAPIProfile ...
func NewAPIProfile(client *appstoreconnect.Client, profile *appstoreconnect.Profile) devportal.Profile {
	return &APIProfile{
		profile: profile,
		client:  client,
	}
}

// ID ...
func (p APIProfile) ID() string {
	return p.profile.ID
}

// Attributes ...
func (p APIProfile) Attributes() appstoreconnect.ProfileAttributes {
	return p.profile.Attributes
}

// CertificateIDs ...
func (p APIProfile) CertificateIDs() (map[string]bool, error) {
	var nextPageURL string
	var certificates []appstoreconnect.Certificate
	for {
		response, err := p.client.Provisioning.Certificates(
			p.profile.Relationships.Certificates.Links.Related,
			&appstoreconnect.PagingOptions{
				Limit: 20,
				Next:  nextPageURL,
			},
		)
		if err != nil {
			return nil, wrapInProfileError(err)
		}

		certificates = append(certificates, response.Data...)

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			break
		}
	}

	ids := map[string]bool{}
	for _, cert := range certificates {
		ids[cert.ID] = true
	}

	return ids, nil
}

// DeviceIDs ...
func (p APIProfile) DeviceIDs() (map[string]bool, error) {
	var nextPageURL string
	ids := map[string]bool{}
	for {
		response, err := p.client.Provisioning.Devices(
			p.profile.Relationships.Devices.Links.Related,
			&appstoreconnect.PagingOptions{
				Limit: 20,
				Next:  nextPageURL,
			},
		)
		if err != nil {
			return nil, wrapInProfileError(err)
		}

		for _, dev := range response.Data {
			ids[dev.ID] = true
		}

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			break
		}
	}

	return ids, nil
}

// BundleID ...
func (p APIProfile) BundleID() (appstoreconnect.BundleID, error) {
	bundleIDresp, err := p.client.Provisioning.BundleID(p.profile.Relationships.BundleID.Links.Related)
	if err != nil {
		return appstoreconnect.BundleID{}, err
	}

	return bundleIDresp.Data, nil
}

// APIProfileClient ...
type APIProfileClient struct {
	client *appstoreconnect.Client
}

// NewAPIProfileClient ...
func NewAPIProfileClient(client *appstoreconnect.Client) devportal.ProfileClient {
	return &APIProfileClient{client: client}
}

// FindProfile ...
func (c *APIProfileClient) FindProfile(name string, profileType appstoreconnect.ProfileType) (devportal.Profile, error) {
	opt := &appstoreconnect.ListProfilesOptions{
		PagingOptions: appstoreconnect.PagingOptions{
			Limit: 1,
		},
		FilterProfileType: profileType,
		FilterName:        name,
	}

	r, err := c.client.Provisioning.ListProfiles(opt)
	if err != nil {
		return nil, err
	}
	if len(r.Data) == 0 {
		return nil, nil
	}

	return NewAPIProfile(c.client, &r.Data[0]), nil
}

// DeleteProfile ...
func (c *APIProfileClient) DeleteProfile(id string) error {
	if err := c.client.Provisioning.DeleteProfile(id); err != nil {
		if respErr, ok := err.(appstoreconnect.ErrorResponse); ok {
			if respErr.Response != nil && respErr.Response.StatusCode == http.StatusNotFound {
				return nil
			}
		}

		return err
	}

	return nil
}

// CreateProfile ...
func (c *APIProfileClient) CreateProfile(name string, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (devportal.Profile, error) {
	profile, err := c.createProfile(name, profileType, bundleID, certificateIDs, deviceIDs)
	if err != nil {
		// Expired profiles are not listed via profiles endpoint,
		// so we can not catch if the profile already exist but expired, before we attempt to create one with the managed profile name.
		// As a workaround we use the BundleID profiles relationship url to find and delete the expired profile.
		if isMultipleProfileErr(err) {
			log.Warnf("  Profile already exists, but expired, cleaning up...")
			if err := c.deleteExpiredProfile(&bundleID, name); err != nil {
				return nil, fmt.Errorf("expired profile cleanup failed: %s", err)
			}

			profile, err = c.createProfile(name, profileType, bundleID, certificateIDs, deviceIDs)
			if err != nil {
				return nil, err
			}

			return profile, nil
		}

		return nil, err
	}

	return profile, nil
}

func (c *APIProfileClient) deleteExpiredProfile(bundleID *appstoreconnect.BundleID, profileName string) error {
	var nextPageURL string
	var profile *appstoreconnect.Profile

	for {
		response, err := c.client.Provisioning.Profiles(bundleID.Relationships.Profiles.Links.Related, &appstoreconnect.PagingOptions{
			Limit: 20,
			Next:  nextPageURL,
		})
		if err != nil {
			return err
		}

		for _, d := range response.Data {
			if d.Attributes.Name == profileName {
				profile = &d
				break
			}
		}

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			break
		}
	}

	if profile == nil {
		return fmt.Errorf("failed to find profile: %s", profileName)
	}

	return c.DeleteProfile(profile.ID)
}

func (c *APIProfileClient) createProfile(name string, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (devportal.Profile, error) {
	// Create new Bitrise profile on App Store Connect
	r, err := c.client.Provisioning.CreateProfile(
		appstoreconnect.NewProfileCreateRequest(
			profileType,
			name,
			bundleID.ID,
			certificateIDs,
			deviceIDs,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provisioning profile for %s bundle ID: %s", profileType.ReadableString(), bundleID.Attributes.Identifier, err)
	}

	return NewAPIProfile(c.client, &r.Data), nil
}

// FindBundleID ...
func (c *APIProfileClient) FindBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
	var nextPageURL string
	var bundleIDs []appstoreconnect.BundleID
	for {
		response, err := c.client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{
			PagingOptions: appstoreconnect.PagingOptions{
				Limit: 20,
				Next:  nextPageURL,
			},
			FilterIdentifier: bundleIDIdentifier,
		})
		if err != nil {
			return nil, err
		}

		bundleIDs = append(bundleIDs, response.Data...)

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			break
		}
	}

	if len(bundleIDs) == 0 {
		return nil, nil
	}

	// The FilterIdentifier works as a Like command. It will not search for the exact match,
	// this is why we need to find the exact match in the list.
	for _, d := range bundleIDs {
		if d.Attributes.Identifier == bundleIDIdentifier {
			return &d, nil
		}
	}
	return nil, nil
}

// CreateBundleID ...
func (c *APIProfileClient) CreateBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
	appIDName := devportal.AppIDName(bundleIDIdentifier)

	r, err := c.client.Provisioning.CreateBundleID(
		appstoreconnect.BundleIDCreateRequest{
			Data: appstoreconnect.BundleIDCreateRequestData{
				Attributes: appstoreconnect.BundleIDCreateRequestDataAttributes{
					Identifier: bundleIDIdentifier,
					Name:       appIDName,
					Platform:   appstoreconnect.IOS,
				},
				Type: "bundleIds",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register AppID for bundleID (%s): %s", bundleIDIdentifier, err)
	}

	return &r.Data, nil
}

// CheckBundleIDEntitlements checks if a given Bundle ID has every capability enabled, required by the project.
func (c *APIProfileClient) CheckBundleIDEntitlements(bundleID appstoreconnect.BundleID, projectEntitlements devportal.Entitlement) error {
	response, err := c.client.Provisioning.Capabilities(bundleID.Relationships.Capabilities.Links.Related)
	if err != nil {
		return err
	}

	return checkBundleIDEntitlements(response.Data, projectEntitlements)
}

// SyncBundleID ...
func (c *APIProfileClient) SyncBundleID(bundleID appstoreconnect.BundleID, entitlements devportal.Entitlement) error {
	for key, value := range entitlements {
		ent := devportal.Entitlement{key: value}
		cap, err := ent.Capability()
		if err != nil {
			return err
		}
		if cap == nil {
			continue
		}

		body := appstoreconnect.BundleIDCapabilityCreateRequest{
			Data: appstoreconnect.BundleIDCapabilityCreateRequestData{
				Attributes: appstoreconnect.BundleIDCapabilityCreateRequestDataAttributes{
					CapabilityType: cap.Attributes.CapabilityType,
					Settings:       cap.Attributes.Settings,
				},
				Relationships: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationships{
					BundleID: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleID{
						Data: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData{
							ID:   bundleID.ID,
							Type: "bundleIds",
						},
					},
				},
				Type: "bundleIdCapabilities",
			},
		}
		_, err = c.client.Provisioning.EnableCapability(body)
		if err != nil {
			return err
		}
	}

	return nil
}

func wrapInProfileError(err error) error {
	if respErr, ok := err.(appstoreconnect.ErrorResponse); ok {
		if respErr.Response != nil && respErr.Response.StatusCode == http.StatusNotFound {
			return devportal.NonmatchingProfileError{
				Reason: fmt.Sprintf("profile was concurrently removed from Developer Portal: %v", err),
			}
		}
	}

	return err
}

func checkBundleIDEntitlements(bundleIDEntitlements []appstoreconnect.BundleIDCapability, projectEntitlements devportal.Entitlement) error {
	for k, v := range projectEntitlements {
		ent := devportal.Entitlement{k: v}

		if !ent.AppearsOnDeveloperPortal() {
			continue
		}

		found := false
		for _, cap := range bundleIDEntitlements {
			equal, err := ent.Equal(cap)
			if err != nil {
				return err
			}

			if equal {
				found = true
				break
			}
		}

		if !found {
			return devportal.NonmatchingProfileError{
				Reason: fmt.Sprintf("bundle ID missing Capability (%s) required by project Entitlement (%s)", appstoreconnect.ServiceTypeByKey[k], k),
			}
		}
	}

	return nil
}

func isMultipleProfileErr(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "multiple profiles found with the name")
}
