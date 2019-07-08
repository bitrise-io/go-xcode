package xcarchive

import (
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/ziputil"
	"github.com/bitrise-io/go-xcode/plistutil"
	pth "github.com/bitrise-io/go-xcode/utility"
)

// IsMacOS try to find the Contents dir under the .app/.
// If its finds it the archive is MacOs. If it does not the archive is iOS.
func IsMacOS(archPath string) (bool, error) {
	log.Debugf("Checking archive is MacOS or iOS")
	infoPlistPath := filepath.Join(archPath, "Info.plist")

	plist, err := plistutil.NewPlistDataFromFile(infoPlistPath)
	if err != nil {
		return false, err
	}

	appProperties, found := plist.GetMapStringInterface("ApplicationProperties")
	if !found {
		return false, err
	}

	applicationPath, found := appProperties.GetString("ApplicationPath")
	if !found {
		return false, err
	}

	applicationPath = filepath.Join(archPath, "Products", applicationPath)
	contentsPath := filepath.Join(applicationPath, "Contents")

	exist, err := pathutil.IsDirExists(contentsPath)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func unwrapFileEmbeddedInXcarchiveDir(xcarchivePth, fileName string) (string, error) {
	// example path: xcarhive.zip/xcarhive/Products/Applications/sample.app/Contents/Info.plist
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__xcarhive__")
	if err != nil {
		return "", err
	}

	if err := ziputil.UnZip(xcarchivePth, tmpDir); err != nil {
		return "", err
	}

	applicationsPth := filepath.Join(tmpDir, filepath.Base(xcarchivePth), "Products", "Applications")
	appName := strings.TrimSuffix(filepath.Base(xcarchivePth), filepath.Ext(xcarchivePth))

	return pth.FindFileInAppDir(applicationsPth, appName, fileName)
}

// UnwrapEmbeddedMobileProvision ...
func UnwrapEmbeddedMobileProvision(xcarchivePth string) (string, error) {
	return unwrapFileEmbeddedInXcarchiveDir(xcarchivePth, "embedded.mobileprovision")
}

// UnwrapEmbeddedInfoPlist ...
func UnwrapEmbeddedInfoPlist(xcarchivePth string) (string, error) {
	return unwrapFileEmbeddedInXcarchiveDir(xcarchivePth, "Info.plist")
}

// CheckForXcarchive checks if the given zip is an xcarhive or not
func CheckForXcarchive(pth string) bool {
	filename := filepath.Base(pth)
	s := strings.Split(filename, ".")
	length := len(s)
	if length < 2 {
		return false
	}
	xc := s[length-2]
	if xc == "xcarchieve" {
		return true
	}
	return false
}
