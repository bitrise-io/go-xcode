package destination

import (
	"errors"
	"time"

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
	FindDevice(destination Simulator) (Device, error)
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
func (d deviceFinder) FindDevice(destination Simulator) (Device, error) {
	var (
		device Device
		err    error
	)

	start := time.Now()
	if d.list == nil {
		d.list, err = d.parseDeviceList()
	}
	if err == nil {
		device, err = d.filterDeviceList(destination)
	}

	d.logger.TDebugf("Parsed simulator list in %s", time.Since(start).Round(time.Second))
	if err == nil {
		return device, nil
	}

	var misingErr *missingDeviceErr
	if !errors.As(err, &misingErr) {
		if err := d.debugDeviceList(); err != nil {
			d.logger.Warnf("failed to log device list: %s", err)
		}

		return Device{}, err
	}

	d.logger.Infof("Creating missing device...")

	start = time.Now()
	err = d.createDevice(misingErr.name, misingErr.deviceTypeID, misingErr.runtimeID)
	d.logger.Debugf("Created device in %s", time.Since(start).Round(time.Second))

	if err == nil {
		d.list, err = d.parseDeviceList()
	}
	if err == nil {
		device, err = d.filterDeviceList(destination)
	}

	return device, err
}
