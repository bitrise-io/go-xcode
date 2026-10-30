package xcodecommand_test

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-xcode/v2/xcodecommand"
)

// ExampleArchive is the call site steps-xcode-archive ends up with. The step parses its
// xcodebuild_options once, when it reads its inputs, and logs what the parser finds there.
// Inputs the step only sometimes has (the API key, the xcconfig path) are assigned as-is: a
// nil pointer or empty string leaves the flag out. The step's default destination is a
// field; when the user's options also carry -destination, the merge drops the default and
// the command reports it.
func ExampleArchive() {
	userOptions, err := xcodecommand.ParseOptions("-destination 'generic/platform=iOS Simulator' -quiet")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, d := range userOptions.Diagnostics() {
		fmt.Println("warning:", d)
	}

	var apiKey *xcodecommand.Authentication // nil when the step runs without an API key
	xcconfigPath := "/tmp/temp.xcconfig"    // "" when the step has no xcconfig content
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
	// warning: "-destination generic/platform=iOS Simulator" replaces the Step's default "-destination generic/platform=iOS".
}
