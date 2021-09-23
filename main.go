package main

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/certdownloder"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient"
	"github.com/bitrise-io/go-xcode/autocodesign/keychain"
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
	autocodesign.NewCodesignAssetManager(devPortalClient, certDownloader, connection.TestDevices, *keychain)
}
