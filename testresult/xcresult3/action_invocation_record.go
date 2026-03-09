package xcresult3

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/testreport"
)

// ActionsInvocationRecord ...
type ActionsInvocationRecord struct {
	Actions struct {
		Values []struct {
			ActionResult struct {
				TestsRef struct {
					ID struct {
						Value string `json:"_value"`
					} `json:"id"`
				} `json:"testsRef"`
			} `json:"actionResult"`
		} `json:"_values"`
	} `json:"actions"`

	Issues Issues `json:"issues"`
}

// Issues ...
type Issues struct {
	TestFailureSummaries TestFailureSummaries `json:"testFailureSummaries"`
}

// TestFailureSummaries ...
type TestFailureSummaries struct {
	Values []TestFailureSummary `json:"_values"`
}

// TestFailureSummary ...
type TestFailureSummary struct {
	DocumentLocationInCreatingWorkspace DocumentLocationInCreatingWorkspace `json:"documentLocationInCreatingWorkspace"`
	Message                             Message                             `json:"message"`
	ProducingTarget                     ProducingTarget                     `json:"producingTarget"`
	TestCaseName                        TestCaseName                        `json:"testCaseName"`
}

// URL ...
type URL struct {
	Value string `json:"_value"`
}

// DocumentLocationInCreatingWorkspace ...
type DocumentLocationInCreatingWorkspace struct {
	URL URL `json:"url"`
}

// ProducingTarget ...
type ProducingTarget struct {
	Value string `json:"_value"`
}

// TestCaseName ...
type TestCaseName struct {
	Value string `json:"_value"`
}

// Message ...
type Message struct {
	Value string `json:"_value"`
}

func testCaseMatching(test ActionTestSummaryGroup, testCaseName string) bool {
	class, method := test.references()

	return testCaseName == class+"."+method ||
		testCaseName == fmt.Sprintf("-[%s %s]", class, method)
}

// failure returns the ActionTestSummaryGroup's failure reason from the ActionsInvocationRecord.
func (r ActionsInvocationRecord) failure(test ActionTestSummaryGroup, testSuite testreport.TestSuite) string {
	for _, failureSummary := range r.Issues.TestFailureSummaries.Values {
		if failureSummary.ProducingTarget.Value == testSuite.Name && testCaseMatching(test, failureSummary.TestCaseName.Value) {
			file, line := failureSummary.fileAndLineNumber()
			return fmt.Sprintf("%s:%s - %s", file, line, failureSummary.Message.Value)
		}
	}
	return ""
}

// fileAndLineNumber unwraps the file path and line number descriptor from a given ActionTestSummaryGroup's.
func (s TestFailureSummary) fileAndLineNumber() (file string, line string) {
	// file:\/\/\/Users\/bitrisedeveloper\/Develop\/ios\/Xcode11Test\/Xcode11TestUITests\/Xcode11TestUITests.swift#CharacterRangeLen=0&EndingLineNumber=42&StartingLineNumber=42
	if s.DocumentLocationInCreatingWorkspace.URL.Value != "" {
		i := strings.LastIndex(s.DocumentLocationInCreatingWorkspace.URL.Value, "#")
		if i > -1 && i+1 < len(s.DocumentLocationInCreatingWorkspace.URL.Value) {
			return s.DocumentLocationInCreatingWorkspace.URL.Value[:i], s.DocumentLocationInCreatingWorkspace.URL.Value[i+1:]
		}
	}
	return
}
