package main

import (
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/certdownloder"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient"
)

func main() {
	f := devportalclient.NewClientFactory()
	devPortalClientFactory, err := f.CreateClient(devportalclient.AppleIDClient, "teamID", "build url", "build api token")
	if err != nil {
		panic(err)
	}
	certDownloader := certdownloder.NewDownloader(nil)
	autocodesign.NewCodesignAssetManager(devPortalClientFactory, certDownloader)
}
