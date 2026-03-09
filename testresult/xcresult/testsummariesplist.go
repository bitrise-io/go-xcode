package xcresult

import (
	"fmt"
	"strings"
)

// TestSummaryPlist ...
type TestSummaryPlist struct {
	FormatVersion     string
	TestableSummaries []TestableSummary
}

func collapseSubtestTree(data Subtests) (tests Subtests) {
	for _, test := range data {
		if len(test.Subtests) > 0 {
			tests = append(tests, collapseSubtestTree(test.Subtests)...)
		}
		if test.TestStatus != "" {
			tests = append(tests, test)
		}
	}
	return
}

// Tests returns the collapsed tree of tests
func (summaryPlist TestSummaryPlist) Tests() ([]string, map[string]Subtests) {
	var keyOrder []string
	tests := map[string]Subtests{}
	var subTests Subtests
	for _, testableSummary := range summaryPlist.TestableSummaries {
		for _, test := range testableSummary.Tests {
			subTests = append(subTests, collapseSubtestTree(test.Subtests)...)
		}
	}
	for _, test := range subTests {
		// TestIdentifier is in a format of testID/testCase
		testID := strings.Split(test.TestIdentifier, "/")[0]
		if _, found := tests[testID]; !found {
			keyOrder = append(keyOrder, testID)
		}
		tests[testID] = append(tests[testID], test)
	}
	return keyOrder, tests
}

// Test ...
type Test struct {
	Subtests Subtests
}

// TestableSummary ...
type TestableSummary struct {
	TargetName      string
	TestKind        string
	TestName        string
	TestObjectClass string
	Tests           []Test
}

// FailureSummary ...
type FailureSummary struct {
	FileName           string
	LineNumber         int
	Message            string
	PerformanceFailure bool
}

// Subtest ...
type Subtest struct {
	Duration         float64
	TestStatus       string
	TestIdentifier   string
	TestName         string
	TestObjectClass  string
	Subtests         Subtests
	FailureSummaries []FailureSummary
}

// Failure ...
func (st Subtest) Failure() (message string) {
	prefix := ""
	for _, failure := range st.FailureSummaries {
		message += fmt.Sprintf("%s%s:%d - %s", prefix, failure.FileName, failure.LineNumber, failure.Message)
		prefix = "\n"
	}
	return
}

// Skipped ...
func (st Subtest) Skipped() bool {
	return st.TestStatus == "Skipped"
}

// Subtests ...
type Subtests []Subtest

// FailuresCount ...
func (sts Subtests) FailuresCount() (count int) {
	for _, test := range sts {
		if len(test.FailureSummaries) > 0 {
			count++
		}
	}
	return count
}

// SkippedCount ...
func (sts Subtests) SkippedCount() (count int) {
	for _, test := range sts {
		if test.Skipped() {
			count++
		}
	}
	return count
}

// TotalTime ...
func (sts Subtests) TotalTime() (time float64) {
	for _, test := range sts {
		time += test.Duration
	}
	return time
}
