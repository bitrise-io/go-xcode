package ipa

import (
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/ziputil"
	"github.com/bitrise-io/go-xcode/utility"
)

func unwrapFileEmbeddedInPayloadAppDir(ipaPth, fileName string) (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__ipa__")
	if err != nil {
		return "", err
	}

	if err := ziputil.UnZip(ipaPth, tmpDir); err != nil {
		return "", err
	}

	appDir := filepath.Join(tmpDir, "Payload", strings.TrimSuffix(filepath.Base(ipaPth), filepath.Ext(ipaPth)))

	return utility.FindFileInAppDir(appDir, fileName)
}

// UnwrapEmbeddedMobileProvision ...
func UnwrapEmbeddedMobileProvision(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInPayloadAppDir(ipaPth, "embedded.mobileprovision")
}

// UnwrapEmbeddedInfoPlist ...
func UnwrapEmbeddedInfoPlist(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInPayloadAppDir(ipaPth, "Info.plist")
}
