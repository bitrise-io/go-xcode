package destination

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/hashicorp/go-version"
)

/*
  "devicetypes" : [{
  "productFamily" : "iPhone",
  "bundlePath" : "\/Applications\/Xcode-beta.app\/Contents\/Developer\/Platforms\/iPhoneOS.platform\/Library\/Developer\/CoreSimulator\/Profiles\/DeviceTypes\/iPhone 11.simdevicetype",
  "maxRuntimeVersion" : 4294967295,
  "maxRuntimeVersionString" : "65535.255.255",
  "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-11",
  "modelIdentifier" : "iPhone12,1",
  "minRuntimeVersionString" : "13.0.0",
  "minRuntimeVersion" : 851968,
  "name" : "iPhone 11"
}, ... ]
*/
type deviceType struct {
	Name          string `json:"name"`
	Identifier    string `json:"identifier"`
	ProductFamily string `json:"productFamily"`
}

/*
  "runtimes" : [
    {
      "bundlePath" : "\/Library\/Developer\/CoreSimulator\/Profiles\/Runtimes\/iOS 12.4.simruntime",
      "buildversion" : "16G73",
      "platform" : "iOS",
      "runtimeRoot" : "\/Library\/Developer\/CoreSimulator\/Profiles\/Runtimes\/iOS 12.4.simruntime\/Contents\/Resources\/RuntimeRoot",
      "identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-12-4",
      "version" : "12.4",
      "isInternal" : false,
      "isAvailable" : true,
      "name" : "iOS 12.4",
      "supportedDeviceTypes" : [
        {
          "bundlePath" : "\/Applications\/Xcode-beta.app\/Contents\/Developer\/Platforms\/iPhoneOS.platform\/Library\/Developer\/CoreSimulator\/Profiles\/DeviceTypes\/iPhone 5s.simdevicetype",
          "name" : "iPhone 5s",
          "identifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-5s",
          "productFamily" : "iPhone"
        }, ... ],
	}, ... ]
*/
type deviceRuntime struct {
	Identifier           string       `json:"identifier"`
	Platform             string       `json:"platform"`
	Version              string       `json:"version"`
	IsAvailable          bool         `json:"isAvailable"`
	Name                 string       `json:"name"`
	SupportedDeviceTypes []deviceType `json:"supportedDeviceTypes"`
}

/*
  "devices" : {
    "com.apple.CoreSimulator.SimRuntime.watchOS-7-4" : [
      {
        "availabilityError" : "runtime profile not found",
        "dataPath" : "\/Users\/lpusok\/Library\/Developer\/CoreSimulator\/Devices\/6503EC5B-2393-46F1-A947-B32677A3360F\/data",
        "dataPathSize" : 0,
        "logPath" : "\/Users\/lpusok\/Library\/Logs\/CoreSimulator\/6503EC5B-2393-46F1-A947-B32677A3360F",
        "udid" : "6503EC5B-2393-46F1-A947-B32677A3360F",
        "isAvailable" : false,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.Apple-Watch-Series-5-40mm",
        "state" : "Shutdown",
        "name" : "Apple Watch Series 5 - 40mm"
      }, ... ],
	"com.apple.CoreSimulator.SimRuntime.iOS-16-0" : [
      {
        "lastBootedAt" : "2022-06-07T11:34:18Z",
        "dataPath" : "\/Users\/lpusok\/Library\/Developer\/CoreSimulator\/Devices\/D64FA78C-5A25-4BF3-9EE8-855761042DEE\/data",
        "dataPathSize" : 311848960,
        "logPath" : "\/Users\/lpusok\/Library\/Logs\/CoreSimulator\/D64FA78C-5A25-4BF3-9EE8-855761042DEE",
        "udid" : "D64FA78C-5A25-4BF3-9EE8-855761042DEE",
        "isAvailable" : true,
        "logPathSize" : 57344,
        "deviceTypeIdentifier" : "com.apple.CoreSimulator.SimDeviceType.iPhone-8",
        "state" : "Shutdown",
        "name" : "iPhone 8"
      }, ... ]
  }
*/
type device struct {
	Name              string `json:"name"`
	TypeIdentifier    string `json:"deviceTypeIdentifier"`
	IsAvailable       bool   `json:"isAvailable,omitempty"`
	AvailabilityError string `json:"availabilityError,omitempty"`
	UDID              string `json:"udid"`
	State             string `json:"state"`
}

type deviceList struct {
	DeviceTypes []deviceType        `json:"deviceTypes"`
	Runtimes    []deviceRuntime     `json:"runtimes"`
	Devices     map[string][]device `json:"devices"`
}

func (d deviceFinder) createDevice(name, deviceTypeID, runtimeID string) error {
	var (
		args      = []string{"simctl", "create", name, deviceTypeID, runtimeID}
		createCmd = d.commandFactory.Create("xcrun", args, &command.Opts{})
	)

	d.logger.Println()
	d.logger.TDonef("$ %s", createCmd.PrintableCommandArgs())

	if out, err := createCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("device create command failed: %s", out)
		}

		return fmt.Errorf("failed to run device create command: %s", err)
	}

	return nil
}

func (d deviceFinder) debugDeviceList() error {
	listCmd := d.commandFactory.Create("xcrun", []string{"simctl", "list"}, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	d.logger.Println()
	d.logger.Donef("$ %s", listCmd.PrintableCommandArgs())

	return listCmd.Run()
}

func (d deviceFinder) parseDeviceList() (*deviceList, error) {
	var list deviceList

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

		for _, deviceList := range list.Devices {
			for _, device := range deviceList {
				if device.IsAvailable {
					return nil
				}
			}
		}

		return fmt.Errorf("no device is available")
	}); err != nil {
		return &deviceList{}, err
	}

	return &list, nil
}

func (d deviceFinder) filterDeviceList(wantedDevice Simulator) (Device, error) {
	if d.list == nil {
		return Device{}, fmt.Errorf("inconsistent state in filterDeviceList: device list should be parsed")
	}

	wantedDevice.Platform = strings.TrimSuffix(wantedDevice.Platform, " Simulator")

	runtime, err := d.filterRuntime(wantedDevice)
	if err != nil {
		return Device{}, err
	}
	runtimeID := runtime.Identifier

	devices, ok := d.list.Devices[runtimeID]
	if !ok {
		return Device{}, fmt.Errorf("runtime (%s) not found", runtimeID)
	}

	deviceTypeIdentifier, err := d.lookupDeviceTypeID(wantedDevice.Name)
	if err != nil {
		return Device{}, err
	}

	for _, device := range devices {
		if device.TypeIdentifier == deviceTypeIdentifier {
			if !device.IsAvailable {
				return Device{}, fmt.Errorf("device type (%s) with runtime OS (%s) is unavailable: %s", wantedDevice.Name, runtime.Version, device.AvailabilityError)
			}

			return Device{
				Name:   device.Name,
				ID:     device.UDID,
				Status: device.State,
				OS:     runtime.Version,
			}, nil
		}
	}

	if !runtime.isDeviceSupported(deviceTypeIdentifier) {
		return Device{}, fmt.Errorf("runtime (%s) is incompatible with device (%s)", runtimeID, deviceTypeIdentifier)
	}

	return Device{}, newMissingDeviceErr(wantedDevice.Name, deviceTypeIdentifier, runtimeID)
}

func (d deviceFinder) lookupDeviceTypeID(wantedDeviceName string) (string, error) {
	for _, dt := range d.list.DeviceTypes {
		if dt.Name == wantedDeviceName {
			return dt.Identifier, nil
		}
	}

	return "", fmt.Errorf("invalid device name (%s) provided", wantedDeviceName)
}

func isEqualVersion(wantVersion *version.Version, runtimeVersion *version.Version) bool {
	wantVersionSegments := wantVersion.Segments()
	runtimeVersionSegments := runtimeVersion.Segments()

	for i := 0; i < 2 && i < len(wantVersionSegments); i++ {
		if wantVersionSegments[i] != runtimeVersionSegments[i] {
			return false
		}
	}

	return true
}

func (d deviceFinder) filterRuntime(wanted Simulator) (deviceRuntime, error) {
	var allVersions []deviceRuntime

	for _, runtime := range d.list.Runtimes {

		if !runtime.IsAvailable {
			continue
		}

		if runtime.Platform != "" && runtime.Platform == wanted.Platform {
			allVersions = append(allVersions, runtime)

			continue
		}

		// using HasPrefix to ignore version in the name added by Xcode 11
		/*{
			"bundlePath" : "\/Library\/Developer\/CoreSimulator\/Profiles\/Runtimes\/iOS 13.1.simruntime",
			"buildversion" : "17A844",
			"runtimeRoot" : "\/Library\/Developer\/CoreSimulator\/Profiles\/Runtimes\/iOS 13.1.simruntime\/Contents\/Resources\/RuntimeRoot",
			"identifier" : "com.apple.CoreSimulator.SimRuntime.iOS-13-1",
			"version" : "13.1",
			"isAvailable" : true,
			"name" : "iOS 13.1"
		},*/
		if runtime.Platform == "" && strings.HasPrefix(runtime.Name, wanted.Platform) {
			allVersions = append(allVersions, runtime)
		}
	}

	if len(allVersions) == 0 {
		return deviceRuntime{}, fmt.Errorf("platform (%s) is unavailable", wanted.Platform)
	}

	wantLatest := wanted.OS == "latest"
	if wantLatest {
		var (
			latestVersion *version.Version
			latestRuntime deviceRuntime = allVersions[0]
		)

		for _, runtime := range allVersions {
			version, err := version.NewVersion(runtime.Version)
			if err != nil {
				return deviceRuntime{}, fmt.Errorf("failed to parse Simulator version (%s): %w", version, err)
			}

			if latestVersion == nil || version.GreaterThan(latestVersion) {
				latestVersion = version
				latestRuntime = runtime
			}
		}

		return latestRuntime, nil
	}

	wantVersion, err := version.NewVersion(wanted.OS)
	if err != nil {
		return deviceRuntime{}, fmt.Errorf("invalid Simulator version (%s) provided: %w", wanted.OS, err)
	}

	for _, runtime := range allVersions {
		runtimeVersion, err := version.NewVersion(runtime.Version)
		if err != nil {
			return deviceRuntime{}, fmt.Errorf("failed to parse Simulator version (%s): %w", runtimeVersion, err)
		}

		runtimeSegments := runtimeVersion.Segments()
		if len(runtimeSegments) < 2 {
			log.Warnf("no minor version found in Simulator version (%s)", runtime.Version)
			continue
		}

		isEqualVersion := isEqualVersion(wantVersion, runtimeVersion)
		if isEqualVersion {
			return runtime, nil
		}
	}

	return deviceRuntime{}, fmt.Errorf("runtime OS (%s) on platform (%s) is unavailable", wanted.OS, wanted.Platform)
}

func (r deviceRuntime) isDeviceSupported(wantedDeviceIdentifier string) bool {
	if len(r.SupportedDeviceTypes) != 0 {
		for _, d := range r.SupportedDeviceTypes {
			if d.Identifier == wantedDeviceIdentifier {
				return true
			}
		}

		return false
	}

	return true
}
