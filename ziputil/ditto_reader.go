package ziputil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

type dittoReader struct {
	extractedDir string
}

// NewDittoReader ...
func NewDittoReader(archivePath string) (ReadCloser, error) {
	factory := command.NewFactory(env.NewRepository())
	tmpDir, err := pathutil.NewPathProvider().CreateTempDir("ditto_reader")
	if err != nil {
		return nil, err
	}

	cmd := factory.Create("ditto", []string{"-x", archivePath, tmpDir}, nil)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return dittoReader{
		extractedDir: tmpDir,
	}, nil
}

// ReadFile ...
func (e dittoReader) ReadFile(relPthPattern string) (File, error) {
	absPthPattern := filepath.Join(e.extractedDir, relPthPattern)
	matches, err := filepath.Glob(absPthPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find file with pattern: %s: %w", absPthPattern, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no file found with pattern: %s", absPthPattern)
	}

	pth := matches[0]
	return newFile(pth), nil
}

// Close ...
func (e dittoReader) Close() error {
	return os.RemoveAll(e.extractedDir)
}

type file struct {
	pth string
}

func newFile(pth string) File {
	return file{
		pth: pth,
	}
}

// Name ...
func (file file) Name() string {
	return file.pth
}

// Open ...
func (file file) Open() (io.ReadCloser, error) {
	return os.Open(file.pth)
}

//func (e dittoReader) AppInfoPlist() (plistutil.PlistData, error) {
//	content, pth, err := e.ReadFile("Payload/*.app/Info.plist", "app Info.plist")
//	if err != nil {
//		return nil, err
//	}
//
//	appInfoPlist, err := plistutil.NewPlistDataFromContent(string(content))
//	if err != nil {
//		return nil, fmt.Errorf("failed to parse app Info.plist (%s): %w", pth, err)
//	}
//
//	return appInfoPlist, nil
//}
//
//func (e dittoReader) ProvisioningProfileInfo() (*profileutil.ProvisioningProfileInfoModel, error) {
//	content, pth, err := e.ReadFile("Payload/*.app/embedded.mobileprovision", "embedded profile")
//	if err != nil {
//		return nil, err
//	}
//
//	embeddedProfilePKCS7, err := profileutil.ProvisioningProfileFromContent(content)
//	if err != nil {
//		return nil, fmt.Errorf("failed to parse embedded profile (%s): %w", pth, err)
//	}
//
//	embeddedProfileInfo, err := profileutil.NewProvisioningProfileInfo(*embeddedProfilePKCS7)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read embedded profile info (%s): %w", pth, err)
//	}
//
//	return &embeddedProfileInfo, nil
//}
