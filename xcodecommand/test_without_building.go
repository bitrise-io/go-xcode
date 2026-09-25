package xcodecommand

// TestWithoutBuildingParams describes an `xcodebuild test-without-building` invocation.
// Zero-valued fields are omitted. Argument order follows the xcode-test-without-building step.
type TestWithoutBuildingParams struct {
	XCTestRun                      string // the .xctestrun file
	Destination                    string // a user -destination is added, not replaced
	ResultBundlePath               string
	TestRepetitionMode             TestRepetitionMode
	MaximumTestRepetitions         int
	RelaunchTestsForEachRepetition bool
	OnlyTesting                    []string // -only-testing:<id>; user entries are added
	SkipTesting                    []string // -skip-testing:<id>; user entries are added
	CollectTestDiagnostics         string   // a default: a user -collect-test-diagnostics replaces it
	AdditionalOptions              []string // the step's xcodebuild_options, shell-split
	Validation                     Validation
}

// TestWithoutBuilding renders params into a test-without-building Command.
func TestWithoutBuilding(params TestWithoutBuildingParams) (Command, error) {
	opts := Options{{Kind: Action, Name: ActionTestWithoutBuilding}}
	opts = appendValue(opts, "-xctestrun", params.XCTestRun)
	opts = appendValue(opts, "-destination", params.Destination)
	opts = appendValue(opts, "-resultBundlePath", params.ResultBundlePath)
	opts = append(opts, testRepetitionOptions(params.TestRepetitionMode, params.MaximumTestRepetitions, params.RelaunchTestsForEachRepetition)...)
	for _, id := range params.OnlyTesting {
		opts = append(opts, Option{Kind: ColonOption, Name: "-only-testing", Value: id})
	}
	for _, id := range params.SkipTesting {
		opts = append(opts, Option{Kind: ColonOption, Name: "-skip-testing", Value: id})
	}
	opts = appendValue(opts, "-collect-test-diagnostics", params.CollectTestDiagnostics)

	return assemble(opts, params.AdditionalOptions, testWithoutBuildingSpec, params.Validation)
}
