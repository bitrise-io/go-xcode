package autocodesign_test

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/certdownloader"
	"github.com/bitrise-io/go-xcode/autocodesign/codesignasset"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient"
	"github.com/bitrise-io/go-xcode/autocodesign/keychain"
	"github.com/bitrise-io/go-xcode/autocodesign/localcodesignasset"
	"github.com/bitrise-io/go-xcode/autocodesign/projectmanager"
	"github.com/bitrise-io/go-xcode/devportalservice"
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
	authClientType := devportalclient.APIKeyClient
	certsWithPrivateKey := []certdownloader.CertificateAndPassphrase{}

	f := devportalclient.NewClientFactory()
	connection, err := f.CreateBitriseConnection(cfg.BuildURL, cfg.BuildAPIToken)
	if err != nil {
		panic(err)
	}

	devPortalClient, err := f.CreateClient(authClientType, cfg.TeamID, connection)
	if err != nil {
		panic(err)
	}

	keychain, err := keychain.New(cfg.KeychainPath, cfg.KeychainPassword, command.NewFactory(env.NewRepository()))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize keychain: %s", err))
	}

	certDownloader := certdownloader.NewDownloader(certsWithPrivateKey, retry.NewHTTPClient().StandardClient())

	profileProvider := localcodesignasset.LocalProvisioningProfileProvider{}
	localCodesignAssetManager := localcodesignasset.NewManager(profileProvider)
	manager := autocodesign.NewCodesignAssetManager(devPortalClient, certDownloader, codesignasset.NewWriter(*keychain), localCodesignAssetManager)

	// Analyzing project
	fmt.Println()
	log.Infof("Analyzing project")
	project, err := projectmanager.NewProject(projectmanager.InitParams{
		ProjectOrWorkspacePath: cfg.ProjectPath,
		SchemeName:             cfg.Scheme,
		ConfigurationName:      cfg.Configuration,
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
		DistributionType:       distribution,
		BitriseTestDevices:     testDevices,
		MinProfileValidityDays: cfg.MinProfileDaysValid,
		VerboseLog:             cfg.VerboseLog,
	})
	if err != nil {
		panic(fmt.Sprintf("Automatic code signing failed: %s", err))
	}

	if err := project.ForceCodesignAssets(distribution, codesignAssetsByDistributionType); err != nil {
		panic(fmt.Sprintf("Failed to force codesign settings: %s", err))
	}
}
