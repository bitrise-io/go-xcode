package main

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/certdownloder"
	models "github.com/bitrise-io/go-xcode/autocodesign/codesignmodels"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient"
	"github.com/bitrise-io/go-xcode/autocodesign/keychain"
	"github.com/bitrise-io/go-xcode/autocodesign/projectmanager"
)

func main() {
	f := devportalclient.NewClientFactory()

	connection, err := f.CreateBitriseConnection("build url", "build api token")
	if err != nil {
		panic(err)
	}

	devPortalClient, err := f.CreateClient(devportalclient.AppleIDClient, "teamID", connection)
	if err != nil {
		panic(err)
	}

	keychain, err := keychain.New("kc path", "kc password", command.NewFactory(env.NewRepository()))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize keychain: %s", err))
	}

	certDownloader := certdownloder.NewDownloader(nil)
	manager := autocodesign.NewCodesignAssetManager(devPortalClient, certDownloader, connection.TestDevices, *keychain)

	// Analyzing project
	fmt.Println()
	log.Infof("Analyzing project")
	project, err := projectmanager.NewProject("path", "scheme", "config")
	if err != nil {
		panic(err.Error())
	}

	appLayout, err := project.GetAppLayout(true)
	if err != nil {
		panic(err.Error())
	}

	distribution := models.Development
	codesignAssetsByDistributionType, err := manager.EnsureCodesignAssets(appLayout, autocodesign.CodesignAssetsOpts{
		DistributionType:       distribution,
		MinProfileValidityDays: 0,
		VerboseLog:             true,
	})
	if err != nil {
		panic(fmt.Sprintf("Automatic code signing failed: %s", err))
	}

	if err := project.ForceCodesignAssets(distribution, codesignAssetsByDistributionType); err != nil {
		panic(fmt.Sprintf("Failed to force codesign settings: %s", err))
	}
}
