package destination

import (
	"fmt"
	"strings"
)

// missingDeviceErr is raised when the selected runtime is available, but the device doesn't exist for that runtime.
// We can recover from this error by creating the given device.
type missingDeviceErr struct {
	name, deviceTypeID, runtimeID string
}

func newMissingDeviceErr(name, deviceTypeID, runtimeID string) *missingDeviceErr {
	return &missingDeviceErr{
		name:         name,
		deviceTypeID: deviceTypeID,
		runtimeID:    runtimeID,
	}
}

func (e *missingDeviceErr) Error() string {
	return fmt.Sprintf("device (%s) with runtime (%s) is not yet created", e.name, e.runtimeID)
}

func newMissingRuntimeErr(platform, version string, availableRuntimes []DeviceRuntime) error {
	runtimeList := prettyRuntimeList(availableRuntimes)
	return fmt.Errorf("%s %s is not installed. Choose one of the available %s runtimes: \n%s", platform, version, platform, runtimeList)
}

func prettyRuntimeList(runtimes []DeviceRuntime) string {
	var items []string
	for _, runtime := range runtimes {
		items = append(items, fmt.Sprintf("- %s", runtime.Name))
	}
	return strings.Join(items, "\n")

}
