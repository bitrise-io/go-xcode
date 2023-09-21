package profile

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/fullsailor/pkcs7"
)

type Type string

const (
	Development Type = "Development"
	AdHoc       Type = "Ad hoc"
	AppStore    Type = "App Store"
	DeveloperID Type = "Developer ID"
	Enterprise  Type = "Enterprise"
)

type Platform string

const (
	IOS      Platform = "iOS"
	XROS     Platform = "xrOS"
	VisionOS Platform = "visionOS"
	TVOS     Platform = "tvOS"
	OSX      Platform = "OSX"
)

type Profile struct {
	AppIDName                   string                 `plist:"AppIDName"`
	ApplicationIdentifierPrefix []string               `plist:"ApplicationIdentifierPrefix"`
	CreationDate                time.Time              `plist:"CreationDate"`
	Platform                    []string               `plist:"Platform"`
	IsXcodeManaged              bool                   `plist:"IsXcodeManaged"`
	DeveloperCertificates       [][]byte               `plist:"DeveloperCertificates"`
	DEREncodedProfile           []byte                 `plist:"DER-Encoded-Profile"`
	Entitlements                map[string]interface{} `plist:"Entitlements"`
	ExpirationDate              time.Time              `plist:"ExpirationDate"`
	Name                        string                 `plist:"Name"`
	ProvisionedDevices          []string               `plist:"ProvisionedDevices,omitempty"`
	ProvisionsAllDevices        *bool                  `plist:"ProvisionsAllDevices,omitempty"`
	TeamIdentifier              []string               `plist:"TeamIdentifier"`
	TeamName                    string                 `plist:"TeamName"`
	TimeToLive                  int                    `plist:"TimeToLive"`
	UUID                        string                 `plist:"UUID"`
	Version                     int                    `plist:"Version"`
}

func NewProfile(reader io.Reader) (*Profile, error) {
	profileMessage, err := parseProfile(reader)
	if err != nil {
		return nil, err
	}

	return newProfileFromPlist(profileMessage.Content)
}

func parseProfile(reader io.Reader) (*pkcs7.PKCS7, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return pkcs7.Parse(content)
}

func newProfileFromPlist(data []byte) (*Profile, error) {
	var profile Profile
	_, err := plist.Unmarshal(data, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (p Profile) Type() Type {
	/*
		| macOS        | ProvisionedDevices | ProvisionsAllDevices |
		|--------------|--------------------|----------------------|
		| Development  | true               | false                |
		| App Store    | false              | false                |
		| Developer ID | false              | true                 |

		---

		| iOS         | ProvisionedDevices | ProvisionsAllDevices | get-task-allow |
		|-------------|--------------------|----------------------|----------------|
		| Development | true               | false                | true           |
		| Ad Hoc      | true               | false                | false          |
		| App Store   | false              | false                | false          |
		| Enterprise  | false              | true                 | false          |
	*/

	hasProvisionedDevices := len(p.ProvisionedDevices) > 0
	provisionsAllDevices := p.ProvisionsAllDevices != nil && *p.ProvisionsAllDevices
	isMacOS := sliceutil.IsStringInSlice(string(OSX), p.Platform)
	isIOS := sliceutil.IsStringInSlice(string(IOS), p.Platform)

	if isMacOS && !isIOS {
		switch {
		case hasProvisionedDevices && !provisionsAllDevices:
			return Development
		case !hasProvisionedDevices && provisionsAllDevices:
			return DeveloperID
		case !hasProvisionedDevices && !provisionsAllDevices:
			return AppStore
		default:
			// TODO: this shouldn't happen
			return Development
		}
	}

	var getTaskAllow bool
	if err := p.ReadEntitlement("get-task-allow", &getTaskAllow); err != nil {
		// TODO: this shouldn't happen
		getTaskAllow = false
	}

	switch {
	case hasProvisionedDevices && !provisionsAllDevices && getTaskAllow:
		return Development
	case hasProvisionedDevices && !provisionsAllDevices && !getTaskAllow:
		return AdHoc
	case !hasProvisionedDevices && provisionsAllDevices && !getTaskAllow:
		return Enterprise
	case !hasProvisionedDevices && !provisionsAllDevices && !getTaskAllow:
		return AppStore
	default:
		// TODO: this shouldn't happen
		return Development
	}
}

func (p Profile) ReadEntitlement(key string, value interface{}) error {
	kind, err := validateType(value)
	if err != nil {
		return err
	}

	v, ok := p.Entitlements[key]
	if !ok {
		return fmt.Errorf("key (%s) not found", key)
	}

	vKind := reflect.ValueOf(v).Kind()
	if vKind != *kind {
		return fmt.Errorf("entitlement value (%s) is not %s", vKind, *kind)
	}

	reflect.ValueOf(value).Elem().Set(reflect.ValueOf(v))

	return nil
}

func validateType(target interface{}) (*reflect.Kind, error) {
	c := reflect.ValueOf(target)
	if c.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("value must be a pointer")
	}
	c = c.Elem()
	switch c.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Float64, reflect.Slice, reflect.Map:
		k := c.Kind()
		return &k, nil
	default:
		return nil, fmt.Errorf("value unsupported type (%s pointer), supported types are string, bool, float64, slice and map", c.Kind())
	}
}
