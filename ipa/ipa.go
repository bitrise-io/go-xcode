package ipa

import (
	"fmt"
	"path/filepath"
	"strings"

	plist "github.com/DHowett/go-plist"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/zip"
)

const (
	notValidParameterErrorMessage = "security: SecPolicySetValue: One or more parameters passed to a function were not valid."
)

func unwrapFileEmbeddedInAppDir(ipaPth, fileName string) (string, error) {
	payloadPth, err := zip.UnZip(ipaPth)
	if err != nil {
		return "", err
	}

	// Check the most common location
	baseName := strings.TrimSuffix(filepath.Base(ipaPth), filepath.Ext(ipaPth))
	appDir := filepath.Join(payloadPth, baseName+".app")

	filePth := filepath.Join(appDir, fileName)
	if exist, err := pathutil.IsPathExists(filePth); err != nil {
		return "", err
	} else if exist {
		return filePth, nil
	}
	// ---

	// It's somewhere else - let's find it!
	pattern := filepath.Join("*.app", fileName)
	if filePths, err := filepath.Glob(pattern); err != nil {
		return "", err
	} else if len(filePths) > 0 {
		return filePths[0], nil
	}
	// ---

	return "", fmt.Errorf("failed to find %s", fileName)
}

// UnwrapEmbeddedMobileProvision ...
func UnwrapEmbeddedMobileProvision(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInAppDir(ipaPth, "embedded.mobileprovision")
}

// UnwrapEmbeddedInfoPlist ...
func UnwrapEmbeddedInfoPlist(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInAppDir(ipaPth, "Info.plist")
}

// EmbeddedMobileProvisionContentJSON ...
func EmbeddedMobileProvisionContentJSON(ipaPth string) (map[string]interface{}, error) {
	embeddedMobileprovisionPth, err := UnwrapEmbeddedMobileProvision(ipaPth)
	if err != nil {
		return nil, err
	}

	cmd := command.New("security", "cms", "-D", "-i", embeddedMobileprovisionPth)

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, err
	}

	outSplit := strings.Split(out, "\n")
	if len(outSplit) > 0 {
		if strings.Contains(outSplit[0], notValidParameterErrorMessage) {
			fixedOutSplit := outSplit[1:len(outSplit)]
			out = strings.Join(fixedOutSplit, "\n")
		}
	}

	var data map[string]interface{}
	if _, err := plist.Unmarshal([]byte(out), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// EmbeddedInfoPlistContentJSON ...
func EmbeddedInfoPlistContentJSON(ipaPth string) (map[string]interface{}, error) {
	embeddedInfoPlistPth, err := UnwrapEmbeddedInfoPlist(ipaPth)
	if err != nil {
		return nil, err
	}

	infoPlistBytes, err := fileutil.ReadBytesFromFile(embeddedInfoPlistPth)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if _, err := plist.Unmarshal(infoPlistBytes, &data); err != nil {
		return nil, err
	}
	return data, nil
}
