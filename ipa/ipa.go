package ipa

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/ziputil"
)

func findFileInPayloadDir(payloadPth, ipaName, fileName string) (string, error) {
	appDir := filepath.Join(payloadPth, ipaName+".app")

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

func unwrapFileEmbeddedInAppDir(ipaPth, fileName string) (string, error) {
	payloadPth, err := ziputil.UnZip(ipaPth)
	if err != nil {
		return "", err
	}

	ipaName := strings.TrimSuffix(filepath.Base(ipaPth), filepath.Ext(ipaPth))

	return findFileInPayloadDir(payloadPth, ipaName, fileName)
}

// UnwrapEmbeddedMobileProvision ...
func UnwrapEmbeddedMobileProvision(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInAppDir(ipaPth, "embedded.mobileprovision")
}

// UnwrapEmbeddedInfoPlist ...
func UnwrapEmbeddedInfoPlist(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInAppDir(ipaPth, "Info.plist")
}
