package xcodecommand_test

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-xcode/v2/xcodecommand"
)

// ExampleArchive is the call site steps-xcode-archive ends up with. Inputs the step only
// sometimes has (the API key, the xcconfig path) are assigned as-is: a nil pointer or empty
// string leaves the flag out. The step's default destination is a field; when the user's
// xcodebuild_options also carry -destination, the merge drops the default and reports it.
func ExampleArchive() {
	var apiKey *xcodecommand.Authentication // nil when the step runs without an API key
	xcconfigPath := "/tmp/temp.xcconfig"    // "" when the step has no xcconfig content
	userOptions := []string{"-destination", "generic/platform=iOS Simulator", "-quiet"}

	cmd, err := xcodecommand.Archive(xcodecommand.ArchiveParams{
		ProjectPath:       "App.xcworkspace",
		Scheme:            "App",
		Configuration:     "Release",
		Destination:       "generic/platform=iOS",
		XCConfigPath:      xcconfigPath,
		ArchivePath:       "/tmp/App.xcarchive",
		Clean:             true,
		Authentication:    apiKey,
		AdditionalOptions: userOptions,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(strings.Join(cmd.Args(), " "))
	for _, d := range cmd.Diagnostics() {
		fmt.Println("warning:", d)
	}
	// Output:
	// clean archive -workspace App.xcworkspace -scheme App -configuration Release -xcconfig /tmp/temp.xcconfig -archivePath /tmp/App.xcarchive -destination generic/platform=iOS Simulator -quiet
	// warning: "-destination generic/platform=iOS" replaced by additional option [-destination generic/platform=iOS Simulator]
}
