package destination

import (
	"fmt"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/simulator"
)

func getSimulatorForDestination(logger log.Logger, destinationSpecifier string) (simulator.InfoModel, error) {
	var sim simulator.InfoModel
	var osVersion string

	simulatorDestination, err := NewSimulator(destinationSpecifier)
	if err != nil {
		return simulator.InfoModel{}, fmt.Errorf("invalid destination specifier (%s): %w", destinationSpecifier, err)
	}

	platform := strings.TrimSuffix(simulatorDestination.Platform, " Simulator")
	// Retry gathering device information since xcrun simctl list can fail to show the complete device list
	if err := retry.Times(3).Wait(10 * time.Second).Try(func(attempt uint) error {
		var errGetSimulator error
		if simulatorDestination.OS == "latest" {
			simulatorDevice := simulatorDestination.Name
			if simulatorDevice == "iPad" {
				logger.Warnf("Given device (%s) is deprecated, using iPad Air (3rd generation)...", simulatorDevice)
				simulatorDevice = "iPad Air (3rd generation)"
			}

			sim, osVersion, errGetSimulator = simulator.GetLatestSimulatorInfoAndVersion(platform, simulatorDevice)
		} else {
			normalizedOsVersion := simulatorDestination.OS
			osVersionSplit := strings.Split(normalizedOsVersion, ".")
			if len(osVersionSplit) > 2 {
				normalizedOsVersion = strings.Join(osVersionSplit[0:2], ".")
			}
			osVersion = fmt.Sprintf("%s %s", platform, normalizedOsVersion)

			sim, errGetSimulator = simulator.GetSimulatorInfo(osVersion, simulatorDestination.Name)
		}

		if errGetSimulator != nil {
			logger.Warnf("attempt %d to get simulator UDID failed with error: %s", attempt, errGetSimulator)
		}

		return errGetSimulator
	}); err != nil {
		return simulator.InfoModel{}, fmt.Errorf("simulator UDID lookup failed: %w", err)
	}

	logger.Infof("Simulator infos")
	logger.Printf("* simulator_name: %s, version: %s, UDID: %s, status: %s", sim.Name, osVersion, sim.ID, sim.Status)

	return sim, nil
}
