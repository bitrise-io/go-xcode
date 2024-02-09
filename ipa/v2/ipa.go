package v2

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/ryanuber/go-glob"
)

func unwrapFileEmbeddedInPayloadAppDir(ipaPth, fileName string) (string, error) {
	zipListing, err := zip.OpenReader(ipaPth)
	if err != nil {
		return "", fmt.Errorf("failed to open IPA file %s: %w", ipaPth, err)
	}
	defer func() {
		if err := zipListing.Close(); err != nil {
			log.Warnf("failed to close IPA file %s: %s", ipaPth, err)
		}
	}()

	var files []*zip.File
	var fileNames []string
	for _, file := range zipListing.File {
		name := file.Name
		pattern := "Payload/*.app/" + fileName
		if glob.Glob(pattern, name) {
			files = append(files, file)
			fileNames = append(fileNames, name)
		}
	}

	if len(files) == 0 {
		return "", fmt.Errorf("noe file found with name: %s", fileName)
	} else if len(files) > 1 {
		return "", fmt.Errorf("multiple files (%s) found with name: %s", strings.Join(fileNames, ", "), fileName)
	}

	file := files[0]
	r, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", file.Name, err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Warnf("failed to close %s: %s", file.Name, err)
		}
	}()

	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", file.Name, err)
	}

	return string(b), nil
}

// UnwrapEmbeddedMobileProvision ...
func UnwrapEmbeddedMobileProvision(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInPayloadAppDir(ipaPth, "embedded.mobileprovision")
}

// UnwrapEmbeddedInfoPlist ...
func UnwrapEmbeddedInfoPlist(ipaPth string) (string, error) {
	return unwrapFileEmbeddedInPayloadAppDir(ipaPth, "Info.plist")
}
