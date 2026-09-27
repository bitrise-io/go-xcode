package xcodecommand

import (
	"fmt"
	"strconv"
)

// TestRepetitionMode says how xcodebuild repeats tests; the values are the test steps'
// test_repetition_mode input values.
type TestRepetitionMode string

// Test repetition modes. Every mode but none runs up to MaximumTestRepetitions times
// (-test-iterations).
const (
	TestRepetitionNone               TestRepetitionMode = "none"
	TestRepetitionUntilFailure       TestRepetitionMode = "until_failure"                // -run-tests-until-failure
	TestRepetitionRetryOnFailure     TestRepetitionMode = "retry_on_failure"             // -retry-tests-on-failure
	TestRepetitionUpUntilMaximumRuns TestRepetitionMode = "up_until_maximum_repetitions" // -test-iterations alone
)

// testRunOptions are the flags of the run half shared by test and test-without-building.
type testRunOptions struct {
	resultBundlePath               string
	repetitionMode                 TestRepetitionMode
	maximumRepetitions             int
	relaunchTestsForEachRepetition bool
	onlyTesting                    []string
	skipTesting                    []string
	collectTestDiagnostics         string
}

// options renders the run flags. A mode it does not know, a repeating mode with fewer
// than two repetitions, and relaunching without repetition are refused: xcodebuild would
// refuse the last two, and an unknown mode used to mean -test-iterations alone.
func (r testRunOptions) options() (Options, error) {
	opts := appendValue(nil, "-resultBundlePath", r.resultBundlePath)

	repeats := true
	switch r.repetitionMode {
	case "", TestRepetitionNone:
		repeats = false
	case TestRepetitionUntilFailure:
		opts = append(opts, Option{Kind: Switch, Name: "-run-tests-until-failure"})
	case TestRepetitionRetryOnFailure:
		opts = append(opts, Option{Kind: Switch, Name: "-retry-tests-on-failure"})
	case TestRepetitionUpUntilMaximumRuns:
	default:
		return nil, fmt.Errorf("unknown test repetition mode %q: use %s, %s, %s or %s", r.repetitionMode,
			TestRepetitionNone, TestRepetitionUntilFailure, TestRepetitionRetryOnFailure, TestRepetitionUpUntilMaximumRuns)
	}
	if repeats {
		if r.maximumRepetitions < 2 {
			return nil, fmt.Errorf("test repetition mode %s needs at least 2 maximum test repetitions, got %d", r.repetitionMode, r.maximumRepetitions)
		}
		opts = appendValue(opts, "-test-iterations", strconv.Itoa(r.maximumRepetitions))
	}
	if r.relaunchTestsForEachRepetition {
		if !repeats {
			return nil, fmt.Errorf("relaunching tests for each repetition needs a test repetition mode other than %s", TestRepetitionNone)
		}
		opts = appendValue(opts, "-test-repetition-relaunch-enabled", "YES")
	}

	for _, id := range r.onlyTesting {
		opts = append(opts, Option{Kind: ColonOption, Name: "-only-testing", Value: id})
	}
	for _, id := range r.skipTesting {
		opts = append(opts, Option{Kind: ColonOption, Name: "-skip-testing", Value: id})
	}
	return appendValue(opts, "-collect-test-diagnostics", r.collectTestDiagnostics), nil
}

// testRunPolicy is the policy shared by the test actions: selection flags and -destination
// are appendable (xcodebuild applies -only-testing before -skip-testing, and runs on
// every destination), -collect-test-diagnostics is a default.
func testRunPolicy(name string, extra ...rejection) actionPolicy {
	return actionPolicy{
		name:       name,
		rejections: append([]rejection{modeSwitching}, extra...),
		defaults:   []string{"-collect-test-diagnostics"},
		appendable: []string{"-only-testing", "-skip-testing", "-only-test-configuration", "-skip-test-configuration", "-destination", "-arch"},
	}
}
