package xcodecommand

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzAdditionalOptions feeds arbitrary input through the parser, its diagnostics and every
// constructor: none may panic, and parsing must round-trip to the arguments it was given.
// `go test -run=^$ -fuzz=FuzzAdditionalOptions ./xcodecommand/` explores beyond the seeds.
func FuzzAdditionalOptions(f *testing.F) {
	for _, seed := range []string{
		"", " ", "-", "-=x", "-only-testing:", "-destination", "-quiet",
		"-destination generic/platform=iOS", "-ENABLE_BITCODE=NO", "-sdk macosx",
		"CODE_SIGN_IDENTITY=Apple Distribution", "clean archive", "-only-testing:App/Test=1",
		"-KEY=a b c", "-allowProvisioningUpdates -configuration Debug", "-a:b=c -a=b:c", "=", ":", "-:",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		variants := [][]string{strings.Fields(input), {input}}
		if split, err := SplitAdditionalOptions(input); err == nil {
			variants = append(variants, split)
		}
		for _, args := range variants {
			opts := ParseAdditionalOptions(args)
			if len(args) == 0 {
				require.Empty(t, opts.Args())
			} else {
				require.Equal(t, args, opts.Args(), "parsing must round-trip")
			}
			_ = opts.Diagnostics()

			for _, validation := range []Validation{Warn, Fail} {
				_, _ = Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Destination: "generic/platform=iOS", Authentication: &testAuth, AdditionalOptions: args, Validation: validation})
				_, _ = Build(BuildParams{ProjectPath: "App.xcworkspace", Configuration: "Debug", AdditionalOptions: args, Validation: validation})
				_, _ = Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", ResultBundlePath: "/tmp/r.xcresult", AdditionalOptions: args, Validation: validation})
				_, _ = BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", AdditionalOptions: args, Validation: validation})
				_, _ = Test(TestParams{ProjectPath: "App.xcodeproj", Destination: "id=SIM", CollectTestDiagnostics: "never", SkipTesting: []string{"AppTests/Flaky"}, AdditionalOptions: args, Validation: validation})
				_, _ = TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "/tmp/App.xctestrun", Destination: "id=SIM", AdditionalOptions: args, Validation: validation})
				_, _ = ExportArchive(ExportArchiveParams{ArchivePath: "/tmp/App.xcarchive", ExportPath: "/tmp/out", ExportOptionsPlist: "/tmp/o.plist", AdditionalOptions: args, Validation: validation})
				_, _ = ResolvePackages(ResolvePackagesParams{ProjectPath: "App.xcodeproj", AdditionalOptions: args})
				_, _ = showBuildSettings(showBuildSettingsParams{projectPath: "App.xcodeproj", scheme: "App", additionalOptions: args, validation: validation})
			}
		}
	})
}
