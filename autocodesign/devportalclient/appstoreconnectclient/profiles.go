package appstoreconnectclient

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
)

// APIProfile ...
type APIProfile struct {
	profile *appstoreconnect.Profile
	client  *appstoreconnect.Client
}

// NewAPIProfile ...
func NewAPIProfile(client *appstoreconnect.Client, profile *appstoreconnect.Profile) autocodesign.Profile {
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
func (p APIProfile) CertificateIDs() ([]string, error) {
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

	ids := []string{}
	for _, cert := range certificates {
		ids = append(ids, cert.ID)
	}

	return ids, nil
}

// DeviceIDs ...
func (p APIProfile) DeviceIDs() ([]string, error) {
	var nextPageURL string
	var ids []string
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

		for _, device := range response.Data {
			ids = append(ids, device.ID)
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
		return appstoreconnect.BundleID{}, wrapInProfileError(err)
	}

	return bundleIDresp.Data, nil
}

// Entitlements ...
func (p APIProfile) Entitlements() (autocodesign.Entitlements, error) {
	return autocodesign.ParseRawProfileEntitlements(p.profile.Attributes.ProfileContent)
}

// ProfileClient ...
type ProfileClient struct {
	client *appstoreconnect.Client
}

// NewProfileClient ...
func NewProfileClient(client *appstoreconnect.Client) *ProfileClient {
	return &ProfileClient{client: client}
}

// FindProfile ...
func (c *ProfileClient) FindProfile(name string, profileType appstoreconnect.ProfileType) (autocodesign.Profile, error) {
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
func (c *ProfileClient) DeleteProfile(id string) error {
	if err := c.client.Provisioning.DeleteProfile(id); err != nil {
		var respErr *appstoreconnect.ErrorResponse
		if ok := errors.As(err, &respErr); ok {
			if respErr.Response != nil && respErr.Response.StatusCode == http.StatusNotFound {
				return nil
			}
		}

		return err
	}

	return nil
}

// CreateProfile ...
func (c *ProfileClient) CreateProfile(name string, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (autocodesign.Profile, error) {
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

func (c *ProfileClient) deleteExpiredProfile(bundleID *appstoreconnect.BundleID, profileName string) error {
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

func (c *ProfileClient) createProfile(name string, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (autocodesign.Profile, error) {
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
func (c *ProfileClient) FindBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
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
func (c *ProfileClient) CreateBundleID(bundleIDIdentifier, appIDName string) (*appstoreconnect.BundleID, error) {
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
func (c *ProfileClient) CheckBundleIDEntitlements(bundleID appstoreconnect.BundleID, appEntitlements autocodesign.Entitlements) error {
	response, err := c.client.Provisioning.Capabilities(bundleID.Relationships.Capabilities.Links.Related)
	if err != nil {
		return err
	}

	return checkBundleIDEntitlements(response.Data, appEntitlements)
}

// SyncBundleID ...
func (c *ProfileClient) SyncBundleID(bundleID appstoreconnect.BundleID, appEntitlements autocodesign.Entitlements) error {
	for key, value := range appEntitlements {
		ent := autocodesign.Entitlement{key: value}
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
	var respErr *appstoreconnect.ErrorResponse
	if ok := errors.As(err, &respErr); ok {
		if respErr.Response != nil && respErr.Response.StatusCode == http.StatusNotFound {
			return autocodesign.NewProfilesInconsistentError(err)
		}
	}

	return err
}

func checkBundleIDEntitlements(bundleIDEntitlements []appstoreconnect.BundleIDCapability, appEntitlements autocodesign.Entitlements) error {
	for k, v := range appEntitlements {
		ent := autocodesign.Entitlement{k: v}

		if !ent.AppearsOnDeveloperPortal() {
			continue
		}

		found := false
		for _, cap := range bundleIDEntitlements {
			equal, err := ent.Equal(cap, appEntitlements)
			if err != nil {
				return err
			}

			if equal {
				found = true
				break
			}
		}

		if !found {
			return autocodesign.NonmatchingProfileError{
				Reason: fmt.Sprintf("bundle ID missing Capability (%s) required by project Entitlement (%s)", appstoreconnect.ServiceTypeByKey[k], k),
			}
		}
	}

	return nil
}

func isMultipleProfileErr(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "multiple profiles found with the name")
}
