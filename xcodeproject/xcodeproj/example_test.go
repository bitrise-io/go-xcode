package xcodeproj_test

import (
	"fmt"
	"log"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	logv2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
)

func ExampleFactory_Open() {
	factory := xcodeproj.NewFactory(
		logv2.NewLogger(),
		// A BuildSettingsProvider is only needed by the methods that ask xcodebuild for the
		// effective build settings of a target, so this example leaves it out.
		nil,
		fileutil.NewFileManager(),
		pathutil.NewPathModifier(),
		nil,
		xcodeproj.NewUserProvider(),
	)

	project, err := factory.Open("testdata/App.xcodeproj")
	if err != nil {
		log.Fatal(err)
	}

	for _, target := range project.Targets() {
		fmt.Println(target.Name, target.IsExecutableProduct())
	}
}
