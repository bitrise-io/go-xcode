package destination

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/simulator"
)

func Test_getSimulatorForDestination(t *testing.T) {
	logger := log.NewLogger()

	tests := []struct {
		name                 string
		destinationSpecifier string
		want                 simulator.InfoModel
		wantErr              bool
	}{
		// {
		// 	name:                 "latest",
		// 	destinationSpecifier: "platform=iOS Simulator,name=iPhone 8 Plus,OS=16.0",
		// 	want:                 simulator.InfoModel{
		// 		// ID:     "09E082F3-0CD3-45E0-BDB3-5901B5E738B8",
		// 		// Name:   "iPhone 8 Plus",
		// 		// Status: "Shutdown",
		// 	},
		// },
		// {
		// 	name:                 "latest",
		// 	destinationSpecifier: "platform=iOS Simulator,name=iPhone 8 Plus",
		// 	want: simulator.InfoModel{
		// 		ID:     "09E082F3-0CD3-45E0-BDB3-5901B5E738B8",
		// 		Name:   "iPhone 8 Plus",
		// 		Status: "Shutdown",
		// 	},
		// },
		/*
					[33;1mattempt 0 to get simulator UDID failed with error: failed to determin latest (iOS) simulator version[0m
			[33;1mattempt 1 to get simulator UDID failed with error: failed to determin latest (iOS) simulator version[0m
			[33;1mattempt 2 to get simulator UDID failed with error: failed to determin latest (iOS) simulator version[0m
			[33;1mattempt 3 to get simulator UDID failed with error: failed to determin latest (iOS) simulator version[0m
		*/
		{
			name:                 "latest",
			destinationSpecifier: "platform=iOS Simulator,name=iPhone 85,OS=latest",
		},
		/*[33;1mattempt 3 to get simulator UDID failed with error: no simulators found for os version: iOS 15.0[0m*/
		// {
		// 	name:                 "latest",
		// 	destinationSpecifier: "platform=iOS Simulator,name=iPhone 8 Plus,OS=15.0",
		// },

		// {
		// 	name:                 "latest",
		// 	destinationSpecifier: "platform=watchOS Simulator,name=Apple Watch Series 7 - 45mm,OS=latest",
		// 	want: simulator.InfoModel{
		// 		Status: "Shutdown",
		// 	},
		// },
		// Apple Watch Series 7 - 45mm
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSimulatorForDestination(logger, tt.destinationSpecifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSimulatorForDestination() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSimulatorForDestination() = %v, want %v", got, tt.want)
			}
		})
	}
}
