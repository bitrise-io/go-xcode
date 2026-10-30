package xcodecommand

import (
	"strings"
	"testing"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/stretchr/testify/require"
)

// FuzzAdditionalOptions feeds arbitrary input through the parser, its diagnostics and every
// builder: none may panic, parsing must round-trip to the arguments it was given, and no
// builder may fail on its options. `go test -run=^$ -fuzz=FuzzAdditionalOptions ./xcodecommand/`
// explores beyond the seeds.
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
		if split, err := shellquote.Split(input); err == nil {
			variants = append(variants, split)
			parsed, err := ParseOptions(input)
			require.NoError(t, err)
			require.Equal(t, parse(split), parsed, "ParseOptions is the shell split, then parse")
		}
		for _, args := range variants {
			opts := parse(args)
			if len(args) == 0 {
				require.Empty(t, opts.Args())
			} else {
				require.Equal(t, args, opts.Args(), "parsing must round-trip")
			}
			_ = opts.Diagnostics()

			// The params are valid, so only the additional options could fail a builder, and
			// none may: every option passes through and is reported instead.
			for _, err := range []error{
				errOf(Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Destination: "generic/platform=iOS", Authentication: &testAuth, AdditionalOptions: opts})),
				errOf(Build(BuildParams{ProjectPath: "App.xcworkspace", Configuration: "Debug", AdditionalOptions: opts})),
				errOf(Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", ResultBundlePath: "/tmp/r.xcresult", AdditionalOptions: opts})),
				errOf(BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", AdditionalOptions: opts})),
				errOf(Test(TestParams{ProjectPath: "App.xcodeproj", Destination: "id=SIM", CollectTestDiagnostics: "never", SkipTesting: []string{"AppTests/Flaky"}, AdditionalOptions: opts})),
				errOf(TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "/tmp/App.xctestrun", Destination: "id=SIM", AdditionalOptions: opts})),
				errOf(TestWithoutBuilding(TestWithoutBuildingParams{ProjectPath: "App.xcworkspace", Scheme: "App", Destination: "id=SIM", AdditionalOptions: opts})),
				errOf(ExportArchive(ExportArchiveParams{ArchivePath: "/tmp/App.xcarchive", ExportPath: "/tmp/out", ExportOptionsPlist: "/tmp/o.plist", AdditionalOptions: opts})),
				errOf(ResolvePackages(ResolvePackagesParams{ProjectPath: "App.xcodeproj", AdditionalOptions: opts})),
			} {
				require.NoError(t, err)
			}
		}
	})
}

func errOf(_ Command, err error) error { return err }
