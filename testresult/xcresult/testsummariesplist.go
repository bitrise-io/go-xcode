package xcresult

import (
	"fmt"
	"strings"
)

type testSummaryPlist struct {
	FormatVersion     string
	TestableSummaries []testableSummary
}

func collapsesubtestTree(data subtests) (tests subtests) {
	for _, test := range data {
		if len(test.Subtests) > 0 {
			tests = append(tests, collapsesubtestTree(test.Subtests)...)
		}
		if test.TestStatus != "" {
			tests = append(tests, test)
		}
	}
	return
}

func (summaryPlist testSummaryPlist) tests() ([]string, map[string]subtests) {
	var keyOrder []string
	tests := map[string]subtests{}
	var subTests subtests
	for _, testableSummary := range summaryPlist.TestableSummaries {
		for _, test := range testableSummary.Tests {
			subTests = append(subTests, collapsesubtestTree(test.subtests)...)
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

type test struct {
	subtests subtests
}

type testableSummary struct {
	TargetName      string
	TestKind        string
	TestName        string
	TestObjectClass string
	Tests           []test
}

type failureSummary struct {
	FileName           string
	LineNumber         int
	Message            string
	PerformanceFailure bool
}

type subtest struct {
	Duration         float64
	TestStatus       string
	TestIdentifier   string
	TestName         string
	TestObjectClass  string
	Subtests         subtests
	FailureSummaries []failureSummary
}

func (st subtest) failure() (message string) {
	prefix := ""
	for _, failure := range st.FailureSummaries {
		message += fmt.Sprintf("%s%s:%d - %s", prefix, failure.FileName, failure.LineNumber, failure.Message)
		prefix = "\n"
	}
	return
}

func (st subtest) skipped() bool {
	return st.TestStatus == "Skipped"
}

type subtests []subtest

func (sts subtests) failuresCount() (count int) {
	for _, test := range sts {
		if len(test.FailureSummaries) > 0 {
			count++
		}
	}
	return count
}

func (sts subtests) skippedCount() (count int) {
	for _, test := range sts {
		if test.skipped() {
			count++
		}
	}
	return count
}

func (sts subtests) totalTime() (time float64) {
	for _, test := range sts {
		time += test.Duration
	}
	return time
}
