package destination_test

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/destination"
	"github.com/bitrise-io/go-xcode/v2/xcodeversion"
)

func ExampleDeviceFinder() {
	commandFactory := command.NewFactory(env.NewRepository())
	logger := log.NewLogger()
	logger.EnableDebugLog(true)

	deviceFinder := destination.NewDeviceFinder(logger, commandFactory, xcodeversion.Version{Major: 26})

	got, err := deviceFinder.FindDevice(destination.Simulator{
		Platform: "iOS Simulator",
		OS:       "latest",
		Name:     "iPhone 16",
	})

	if err != nil {
		logger.Errorf("failed to find device: %s", err)
		return
	}

	logger.Infof("%+v", got.TypeIdentifier)
}
