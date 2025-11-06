package autocodesign_test

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/certdownloader"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/codesignasset"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/keychain"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/localcodesignasset"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/projectmanager"
	"github.com/bitrise-io/go-xcode/v2/codesign"
	"github.com/bitrise-io/go-xcode/v2/devportalservice"
)

type config struct {
	BuildURL            string
	BuildAPIToken       string
	TeamID              string
	KeychainPath        string
	KeychainPassword    stepconf.Secret
	ProjectPath         string
	Scheme              string
	Configuration       string
	DistributionType    autocodesign.DistributionType
	RegisterTestDevices bool
	MinProfileDaysValid int
	VerboseLog          bool
}

func Example() {
	cfg := config{
		DistributionType: autocodesign.Development,
	}
	var authClientType codesign.AuthType
	certsWithPrivateKey := []certdownloader.CertificateAndPassphrase{}

	logger := log.NewLogger()
	enRepo := env.NewRepository()
	filemanager := fileutil.NewFileManager()
	projectFactory := projectmanager.NewFactory(logger, enRepo, projectmanager.BuildActionArchive)

	f := devportalclient.NewFactory(logger, filemanager)
	connection, err := f.CreateBitriseConnection(cfg.BuildURL, cfg.BuildAPIToken)
	if err != nil {
		panic(err)
	}

	var authType codesign.AuthType
	switch authClientType {
	case codesign.APIKeyAuth:
		authType = codesign.APIKeyAuth
	case codesign.AppleIDAuth:
		authType = codesign.AppleIDAuth
	default:
		panic("missing implementation")
	}

	authConfig, err := codesign.SelectConnectionCredentials(authType, connection, codesign.ConnectionOverrideInputs{}, logger)
	if err != nil {
		panic(fmt.Sprintf("could not select Apple authentication credentials: %s", err))
	}

	devPortalClient, err := f.Create(authConfig, cfg.TeamID)
	if err != nil {
		panic(err)
	}

	keychain, err := keychain.New(cfg.KeychainPath, cfg.KeychainPassword, command.NewFactory(env.NewRepository()))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize keychain: %s", err))
	}

	certDownloader := certdownloader.NewDownloader(certsWithPrivateKey, retry.NewHTTPClient().StandardClient())
	certs, err := certDownloader.GetCertificates()
	if err != nil {
		panic(fmt.Errorf("failed to download certificates: %w", err))
	}

	typeToLocalCerts, err := autocodesign.GetValidLocalCertificates(certs)
	if err != nil {
		panic(err)
	}

	profileProvider := localcodesignasset.NewProvisioningProfileProvider()
	profileConverter := localcodesignasset.NewProvisioningProfileConverter()
	localCodesignAssetManager := localcodesignasset.NewManager(profileProvider, profileConverter)
	manager := autocodesign.NewCodesignAssetManager(devPortalClient, codesignasset.NewWriter(*keychain), localCodesignAssetManager)

	// Analyzing project
	fmt.Println()
	logger.Infof("Analyzing project")
	project, err := projectFactory.Create(projectmanager.InitParams{
		ProjectOrWorkspacePath: cfg.ProjectPath,
		SchemeName:             cfg.Scheme,
		ConfigurationName:      cfg.Configuration,
		AdditionalXcodebuildShowbuildsettingsOptions: []string{},
	})
	if err != nil {
		panic(err)
	}

	appLayout, err := project.GetAppLayout(true)
	if err != nil {
		panic(err)
	}

	distribution := cfg.DistributionType
	var testDevices []devportalservice.TestDevice
	if cfg.RegisterTestDevices {
		testDevices = connection.TestDevices
	}
	codesignAssetsByDistributionType, err := manager.EnsureCodesignAssets(appLayout, autocodesign.CodesignAssetsOpts{
		DistributionType:        distribution,
		TypeToLocalCertificates: typeToLocalCerts,
		BitriseTestDevices:      testDevices,
		MinProfileValidityDays:  cfg.MinProfileDaysValid,
		VerboseLog:              cfg.VerboseLog,
	})
	if err != nil {
		panic(fmt.Sprintf("Automatic code signing failed: %s", err))
	}

	if err := project.ForceCodesignAssets(distribution, codesignAssetsByDistributionType); err != nil {
		panic(fmt.Sprintf("Failed to force codesign settings: %s", err))
	}
}
