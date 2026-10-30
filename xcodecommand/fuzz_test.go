package xcodecommand

import (
	"slices"
	"strings"
	"testing"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/stretchr/testify/require"
)

// FuzzAdditionalOptions feeds arbitrary input through the parser, its diagnostics and every
// builder, and checks the rules the package promises for any input:
//
//   - parsing round-trips: Args() gives back the arguments it was given;
//   - no builder fails on its options, and building twice gives the same result;
//   - the user's options end every command line verbatim, in the order given;
//   - what xcodebuild refuses gets no line: no finding names an Unknown option;
//   - every finding is about one of the user's options, still on the command line;
//   - a finding is informational exactly for the kinds and hints that say so;
//   - package resolution forwards the options that apply to it, in order, and reports the
//     rest in one note, or none.
//
// `go test -run=^$ -fuzz=FuzzAdditionalOptions ./xcodecommand/` explores beyond the seeds.
func FuzzAdditionalOptions(f *testing.F) {
	for _, seed := range []string{
		"", " ", "-", "-=x", "-only-testing:", "-destination", "-quiet",
		"-destination generic/platform=iOS", "-ENABLE_BITCODE=NO", "-sdk macosx",
		"CODE_SIGN_IDENTITY=Apple Distribution", "clean archive", "-only-testing:App/Test=1",
		"-KEY=a b c", "-allowProvisioningUpdates -configuration Debug", "-a:b=c -a=b:c", "=", ":", "-:",
		"-derivedDataPath -quiet", "-destination -quiet", "-test-iterations 2 -retry-tests-on-failure",
		"-allowProvisioningUpdates=YES -quiet=YES -destination=generic/platform=iOS", "-resolvePackageDependencies -json",
		"-xcconfig mine.xcconfig -xcconfig=other.xcconfig", "-allowProvisioningUpdates YES -allowProvisioningUpdates",
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
			checkFindings(t, opts.Diagnostics(), opts.Args(), false)

			// The params are valid, so only the additional options could fail a builder, and
			// none may: every option passes through and is reported, or not.
			for _, build := range []func() (Command, error){
				func() (Command, error) {
					return Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Destination: "generic/platform=iOS", Authentication: &testAuth, AdditionalOptions: opts})
				},
				func() (Command, error) {
					return Build(BuildParams{ProjectPath: "App.xcworkspace", Configuration: "Debug", AdditionalOptions: opts})
				},
				func() (Command, error) {
					return Analyze(AnalyzeParams{ProjectPath: "App.xcodeproj", ResultBundlePath: "/tmp/r.xcresult", AdditionalOptions: opts})
				},
				func() (Command, error) {
					return BuildForTesting(BuildForTestingParams{ProjectPath: "App.xcodeproj", AdditionalOptions: opts})
				},
				func() (Command, error) {
					return Test(TestParams{ProjectPath: "App.xcodeproj", Destination: "id=SIM", CollectTestDiagnostics: "never", SkipTesting: []string{"AppTests/Flaky"}, AdditionalOptions: opts})
				},
				func() (Command, error) {
					return TestWithoutBuilding(TestWithoutBuildingParams{XCTestRun: "/tmp/App.xctestrun", Destination: "id=SIM", AdditionalOptions: opts})
				},
				func() (Command, error) {
					return TestWithoutBuilding(TestWithoutBuildingParams{ProjectPath: "App.xcworkspace", Scheme: "App", Destination: "id=SIM", AdditionalOptions: opts})
				},
			} {
				cmd, err := build()
				require.NoError(t, err)
				again, err := build()
				require.NoError(t, err)
				require.Equal(t, cmd.Args(), again.Args(), "building must be deterministic")
				require.Equal(t, cmd.Diagnostics(), again.Diagnostics(), "building must be deterministic")

				require.True(t, endsWith(cmd.Args(), args), "the user's options end the command line verbatim: %q", cmd.Args())
				checkFindings(t, cmd.Diagnostics(), cmd.Args(), true)
			}

			// Package resolution: the options that apply are forwarded in order, the rest make
			// one note.
			cmd, err := ResolvePackages(ResolvePackagesParams{ProjectPath: "App.xcodeproj", AdditionalOptions: opts})
			require.NoError(t, err)
			kept := opts.Filter(func(o Option) bool {
				_, rejected := resolvePackagesPolicy.rejects(o)
				return o.Kind != Action && o.Kind != Unknown && !rejected
			})
			require.True(t, endsWith(cmd.Args(), kept.Args()), "resolution forwards what applies, in order: %q", cmd.Args())
			for _, a := range cmd.Args() {
				require.NotContains(t, knownActions, a, "no build action reaches resolution")
			}
			if len(kept) == len(opts) {
				require.Empty(t, cmd.Diagnostics())
			} else {
				require.Equal(t, []DiagnosticKind{LeftOut}, kinds(cmd.Diagnostics()))
				checkFindings(t, cmd.Diagnostics(), cmd.Args(), true)
			}
		}
	})
}

// checkFindings checks what every finding promises: it names one of the user's options,
// never an Unknown one (xcodebuild refuses those with its own message), the option is on the
// command line, and the level follows the kind, or the hint for a Step input. A command's
// findings are about options the merge marked as the user's.
func checkFindings(t *testing.T, diagnostics []Diagnostic, args []string, fromCommand bool) {
	t.Helper()
	for _, d := range diagnostics {
		require.NotEmpty(t, d.Message)
		if fromCommand {
			require.True(t, d.Option.user, "a command's finding is about the user's option: %s", d.Message)
		}
		if d.Kind != LeftOut { // the note lists what resolution leaves out, Unknown arguments included
			require.NotEqual(t, Unknown, d.Option.Kind, "what xcodebuild refuses gets no line: %s", d.Message)
			require.True(t, contains(args, d.Option.args()), "the option is on the command line: %s", d.Message)
		}
		informational := d.Kind == Override || d.Kind == RedundantOption || d.Kind == LeftOut || (d.Kind == PreferStepInput && stepInputFlags[d.Option.Name].informational)
		require.Equal(t, informational, d.Informational, d.Message)
	}
}

// endsWith says whether args ends with tail.
func endsWith(args, tail []string) bool {
	return len(tail) <= len(args) && slices.Equal(args[len(args)-len(tail):], tail)
}

// contains says whether args holds sub as a contiguous run.
func contains(args, sub []string) bool {
	for i := 0; i+len(sub) <= len(args); i++ {
		if slices.Equal(args[i:i+len(sub)], sub) {
			return true
		}
	}
	return false
}

// FuzzShellQuoted checks that a corrected form rendered into a message reads back as the same
// value when pasted into xcodebuild_options, whatever the value holds.
func FuzzShellQuoted(f *testing.F) {
	for _, seed := range []string{"", "iphoneos", "platform=iOS Simulator,name=iPhone 15", "it's here", `say "hi"`, `a\b`, "$HOME/dd", "-quiet", "clean", "a\nb", "平台=iOS"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		opts, err := ParseOptions("-flag " + shellQuoted(value))
		require.NoError(t, err)
		require.Equal(t, []string{"-flag", value}, opts.Args())
	})
}
