package destination

import (
	"fmt"
	"os"
	"slices"
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

// runtimeDevice is a Device together with the runtime it belongs to.
type runtimeDevice struct {
	Device
	runtime DeviceRuntime
}

func (d deviceFinder) deviceForDestination(wantedDestination Simulator) (Device, error) {
	if d.list == nil {
		return Device{}, fmt.Errorf("inconsistent state in filterDeviceList: device list should be parsed")
	}

	wantedPlatform := wantedDestination.Platform
	wantedDestination.Platform = strings.TrimSuffix(wantedDestination.Platform, " Simulator")

	runtimes, err := d.runtimesForPlatformVersion(wantedDestination.Platform, wantedDestination.OS)
	if err != nil {
		return Device{}, err
	}

	// More than one runtime can have the same version: Xcode 27.2 ships a generic iOS 27.2 runtime and a device-specific one
	// (iOS-27-2-iPhone19-4) supporting only the iPhone Duo, each with its own devices.
	devices, err := d.devicesOfRuntimes(runtimes)
	if err != nil {
		return Device{}, err
	}

	// As the name of the device matches the device type ('iPhone 11') for factory created devices, look up device by name.
	// If the default name is required and already created, it will use that.
	for _, device := range devices {
		if device.Name == wantedDestination.Name {
			if !device.IsAvailable {
				d.logger.Warnf("device %s for %s is unavailable: %s", device.Name, device.runtime.Name, device.AvailabilityError)
				continue
			}

			return d.describeDevice(device, wantedPlatform, wantedDestination.Arch), nil
		}
	}

	// Returns the first available device in case the default device name is specified, but not yet created.
	// The devices of the generic runtime come first (see runtimesForPlatformVersion), so a device-specific runtime
	// (iOS-27-2-iPhone19-4, iPhone Duo only) is used only as a last resort.
	if wantedDestination.Name == defaultDeviceName {
		for _, device := range devices {
			if !device.IsAvailable {
				continue
			}

			return d.describeDevice(device, wantedPlatform, wantedDestination.Arch), nil
		}
	}

	// If there is no matching device, look up device type so we can create device in a later step
	deviceTypeID, err := d.convertDeviceNameToDeviceTypeID(wantedDestination.Name)
	if err != nil {
		return Device{}, err
	}

	for _, runtime := range runtimes {
		if runtime.isDeviceSupported(deviceTypeID) {
			return Device{}, newMissingDeviceErr(wantedDestination.Name, deviceTypeID, runtime.Identifier)
		}
	}

	return Device{}, fmt.Errorf("runtime (%s) is incompatible with device type (%s)", strings.Join(runtimeIdentifiers(runtimes), ", "), deviceTypeID)
}

// devicesOfRuntimes merges the device lists of the given runtimes, keeping the order of the runtimes.
func (d deviceFinder) devicesOfRuntimes(runtimes []DeviceRuntime) ([]runtimeDevice, error) {
	var (
		devices []runtimeDevice
		found   bool
	)

	for _, runtime := range runtimes {
		devicesOfRuntime, ok := d.list.Devices[runtime.Identifier]
		if !ok {
			continue
		}

		found = true
		for _, device := range devicesOfRuntime {
			devices = append(devices, runtimeDevice{Device: device, runtime: runtime})
		}
	}

	if !found {
		return nil, fmt.Errorf("no device exists for runtime %s", strings.Join(runtimeIdentifiers(runtimes), ", "))
	}

	return devices, nil
}

// describeDevice fills the fields of a found device that are not part of the simctl output.
func (d deviceFinder) describeDevice(device runtimeDevice, platform, arch string) Device {
	found := device.Device
	found.Type = d.convertDeviceTypeIDToDeviceName(device.TypeIdentifier)
	found.Platform = platform
	found.OS = device.runtime.Version
	found.Arch = arch

	return found
}

func runtimeIdentifiers(runtimes []DeviceRuntime) []string {
	identifiers := make([]string, 0, len(runtimes))
	for _, runtime := range runtimes {
		identifiers = append(identifiers, runtime.Identifier)
	}

	return identifiers
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

// runtimesForPlatformVersion returns every available runtime of the platform with the wanted major.minor version
// (for latest: the newest major.minor version supported by the current Xcode), the generic runtime first.
func (d deviceFinder) runtimesForPlatformVersion(wantedPlatform, wantedVersion string) ([]DeviceRuntime, error) {
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
			return nil, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, IOSSimulator)
		}
		if wantedPlatform == string(WatchOS) {
			return nil, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, WatchOSSimulator)
		}
		if wantedPlatform == string(TvOS) {
			return nil, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, TvOSSimulator)
		}
		if wantedPlatform == string(VisionOS) {
			return nil, fmt.Errorf("the platform %s is unavailable. Did you mean %s?", wantedPlatform, VisionOSSimulator)
		}
		return nil, fmt.Errorf("no runtime installed for platform %s", wantedPlatform)
	}

	if wantedVersion == "latest" {
		return d.latestRuntimes(wantedPlatform, runtimesOfPlatform)
	}

	semanticVersion, err := version.NewVersion(wantedVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid Simulator version (%s) provided: %w", wantedVersion, err)
	}

	matchingRuntimes, err := runtimesOfVersion(semanticVersion, runtimesOfPlatform)
	if err != nil {
		return nil, err
	}
	if len(matchingRuntimes) == 0 {
		return nil, newMissingRuntimeErr(wantedPlatform, wantedVersion, runtimesOfPlatform)
	}

	return matchingRuntimes, nil
}

// latestRuntimes returns the runtimes of the newest major.minor version supported by the current Xcode.
func (d deviceFinder) latestRuntimes(wantedPlatform string, runtimesOfPlatform []DeviceRuntime) ([]DeviceRuntime, error) {
	var (
		latestVersion     *version.Version
		supportedRuntimes []DeviceRuntime
	)

	for _, runtime := range runtimesOfPlatform {
		runtimeVersion, err := version.NewVersion(runtime.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Simulator version (%s): %w", runtime.Version, err)
		}

		if !isRuntimeSupportedByXcode(wantedPlatform, runtimeVersion, d.xcodeVersion) {
			continue
		}

		supportedRuntimes = append(supportedRuntimes, runtime)
		if latestVersion == nil || runtimeVersion.GreaterThan(latestVersion) {
			latestVersion = runtimeVersion
		}
	}

	if latestVersion == nil {
		// None of the runtimes is supported by the current Xcode, fall back to the first one.
		return runtimesOfPlatform[:1], nil
	}

	return runtimesOfVersion(latestVersion, supportedRuntimes)
}

// runtimesOfVersion returns the runtimes with the same major.minor version as the wanted one.
// A device-specific runtime (iOS-27-2-iPhone19-4) supports a single device type, the generic one every device type
// of the version, so the generic runtime comes first: it is preferred for the default device and for creating a missing device.
func runtimesOfVersion(wantedVersion *version.Version, runtimes []DeviceRuntime) ([]DeviceRuntime, error) {
	var matchingRuntimes []DeviceRuntime
	for _, runtime := range runtimes {
		runtimeVersion, err := version.NewVersion(runtime.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Simulator version (%s): %w", runtime.Version, err)
		}

		if isEqualVersion(wantedVersion, runtimeVersion) {
			matchingRuntimes = append(matchingRuntimes, runtime)
		}
	}

	slices.SortStableFunc(matchingRuntimes, func(a, b DeviceRuntime) int {
		return len(b.SupportedDeviceTypes) - len(a.SupportedDeviceTypes)
	})

	return matchingRuntimes, nil
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
