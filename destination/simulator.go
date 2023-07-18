package destination

import "fmt"

// Simulator ...
type Simulator struct {
	Platform string
	Name     string
	OS       string
	Arch     string
}

// NewSimulator ...
func NewSimulator(destination string) (*Simulator, error) {
	specifier, err := NewSpecifier(destination)
	if err != nil {
		return nil, err
	}

	platform, isGeneric := specifier.Platform()
	if isGeneric {
		return nil, fmt.Errorf("can't create a simulator from generic destination: %s", destination)
	}

	simulator := Simulator{
		Platform: string(platform),
		Name:     specifier.Name(),
		OS:       specifier.OS(),
		Arch:     specifier.Arch(),
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
