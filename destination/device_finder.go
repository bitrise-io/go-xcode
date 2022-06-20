package destination

import (
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
)

// Device is an available device
type Device struct {
	Name   string
	ID     string
	Status string
	OS     string
}

// DeviceFinder is an interface that find a matching device for a given destination
type DeviceFinder interface {
	GetSimulator(destination Simulator) (Device, error)
}

type deviceFinder struct {
	logger         log.Logger
	commandFactory command.Factory

	list *deviceList
}

// NewDeviceFinder retruns the default implementation of DeviceFinder
func NewDeviceFinder(log log.Logger, commandFactory command.Factory) DeviceFinder {
	return &deviceFinder{
		logger:         log,
		commandFactory: commandFactory,
	}
}

// GetSimulator returns a Simulator matching the destination
func (d deviceFinder) GetSimulator(destination Simulator) (Device, error) {
	if d.list == nil {
		list, err := d.parseDeviceList()
		if err != nil {
			return Device{}, err
		}

		d.list = &list
	}

	osVersionSplit := strings.Split(destination.OS, ".")
	if len(osVersionSplit) > 2 {
		destination.OS = strings.Join(osVersionSplit[0:2], ".")
	}

	device, err := d.filterDeviceList(destination)
	if err != nil {
		if err := d.debugDeviceList(); err != nil {
			d.logger.Warnf("failed to log device list: %s", err)
		}
	}

	return device, err
}
