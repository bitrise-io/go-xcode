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
				// The params are valid, so only the additional options can fail a constructor, and
				// only under Fail: Warn reports them and still builds the command.
				errs := []error{
					errOf(Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Destination: "generic/platform=iOS", Authentication: &testAuth, AdditionalOptions: args, Validation: validation})),
					errOf(Build(BuildParams{ProjectPath: "App.xcworkspace", Configuration: "Debug", AdditionalOptions: args, Validation: validation})),
					errOf(Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", ResultBundlePath: "/tmp/r.xcresult", AdditionalOptions: args, Validation: validation})),
					errOf(BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", AdditionalOptions: args, Validation: validation})),
					errOf(Test(TestParams{ProjectPath: "App.xcodeproj", Destination: "id=SIM", CollectTestDiagnostics: "never", SkipTesting: []string{"AppTests/Flaky"}, AdditionalOptions: args, Validation: validation})),
					errOf(TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "/tmp/App.xctestrun", Destination: "id=SIM", AdditionalOptions: args, Validation: validation})),
					errOf(TestWithoutBuilding(TestWithoutBuildingParams{ProjectPath: "App.xcworkspace", Scheme: "App", Destination: "id=SIM", AdditionalOptions: args, Validation: validation})),
					errOf(ExportArchive(ExportArchiveParams{ArchivePath: "/tmp/App.xcarchive", ExportPath: "/tmp/out", ExportOptionsPlist: "/tmp/o.plist", AdditionalOptions: args, Validation: validation})),
					errOf(ResolvePackages(ResolvePackagesParams{ProjectPath: "App.xcodeproj", AdditionalOptions: args})),
				}
				if validation == Warn {
					for _, err := range errs {
						require.NoError(t, err)
					}
				}
			}
		}
	})
}

func errOf(_ Command, err error) error { return err }
