package profile

import (
	"fmt"
	"reflect"
	"time"

	"github.com/bitrise-io/go-utils/sliceutil"
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

type Details struct {
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

func (d Details) Type() Type {
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

	hasProvisionedDevices := len(d.ProvisionedDevices) > 0
	provisionsAllDevices := d.ProvisionsAllDevices != nil && *d.ProvisionsAllDevices
	isMacOS := sliceutil.IsStringInSlice(string(OSX), d.Platform)
	isIOS := sliceutil.IsStringInSlice(string(IOS), d.Platform)

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
	if err := d.ReadEntitlement("get-task-allow", &getTaskAllow); err != nil {
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

func (d Details) ReadEntitlement(key string, value interface{}) error {
	kind, err := validateType(value)
	if err != nil {
		return err
	}

	v, ok := d.Entitlements[key]
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
