package destination

import (
	"fmt"
	"strings"
)

const (
	platform = "platform"
	name     = "name"
	os       = "OS"
)

// Simulator ...
type Simulator struct {
	Platform string
	Name     string
	OS       string
}

// NewSimulator ...
func NewSimulator(destination string) (*Simulator, error) {
	simulator := Simulator{}
	destinationParts := strings.Split(destination, ",")

	for _, part := range destinationParts {
		keyAndValue := strings.Split(part, "=")

		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf(`could not parse "%s" because it is not a valid key=value pair in destination: %s`, keyAndValue, destination)
		}

		key := keyAndValue[0]
		value := keyAndValue[1]

		switch key {
		case platform:
			simulator.Platform = value
		case name:
			simulator.Name = value
		case os:
			simulator.OS = value
		default:
			return nil, fmt.Errorf(`could not parse key "%s" with value "%s" in destination: %s`, key, value, destination)
		}
	}

	if simulator.Platform == "" {
		return nil, fmt.Errorf(`missing key "platform" in destination: %s`, destination)
	}

	if simulator.Name == "" {
		return nil, fmt.Errorf(`missing key "name" in destination: %s`, destination)
	}

	if simulator.OS == "" {
		// OS=latest can be omitted in the destination specifier, because it's the default value.
		simulator.OS = "latest"
	}

	return &simulator, nil
}
