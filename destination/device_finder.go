package destination

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retry"
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
)

// Keep it in sync with https://github.com/bitrise-io/image-build-utils/blob/master/roles/simulators/defaults/main.yml#L14
const defaultDeviceName = "Bitrise iOS default"

// DeviceFinder is an interface that find a matching device for a given destination
type DeviceFinder interface {
	FindDevice(destination Simulator) (Device, error)
	ListDevices() (*DeviceList, error)
}

type deviceFinder struct {
	logger         log.Logger
	commandFactory command.Factory
	xcodeVersion   xcodeversion.Version

	list *DeviceList
}

// NewDeviceFinder retruns the default implementation of DeviceFinder
func NewDeviceFinder(log log.Logger, commandFactory command.Factory, xcodeVersion xcodeversion.Version) DeviceFinder {
	return &deviceFinder{
		logger:         log,
		commandFactory: commandFactory,
		xcodeVersion:   xcodeVersion,
	}
}

// ListDevices ...
func (d deviceFinder) ListDevices() (*DeviceList, error) {
	var list DeviceList

	// Retry gathering device information since xcrun simctl list can fail to show the complete device list
	// Originally added in https://github.com/bitrise-steplib/steps-xcode-test/pull/155
	if err := retry.Times(3).Wait(10 * time.Second).Try(func(attempt uint) error {
		listCmd := d.commandFactory.Create("xcrun", []string{"simctl", "list", "--json"}, &command.Opts{
			Stderr: os.Stderr,
		})

		d.logger.TDebugf("$ %s", listCmd.PrintableCommandArgs())
		output, err := listCmd.RunAndReturnTrimmedOutput()
		if err != nil {
			if errorutil.IsExitStatusError(err) {
				return fmt.Errorf("device list command failed: %w", err)
			}

			return fmt.Errorf("failed to run device list command: %w", err)
		}

		if err := json.Unmarshal([]byte(output), &list); err != nil {
			return fmt.Errorf("failed to unmarshal device list: %w, json: %s", err, output)
		}

		hasAvailableDevice := false
		for _, deviceList := range list.Devices {
			for _, device := range deviceList {
				if !device.IsAvailable {
					d.logger.Warnf("device %s is unavailable: %s", device.Name, device.AvailabilityError)
				} else {
					hasAvailableDevice = true
				}
			}
		}

		if hasAvailableDevice {
			return nil
		} else {
			return fmt.Errorf("no available device found")
		}
	}); err != nil {
		return &DeviceList{}, err
	}

	return &list, nil
}

// FindDevice returns a Simulator matching the destination
func (d deviceFinder) FindDevice(destination Simulator) (Device, error) {
	var (
		device Device
		err    error
	)

	start := time.Now()
	if d.list == nil {
		d.list, err = d.ListDevices()
	}
	if err == nil {
		device, err = d.deviceForDestination(destination)
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
		d.list, err = d.ListDevices()
	}
	if err == nil {
		device, err = d.deviceForDestination(destination)
	}

	return device, err
}
