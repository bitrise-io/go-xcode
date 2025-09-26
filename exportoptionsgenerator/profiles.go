package exportoptionsgenerator

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/profileutil"
)

// ProvisioningProfileProvider can list profile infos.
type ProvisioningProfileProvider interface {
	ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error)
	GetDefaultProvisioningProfile() (profileutil.ProvisioningProfileInfoModel, error)
}

// LocalProvisioningProfileProvider ...
type LocalProvisioningProfileProvider struct {
	logger log.Logger
}

// ListProvisioningProfiles ...
func (p LocalProvisioningProfileProvider) ListProvisioningProfiles() ([]profileutil.ProvisioningProfileInfoModel, error) {
	profileProvider := profileutil.NewProfileProvider(fileutil.NewFileManager(), pathutil.NewPathModifier(), pathutil.NewPathChecker())
	return profileProvider.InstalledProvisioningProfileInfos(profileutil.ProfileTypeIos)
}

// GetDefaultProvisioningProfile ...
func (p LocalProvisioningProfileProvider) GetDefaultProvisioningProfile() (profileutil.ProvisioningProfileInfoModel, error) {
	defaultProfileURL := os.Getenv("BITRISE_DEFAULT_PROVISION_URL")
	if defaultProfileURL == "" {
		return profileutil.ProvisioningProfileInfoModel{}, nil
	}

	pathProvider := pathutil.NewPathProvider()
	tmpDir, err := pathProvider.CreateTempDir("tmp_default_profile")
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}

	tmpDst := filepath.Join(tmpDir, "default.mobileprovision")
	tmpDstFile, err := os.Create(tmpDst)
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}
	defer func() {
		if err := tmpDstFile.Close(); err != nil {
			p.logger.Warnf("Failed to close file (%s), error: %s", tmpDst, err)
		}
	}()

	response, err := http.Get(defaultProfileURL)
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			p.logger.Warnf("Failed to close response body, error: %s", err)
		}
	}()

	if _, err := io.Copy(tmpDstFile, response.Body); err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}

	profileProvider := profileutil.NewProfileProvider(fileutil.NewFileManager(), pathutil.NewPathModifier(), pathutil.NewPathChecker())
	defaultProfile, err := profileProvider.ProvisioningProfileInfoFromFile(tmpDst)
	if err != nil {
		return profileutil.ProvisioningProfileInfoModel{}, err
	}

	return defaultProfile, nil
}
