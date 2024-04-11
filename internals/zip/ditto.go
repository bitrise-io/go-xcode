package zip

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/bitrise-io/go-xcode/profileutil"
)

type Reader interface {
}

type DittoExtractor struct {
	extractedDir string
	logger       log.Logger
}

func NewDittoExtractor(archivePath string, logger log.Logger) (*DittoExtractor, error) {
	factory := command.NewFactory(env.NewRepository())
	tmpDir, err := pathutil.NewPathProvider().CreateTempDir("ditto_reader")
	if err != nil {
		return nil, err
	}

	cmd := factory.Create("ditto", []string{"-x", archivePath, tmpDir}, nil)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &DittoExtractor{
		extractedDir: tmpDir,
		logger:       logger,
	}, nil
}

func (e DittoExtractor) AppInfoPlist() (plistutil.PlistData, error) {
	content, pth, err := e.readFile("Payload/*.app/Info.plist", "app Info.plist")
	if err != nil {
		return nil, err
	}

	appInfoPlist, err := plistutil.NewPlistDataFromContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse app Info.plist (%s): %w", pth, err)
	}

	return appInfoPlist, nil
}

func (e DittoExtractor) ProvisioningProfileInfo() (*profileutil.ProvisioningProfileInfoModel, error) {
	content, pth, err := e.readFile("Payload/*.app/embedded.mobileprovision", "embedded profile")
	if err != nil {
		return nil, err
	}

	embeddedProfilePKCS7, err := profileutil.ProvisioningProfileFromContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded profile (%s): %w", pth, err)
	}

	embeddedProfileInfo, err := profileutil.NewProvisioningProfileInfo(*embeddedProfilePKCS7)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded profile info (%s): %w", pth, err)
	}

	return &embeddedProfileInfo, nil
}

func (e DittoExtractor) readFile(relPthPattern string, fileName string) ([]byte, string, error) {
	absPthPattern := filepath.Join(e.extractedDir, relPthPattern)
	matches, err := filepath.Glob(absPthPattern)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find %s with pattern: %s: %w", fileName, absPthPattern, err)
	}
	if len(matches) == 0 {
		return nil, "", fmt.Errorf("no %s found with pattern: %s", fileName, absPthPattern)
	}

	pth := matches[0]
	reader, err := os.Open(pth)
	if err != nil {
		return nil, pth, fmt.Errorf("failed to open %s (%s): %w", fileName, pth, err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			e.logger.Warnf("failed to close %s (%s): %s", fileName, pth, err)
		}
	}()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, pth, fmt.Errorf("failed to read %s (%s): %w", fileName, pth, err)
	}

	return content, pth, nil
}
