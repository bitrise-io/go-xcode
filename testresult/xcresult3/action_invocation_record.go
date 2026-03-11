package xcresult3

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/testreport"
)

type actionsInvocationRecord struct {
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

	Issues issues `json:"issues"`
}

type issues struct {
	TestFailureSummaries testFailureSummaries `json:"testFailureSummaries"`
}

type testFailureSummaries struct {
	Values []testFailureSummary `json:"_values"`
}

type testFailureSummary struct {
	DocumentLocationInCreatingWorkspace documentLocationInCreatingWorkspace `json:"documentLocationInCreatingWorkspace"`
	Message                             message                             `json:"message"`
	ProducingTarget                     producingTarget                     `json:"producingTarget"`
	TestCaseName                        testCaseName                        `json:"testCaseName"`
}

type url struct {
	Value string `json:"_value"`
}

type documentLocationInCreatingWorkspace struct {
	URL url `json:"url"`
}

type producingTarget struct {
	Value string `json:"_value"`
}

type testCaseName struct {
	Value string `json:"_value"`
}

type message struct {
	Value string `json:"_value"`
}

func testCaseMatching(test actionTestSummaryGroup, tcName string) bool {
	class, method := test.references()

	return tcName == class+"."+method ||
		tcName == fmt.Sprintf("-[%s %s]", class, method)
}

// failure returns the failure reason for the given test from the invocation record.
func (r actionsInvocationRecord) failure(test actionTestSummaryGroup, testSuite testreport.TestSuite) string {
	for _, failureSummary := range r.Issues.TestFailureSummaries.Values {
		if failureSummary.ProducingTarget.Value == testSuite.Name && testCaseMatching(test, failureSummary.TestCaseName.Value) {
			file, line := failureSummary.fileAndLineNumber()
			return fmt.Sprintf("%s:%s - %s", file, line, failureSummary.Message.Value)
		}
	}
	return ""
}

// fileAndLineNumber unwraps the file path and line number from the document location URL.
func (s testFailureSummary) fileAndLineNumber() (file string, line string) {
	// file:\/\/\/Users\/bitrisedeveloper\/Develop\/ios\/Xcode11Test\/Xcode11TestUITests\/Xcode11TestUITests.swift#CharacterRangeLen=0&EndingLineNumber=42&StartingLineNumber=42
	if s.DocumentLocationInCreatingWorkspace.URL.Value != "" {
		i := strings.LastIndex(s.DocumentLocationInCreatingWorkspace.URL.Value, "#")
		if i > -1 && i+1 < len(s.DocumentLocationInCreatingWorkspace.URL.Value) {
			return s.DocumentLocationInCreatingWorkspace.URL.Value[:i], s.DocumentLocationInCreatingWorkspace.URL.Value[i+1:]
		}
	}
	return
}
