package destination

import (
	"fmt"
	"strings"
)

const (
	genericPlatformKey = "generic/platform"
	platformKey        = "platform"
	nameKey            = "name"
	osKey              = "OS"
	archKey            = "arch"
)

type Platform string

const (
	MacOS             Platform = "macOS"
	IOS               Platform = "iOS"
	IOSSimulator      Platform = "iOS Simulator"
	WatchOS           Platform = "watchOS"
	WatchOSSimulator  Platform = "watchOS Simulator"
	TvOS              Platform = "tvOS"
	TvOSSimulator     Platform = "tvOS Simulator"
	DriverKit         Platform = "DriverKit"
	VisionOSSimulator Platform = "visionOS Simulator"
)

// Specifier ...
type Specifier map[string]string

// NewSpecifier ...
func NewSpecifier(destination string) (Specifier, error) {
	specifier := Specifier{}

	parts := strings.Split(destination, ",")
	for _, part := range parts {
		keyAndValue := strings.Split(part, "=")

		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf(`could not parse "%s" because it is not a valid key=value pair in destination: %s`, part, destination)
		}

		key := keyAndValue[0]
		value := keyAndValue[1]

		specifier[key] = value
	}

	return specifier, nil
}

// Platform ...
func (s Specifier) Platform() (Platform, bool) {
	p, ok := s[genericPlatformKey]
	if ok {
		return Platform(p), true
	}

	return Platform(s[platformKey]), false
}

// Name ...
func (s Specifier) Name() string {
	return s[nameKey]
}

// OS ...
func (s Specifier) OS() string {
	return s[osKey]
}

// Arch ...
func (s Specifier) Arch() string {
	return s[archKey]
}
