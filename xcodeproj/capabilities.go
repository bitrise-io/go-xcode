package xcodeproj

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/bitrise-tools/go-xcode/plistutil"

	"github.com/bitrise-io/go-utils/command/rubyscript"
)

// CapabilitiesInfo ...
type CapabilitiesInfo struct {
	Capabilities []string
	Entitlements plistutil.PlistData
}

func readProjectTargetCapabilitiesMapping(projectPth string) (map[string]CapabilitiesInfo, error) {
	runner := rubyscript.New(capabilitiesScriptContent)
	bundleInstallCmd, err := runner.BundleInstallCommand(gemfileContent, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle install command, error: %s", err)
	}

	if out, err := bundleInstallCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return nil, fmt.Errorf("bundle install failed, output: %s, error: %s", out, err)
	}

	runCmd, err := runner.RunScriptCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create script runner command, error: %s", err)
	}

	envsToAppend := []string{"project=" + projectPth}
	envs := append(runCmd.GetCmd().Env, envsToAppend...)

	runCmd.SetEnvs(envs...)

	out, err := runCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run capabilities analyzer script, output: %s, error: %s", out, err)
	}

	type OutputModel struct {
		Data  map[string][]string `json:"data"`
		Error string              `json:"error"`
	}
	var output OutputModel
	if err := json.Unmarshal([]byte(out), &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %s", out)
	}

	if output.Error != "" {
		return nil, fmt.Errorf("failed to get provisioning profile - bundle id mapping, error: %s", output.Error)
	}

	targetCapabilitiesInfo := map[string]CapabilitiesInfo{}
	for targetName, capabilities := range output.Data {
		buildSettings, err := getTargetBuildSettingsWithXcodebuild(projectPth, targetName, "")
		if err != nil {
			return nil, fmt.Errorf("failed to read project build settings, error: %s", err)
		}

		entitlementsPth := buildSettings["CODE_SIGN_ENTITLEMENTS"]
		if entitlementsPth == "" {
			return nil, fmt.Errorf("no CODE_SIGN_ENTITLEMENTS found for target: %s", targetName)
		}

		projectDir := filepath.Dir(projectPth)
		entitlementsPth = filepath.Join(projectDir, entitlementsPth)

		entitlements, err := plistutil.NewPlistDataFromFile(entitlementsPth)
		if err != nil {
			return nil, fmt.Errorf("failed to parse entitlements for target: %s", targetName)
		}

		targetCapabilitiesInfo[targetName] = CapabilitiesInfo{
			Capabilities: capabilities,
			Entitlements: entitlements,
		}
	}

	return targetCapabilitiesInfo, nil
}

// ReadProjectTargetCapabilitiesMapping ...
func ReadProjectTargetCapabilitiesMapping(projectTargetsMap map[string][]string) (map[string]CapabilitiesInfo, error) {
	filteredTargetCapabilitiesMap := map[string]CapabilitiesInfo{}

	for projectPth, targets := range projectTargetsMap {
		targetCapabilitiesMap, err := readProjectTargetCapabilitiesMapping(projectPth)
		if err != nil {
			return nil, err
		}

		for target, capabilities := range targetCapabilitiesMap {
			for _, targetToCare := range targets {
				if target == targetToCare {
					filteredTargetCapabilitiesMap[target] = capabilities
				}
			}
		}
	}

	return filteredTargetCapabilitiesMap, nil
}
