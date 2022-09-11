package destination

import "fmt"

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
