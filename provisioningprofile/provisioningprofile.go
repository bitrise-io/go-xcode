package provisioningprofile

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubyscript"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/plistutil"
)

const (
	notValidParameterErrorMessage = "security: SecPolicySetValue: One or more parameters passed to a function were not valid."
)

// NewPlistDataFromFile ...
func NewPlistDataFromFile(provisioningProfilePth string) (plistutil.PlistData, error) {
	cmd := command.New("security", "cms", "-D", "-i", provisioningProfilePth)

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed, error: %s", err)
	}

	// fix: security: SecPolicySetValue: One or more parameters passed to a function were not valid.
	outSplit := strings.Split(out, "\n")
	if len(outSplit) > 0 {
		if strings.Contains(outSplit[0], notValidParameterErrorMessage) {
			fixedOutSplit := outSplit[1:len(outSplit)]
			out = strings.Join(fixedOutSplit, "\n")
		}
	}
	// ---

	return plistutil.NewPlistDataFromContent(out)
}

// GetExportMethod ...
func GetExportMethod(data plistutil.PlistData) exportoptions.Method {
	_, ok := data.GetStringArray("ProvisionedDevices")
	if !ok {
		if allDevices, ok := data.GetBool("ProvisionsAllDevices"); ok && allDevices {
			return exportoptions.MethodEnterprise
		}
		return exportoptions.MethodAppStore
	}

	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if ok {
		if allow, ok := entitlements.GetBool("get-task-allow"); ok && allow {
			return exportoptions.MethodDevelopment
		}
		return exportoptions.MethodAdHoc
	}

	return exportoptions.MethodDefault
}

// GetDeveloperTeam ...
func GetDeveloperTeam(data plistutil.PlistData) string {
	entitlements, ok := data.GetMapStringInterface("Entitlements")
	if !ok {
		return ""
	}

	teamID, ok := entitlements.GetString("com.apple.developer.team-identifier")
	if !ok {
		return ""
	}
	return teamID
}

func parseBuildSettingsOut(out string) (map[string]string, error) {
	reader := strings.NewReader(out)
	scanner := bufio.NewScanner(reader)

	buildSettings := map[string]string{}
	isBuildSettings := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Build settings for") {
			isBuildSettings = true
			continue
		}
		if !isBuildSettings {
			continue
		}

		split := strings.Split(line, " = ")
		if len(split) > 1 {
			key := strings.TrimSpace(split[0])
			value := strings.TrimSpace(strings.Join(split[1:], " = "))

			buildSettings[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return map[string]string{}, err
	}

	return buildSettings, nil
}

func projectBuildSettings(projectPth, target string) (map[string]string, error) {
	args := []string{"-showBuildSettings"}
	if target != "" {
		args = append(args, "-target", target)
	}

	cmd := command.New("xcodebuild", args...)
	cmd.SetDir(filepath.Dir(projectPth))

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return map[string]string{}, err
	}

	return parseBuildSettingsOut(out)
}

const getCodeSignMappingScriptContent = `require "xcodeproj"
require "json"

def contained_projects(project_or_workspace_pth)
  project_paths = []
  if File.extname(project_or_workspace_pth) == '.xcodeproj'
    project_paths = [project_or_workspace_pth]
  else
    workspace_contents_pth = File.join(project_or_workspace_pth, 'contents.xcworkspacedata')
    workspace_contents = File.read(workspace_contents_pth)
    project_paths = workspace_contents.scan(/\"group:(.*)\"/).collect do |current_match|
      # skip cocoapods projects
      return nil if current_match.end_with?('Pods/Pods.xcodeproj')

      File.join(File.expand_path('..', project_or_workspace_pth), current_match.first)
    end
  end
  project_paths
end

def read_code_sign_map(project_or_workspace_pth)
  code_sign_map = {}

  project_paths = contained_projects(project_or_workspace_pth)
  project_paths.each do |project_path|
    project = Xcodeproj::Project.open(project_path)
    project.targets.each do |target|
      next if target.test_target_type?

      target.build_configuration_list.build_configurations.each do |build_configuration|
        target_id = target.uuid
        attributes = project.root_object.attributes['TargetAttributes']
        target_attributes = attributes[target_id]
        provisioning_style = target_attributes['ProvisioningStyle']

        bundle_identifier = build_configuration.resolve_build_setting("PRODUCT_BUNDLE_IDENTIFIER")
        provisioning_profile_specifier = build_configuration.resolve_build_setting("PROVISIONING_PROFILE_SPECIFIER")
        provisioning_profile_uuid = build_configuration.resolve_build_setting("PROVISIONING_PROFILE")

        project_code_sign_map = code_sign_map[project_path]
        project_code_sign_map = {} unless project_code_sign_map
        
        project_code_sign_map[target] = {
          :PRODUCT_BUNDLE_IDENTIFIER => bundle_identifier,
          :PROVISIONING_PROFILE_SPECIFIER => provisioning_profile_specifier,
          :PROVISIONING_PROFILE => provisioning_profile_uuid,
          :ProvisioningStyle => provisioning_style,
        }

        code_sign_map[project_path] = project_code_sign_map
      end
    end
  end

  code_sign_map
end

begin
  project_path = ENV["project_path"]
  mapping = read_code_sign_map(project_path)
  puts "#{{ :data =>  mapping }.to_json}"
rescue => e
  puts "#{{ :error => e.to_s }.to_json}"
end`

const gemfileContent = `source "https://rubygems.org"
gem "xcodeproj"
gem "json"
`

// ProjectCodeSignMapping ...
func ProjectCodeSignMapping(projectPth string) (map[string]string, error) {
	runner := rubyscript.New(getCodeSignMappingScriptContent)
	bundleInstallCmd, err := runner.BundleInstallCommand(gemfileContent, "")
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to create bundle install command, error: %s", err)
	}

	if out, err := bundleInstallCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return map[string]string{}, fmt.Errorf("bundle install failed, output: %s, error: %s", out, err)
	}

	runCmd, err := runner.RunScriptCommand()
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to create script runner command, error: %s", err)
	}
	runCmd.SetEnvs(append(runCmd.GetCmd().Env, "project_path="+projectPth)...)

	out, err := runCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to run ruby script, output: %s, error: %s", out, err)
	}

	type CodeSignProperties struct {
		BundleIdentifier             string
		ProvisioningProfileSpecifier string
		ProvisioningProfile          string
		ProvisioningStyle            string
	}
	type OutputModel struct {
		Data  map[string]map[string]CodeSignProperties
		Error string
	}
	var output OutputModel
	if err := json.Unmarshal([]byte(out), &output); err != nil {
		return map[string]string{}, fmt.Errorf("failed to unmarshal output: %s", out)
	}

	if output.Error != "" {
		return map[string]string{}, fmt.Errorf("failed to get provisioning profile - bundle id mapping, error: %s", output.Error)
	}

	bundleIDProfileMapping := map[string]string{}
	for projectPth, targetCodeSignMap := range output.Data {
		for target, codeSignProperties := range targetCodeSignMap {
			if codeSignProperties.ProvisioningStyle == "Automatic" {
				log.Warnf("failed to determine code sign settings for target: %s, error: uses Automatic code singing", target)
				continue
			}

			buildSettings, err := projectBuildSettings(projectPth, target)
			if err != nil {
				return map[string]string{}, fmt.Errorf("failed to read project build settings, error: %s", err)
			}

			buildSettingsBundleID := buildSettings["PRODUCT_BUNDLE_IDENTIFIER"]
			buildSettingsProvisioningProfileSpecifier := buildSettings["PROVISIONING_PROFILE_SPECIFIER"]
			buildSettingsProvisioningProfile := buildSettings["PROVISIONING_PROFILE"]

			bundleID := buildSettingsBundleID
			profile := codeSignProperties.ProvisioningProfileSpecifier
			if profile == "" && buildSettingsProvisioningProfileSpecifier != "" {
				profile = buildSettingsProvisioningProfileSpecifier
			}
			if profile == "" && codeSignProperties.ProvisioningProfile != "" {
				profile = codeSignProperties.ProvisioningProfile
			}
			if profile == "" && buildSettingsProvisioningProfile != "" {
				profile = buildSettingsProvisioningProfile
			}

			if profile == "" {
				return map[string]string{}, fmt.Errorf("Failed to find provisioning profile in build settings, error: no PROVISIONING_PROFILE_SPECIFIER nor PROVISIONING_PROFILE set")
			}

			bundleIDProfileMapping[bundleID] = profile
		}
	}

	return bundleIDProfileMapping, nil
}
