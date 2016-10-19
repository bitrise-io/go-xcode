package simulator

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/cmdex"
)

// InfoModel ...
type InfoModel struct {
	Name        string
	ID          string
	Status      string
	StatusOther string
}

// OsVersionSimulatorInfosMap ...
type OsVersionSimulatorInfosMap map[string][]InfoModel // Os version - []Info map

// getSimulatorInfoFromLine ...
// a simulator info line should look like this:
//  iPhone 5s (EA1C7E48-8137-428C-A0A5-B2C63FF276EB) (Shutdown)
// or
//  iPhone 4s (51B10EBD-C949-49F5-A38B-E658F41640FF) (Shutdown) (unavailable, runtime profile not found)
func getSimulatorInfoFromLine(lineStr string) (InfoModel, error) {
	baseInfosExp := regexp.MustCompile(`(?P<deviceName>[a-zA-Z].*[a-zA-Z0-9 -]*) \((?P<simulatorID>[a-zA-Z0-9-]{36})\) \((?P<status>[a-zA-Z]*)\)`)
	baseInfosRes := baseInfosExp.FindStringSubmatch(lineStr)
	if baseInfosRes == nil {
		return InfoModel{}, fmt.Errorf("No match found")
	}

	simInfo := InfoModel{
		Name:   baseInfosRes[1],
		ID:     baseInfosRes[2],
		Status: baseInfosRes[3],
	}

	// StatusOther
	restOfTheLine := lineStr[len(baseInfosRes[0]):]
	if len(restOfTheLine) > 0 {
		statusOtherExp := regexp.MustCompile(`\((?P<statusOther>[a-zA-Z ,]*)\)`)
		statusOtherRes := statusOtherExp.FindStringSubmatch(restOfTheLine)
		if statusOtherRes != nil {
			simInfo.StatusOther = statusOtherRes[1]
		}
	}
	return simInfo, nil
}

func getOsVersionSimulatorInfosMapFromSimctlList(simctlList string) (OsVersionSimulatorInfosMap, error) {
	simulatorsByIOSVersions := OsVersionSimulatorInfosMap{}
	currIOSVersion := ""

	fscanner := bufio.NewScanner(strings.NewReader(simctlList))
	isDevicesSectionFound := false
	for fscanner.Scan() {
		aLine := fscanner.Text()

		if aLine == "== Devices ==" {
			isDevicesSectionFound = true
			continue
		}

		if !isDevicesSectionFound {
			continue
		}
		if strings.HasPrefix(aLine, "==") {
			isDevicesSectionFound = false
			continue
		}
		if strings.HasPrefix(aLine, "--") {
			iosVersionSectionExp := regexp.MustCompile(`-- (?P<iosVersionSection>.*) --`)
			iosVersionSectionRes := iosVersionSectionExp.FindStringSubmatch(aLine)
			if iosVersionSectionRes != nil {
				currIOSVersion = iosVersionSectionRes[1]
			}
			continue
		}

		simInfo, err := getSimulatorInfoFromLine(aLine)
		if err != nil {
			fmt.Println(" [!] Error scanning the line for Simulator info: ", err)
		}

		currIOSVersionSimList := simulatorsByIOSVersions[currIOSVersion]
		currIOSVersionSimList = append(currIOSVersionSimList, simInfo)
		simulatorsByIOSVersions[currIOSVersion] = currIOSVersionSimList
	}

	return simulatorsByIOSVersions, nil
}

// GetOsVersionSimulatorInfosMap ...
func GetOsVersionSimulatorInfosMap() (OsVersionSimulatorInfosMap, error) {
	cmd := cmdex.NewCommand("xcrun", "simctl", "list")
	simctlListOut, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return OsVersionSimulatorInfosMap{}, err
	}

	return getOsVersionSimulatorInfosMapFromSimctlList(simctlListOut)
}
