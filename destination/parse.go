package destination

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/hashicorp/go-version"
)

// DeviceList ...
type DeviceList struct {
	DeviceTypes []DeviceType        `json:"deviceTypes"`
	Runtimes    []DeviceRuntime     `json:"runtimes"`
	Devices     map[string][]Device `json:"devices"`
}

/*
DeviceType ...

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
type DeviceType struct {
	Name          string `json:"name"`
	Identifier    string `json:"identifier"`
	ProductFamily string `json:"productFamily"`
}

/*
DeviceRuntime ...

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
type DeviceRuntime struct {
	Identifier           string       `json:"identifier"`
	Platform             string       `json:"platform"`
	Version              string       `json:"version"`
	IsAvailable          bool         `json:"isAvailable"`
	Name                 string       `json:"name"`
	SupportedDeviceTypes []DeviceType `json:"supportedDeviceTypes"`
}

/*
Device ...

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
type Device struct {
	Name              string `json:"name"`
	TypeIdentifier    string `json:"deviceTypeIdentifier"`
	IsAvailable       bool   `json:"isAvailable,omitempty"`
	AvailabilityError string `json:"availabilityError,omitempty"`
	UDID              string `json:"udid"`
	State             string `json:"state"`

	Type     string `json:"-"`
	Platform string `json:"-"`
	OS       string `json:"-"`
	Arch     string `json:"-"`
}

// XcodebuildDestination returns the required xcodebuild -destination flag value for a device
func (d Device) XcodebuildDestination() string {
	// `arch` doesn't seem to work together with `id`
	if d.Arch == "" {
		return fmt.Sprintf("id=%s", d.UDID)
	}

	return fmt.Sprintf("platform=%s,name=%s,OS=%s,arch=%s", d.Platform, d.Name, d.OS, d.Arch)
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

func (d deviceFinder) deviceForDestination(wantedDestination Simulator) (Device, error) {
	if d.list == nil {
		return Device{}, fmt.Errorf("inconsistent state in filterDeviceList: device list should be parsed")
	}

	wantedPlatform := wantedDestination.Platform
	wantedDestination.Platform = strings.TrimSuffix(wantedDestination.Platform, " Simulator")

	runtime, err := d.runtimeForPlatformVersion(wantedDestination.Platform, wantedDestination.OS)
	if err != nil {
		return Device{}, err
	}
	runtimeID := runtime.Identifier

	devicesOfRuntime, ok := d.list.Devices[runtimeID]
	if !ok {
		return Device{}, fmt.Errorf("no device exists for runtime %s", runtime.Name)
	}

	// As the name of the device matches the device type ('iPhone 11') for factory created devices, look up device by name.
	// If the default name is required and already created, it will use that.
	for _, device := range devicesOfRuntime {
		if device.Name == wantedDestination.Name {
			if !device.IsAvailable {
				d.logger.Warnf("device %s for %s is unavailable: %s", device.Name, runtime.Name, device.AvailabilityError)
				continue
			}

			device.Type = d.convertDeviceTypeIDToDeviceName(device.TypeIdentifier)
			device.Platform = wantedPlatform
			device.OS = runtime.Version
			device.Arch = wantedDestination.Arch

			return device, nil
		}
	}

	// Returns the first available device in case the default device name is specified, but not yet created.
	if wantedDestination.Name == defaultDeviceName {
		for _, device := range devicesOfRuntime {
			if !device.IsAvailable {
				continue
			}

			device.Type = d.convertDeviceTypeIDToDeviceName(device.TypeIdentifier)
			device.Platform = wantedPlatform
			device.OS = runtime.Version
			device.Arch = wantedDestination.Arch

			return device, nil
		}
	}

	// If there is no matching device, look up device type so we can create device in a later step
	deviceTypeID, err := d.convertDeviceNameToDeviceTypeID(wantedDestination.Name)
	if err != nil {
		return Device{}, err
	}

	if !runtime.isDeviceSupported(deviceTypeID) {
		return Device{}, fmt.Errorf("runtime (%s) is incompatible with device type (%s)", runtimeID, deviceTypeID)
	}

	return Device{}, newMissingDeviceErr(wantedDestination.Name, deviceTypeID, runtimeID)
}

func (d deviceFinder) convertDeviceNameToDeviceTypeID(wantedDeviceName string) (string, error) {
	for _, dt := range d.list.DeviceTypes {
		if dt.Name == wantedDeviceName {
			return dt.Identifier, nil
		}
	}

	return "", fmt.Errorf("invalid device name (%s) provided", wantedDeviceName)
}

// convertDeviceTypeIDToDeviceName returns the device type (e.g. iPhone 11) for logging purposes.
// The device name equals this by default, but not for all manually created devices like `Bitrise iOS default`
func (d deviceFinder) convertDeviceTypeIDToDeviceName(wantedDeviceTypeID string) string {
	for _, dt := range d.list.DeviceTypes {
		if dt.Identifier == wantedDeviceTypeID {
			return dt.Name
		}
	}

	// Should not happen. Falling back to the device type ID, as used for logging only.
	return wantedDeviceTypeID
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

func (d deviceFinder) runtimeForPlatformVersion(wantedPlatform, wantedVersion string) (DeviceRuntime, error) {
	var runtimesOfPlatform []DeviceRuntime

	for _, runtime := range d.list.Runtimes {
		if !runtime.IsAvailable {
			continue
		}

		if runtime.Platform != "" && runtime.Platform == string(wantedPlatform) {
			runtimesOfPlatform = append(runtimesOfPlatform, runtime)
			continue
		}

		// simctl reports visionOS as xrOS (as of Xcode 15.1 Beta 3)
		if (runtime.Platform == "xrOS" || runtime.Platform == "visionOS") && wantedPlatform == string(VisionOS) {
			runtimesOfPlatform = append(runtimesOfPlatform, runtime)
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
		if runtime.Platform == "" && strings.HasPrefix(runtime.Name, wantedPlatform) {
			runtimesOfPlatform = append(runtimesOfPlatform, runtime)
		}
	}

	if len(runtimesOfPlatform) == 0 {
		if wantedPlatform == string(IOS) {
			return DeviceRuntime{}, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, IOSSimulator)
		}
		if wantedPlatform == string(WatchOS) {
			return DeviceRuntime{}, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, WatchOSSimulator)
		}
		if wantedPlatform == string(TvOS) {
			return DeviceRuntime{}, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, TvOSSimulator)
		}
		if wantedPlatform == string(VisionOS) {
			return DeviceRuntime{}, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, VisionOSSimulator)
		}
		return DeviceRuntime{}, fmt.Errorf("no runtime installed for platform %s", wantedPlatform)
	}

	wantLatest := wantedVersion == "latest"
	if wantLatest {
		var (
			latestVersion *version.Version
			latestRuntime DeviceRuntime = runtimesOfPlatform[0]
		)

		for _, runtime := range runtimesOfPlatform {
			runtimeVersion, err := version.NewVersion(runtime.Version)
			if err != nil {
				return DeviceRuntime{}, fmt.Errorf("failed to parse Simulator version (%s): %w", runtimeVersion, err)
			}

			if !isRuntimeSupportedByXcode(wantedPlatform, runtimeVersion, d.xcodeVersion) {
				continue
			}

			if latestVersion == nil || runtimeVersion.GreaterThan(latestVersion) {
				latestVersion = runtimeVersion
				latestRuntime = runtime
			}
		}

		return latestRuntime, nil
	}

	semanticVersion, err := version.NewVersion(wantedVersion)
	if err != nil {
		return DeviceRuntime{}, fmt.Errorf("invalid Simulator version (%s) provided: %w", wantedVersion, err)
	}

	for _, runtime := range runtimesOfPlatform {
		runtimeVersion, err := version.NewVersion(runtime.Version)
		if err != nil {
			return DeviceRuntime{}, fmt.Errorf("failed to parse Simulator version (%s): %w", runtimeVersion, err)
		}

		runtimeSegments := runtimeVersion.Segments()
		if len(runtimeSegments) < 2 {
			d.logger.Warnf("no minor version found in Simulator version (%s)", runtime.Version)
			continue
		}

		isEqualVersion := isEqualVersion(semanticVersion, runtimeVersion)
		if isEqualVersion {
			return runtime, nil
		}
	}

	return DeviceRuntime{}, newMissingRuntimeErr(wantedPlatform, wantedVersion, runtimesOfPlatform)
}

func (r DeviceRuntime) isDeviceSupported(wantedDeviceIdentifier string) bool {
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
