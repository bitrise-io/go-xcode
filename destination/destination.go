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
)

type Platform string

const (
	macOS            Platform = "macOS"
	iOS              Platform = "iOS"
	iOSSimulator     Platform = "iOS Simulator"
	watchOS          Platform = "watchOS"
	watchOSSimulator Platform = "watchOS Simulator"
	tvOS             Platform = "tvOS"
	tvOSSimulator    Platform = "tvOS Simulator"
	DriverKit        Platform = "DriverKit"
)

type Specifier map[string]string

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

func (s Specifier) Platform() (Platform, bool) {
	p, ok := s[genericPlatformKey]
	if ok {
		return Platform(p), true
	}

	return Platform(s[platformKey]), false
}

func (s Specifier) Name() string {
	return s[nameKey]
}

func (s Specifier) OS() string {
	return s[osKey]
}
