package model3

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// retryRepetitionPrefix names the repetitions of a retry-on-failure run. Any other repetition name
// comes from running a test a fixed number of times, where every run counts.
const retryRepetitionPrefix = "Retry "

// Convert converts TestData into a TestSummary and a list of warnings.
func Convert(data *TestData) (*TestSummary, []string, error) {
	var warnings []string
	summary := TestSummary{}

	for _, testPlanNode := range data.TestNodes {
		if testPlanNode.Type != TestNodeTypeTestPlan {
			return nil, warnings, fmt.Errorf("test plan expected but got: %s", testPlanNode.Type)
		}

		testPlan := TestPlan{Name: testPlanNode.Name}

		for _, testBundleNode := range testPlanNode.Children {
			if testBundleNode.Type != TestNodeTypeUnitTestBundle && testBundleNode.Type != TestNodeTypeUITestBundle {
				return nil, warnings, fmt.Errorf("test bundle expected but got: %s", testBundleNode.Type)
			}

			testBundle := TestBundle{Name: testBundleNode.Name}

			for _, testSuiteNode := range testBundleNode.Children {
				var name string
				var testNodes []TestNode

				if testSuiteNode.Type == TestNodeTypeTestCase {
					name = testBundleNode.Name
					testNodes = []TestNode{testSuiteNode}
				} else if testSuiteNode.Type == TestNodeTypeTestSuite {
					name = testSuiteNode.Name
					testNodes = testSuiteNode.Children
				} else {
					return nil, warnings, fmt.Errorf("test suite or test case expected but got: %s", testSuiteNode.Type)
				}

				testSuite := TestSuite{Name: name}

				testCases, testCaseWarnings, err := extractTestCases(testNodes, name)
				warnings = append(warnings, testCaseWarnings...)

				if err != nil {
					return nil, warnings, err
				}

				testSuite.TestCases = testCases
				testBundle.TestSuites = append(testBundle.TestSuites, testSuite)
			}

			testPlan.TestBundles = append(testPlan.TestBundles, testBundle)
		}

		summary.TestPlans = append(summary.TestPlans, testPlan)
	}

	return &summary, warnings, nil
}

func extractTestCases(nodes []TestNode, fallbackName string) ([]TestCaseWithRetries, []string, error) {
	var testCases []TestCaseWithRetries
	var warnings []string

	for _, testCaseNode := range nodes {
		// A customer's xcresult file contained this use case where a test suite is a child of a test suite.
		if testCaseNode.Type == TestNodeTypeTestSuite {
			nestedTestCases, nestedWarnings, err := extractTestCases(testCaseNode.Children, fallbackName)
			warnings = append(warnings, nestedWarnings...)

			if err != nil {
				return nil, warnings, err
			}

			testCases = append(testCases, nestedTestCases...)

			continue
		}

		if testCaseNode.Type != TestNodeTypeTestCase {
			return nil, warnings, fmt.Errorf("test case expected but got: %s", testCaseNode.Type)
		}

		testCase, testCaseWarnings := extractTestCase(testCaseNode, "", "", fallbackName)
		warnings = append(warnings, testCaseWarnings...)

		retries, retryWarnings, err := extractRetries(testCaseNode, fallbackName)
		if err != nil {
			return nil, warnings, err
		}
		warnings = append(warnings, retryWarnings...)

		argumentRetries, retryMode, argumentWarnings := extractArgumentRetries(testCaseNode, fallbackName)
		warnings = append(warnings, argumentWarnings...)
		if len(argumentRetries) > 0 {
			retries = argumentRetries
			testCase.Result = parametrizedResult(argumentRetries, retryMode)
		}

		testCases = append(testCases, TestCaseWithRetries{
			TestCase: testCase,
			Retries:  retries,
		})
	}

	return testCases, warnings, nil
}

func extractTestCase(testNode TestNode, customNodeIdentifier, customName, fallbackClassName string) (TestCase, []string) {
	var warnings []string

	nodeIdentifier := testNode.Identifier
	if customNodeIdentifier != "" {
		nodeIdentifier = customNodeIdentifier
	}

	name := testNode.Name
	if customName != "" {
		name = customName
	}

	message, failureMessageWarnings := extractFailureMessage(testNode)
	if len(failureMessageWarnings) > 0 {
		warnings = append(warnings, failureMessageWarnings...)
	}

	return TestCase{
		Name:       name,
		ClassName:  extractClassName(nodeIdentifier, fallbackClassName),
		Identifier: nodeIdentifier,
		Time:       extractDuration(testNode.Duration),
		Result:     testNode.Result,
		Message:    message,
	}, warnings
}

func extractClassName(nodeIdentifier, fallbackClassName string) string {
	className := strings.Split(nodeIdentifier, "/")[0]
	if className == "" {
		// In rare cases the identifier is an empty string, so we need to use the test suite name which is the
		// same as the first part of the identifier in normal cases.
		className = fallbackClassName
	}

	return className
}

func extractFailureMessage(testNode TestNode) (string, []string) {
	childrenCount := len(testNode.Children)
	if childrenCount == 0 {
		return "", nil
	}

	lastChild := testNode.Children[childrenCount-1]
	if lastChild.Type == TestNodeTypeRepetition {
		return extractFailureMessage(lastChild)
	}
	if lastChild.Type == TestNodeTypeTestPlanConfig || lastChild.Type == TestNodeTypeArguments {
		return extractFailureMessageFromChildren(testNode.Children)
	}

	var warnings []string
	failureMessage := ""

	for _, child := range testNode.Children {
		if child.Type == TestNodeTypeFailureMessage {
			// The failure message appears in the Name field and not in the Details field.
			if child.Name == "" {
				warnings = append(warnings, fmt.Sprintf("'%s' type has empty name field", child.Type))
			}
			if child.Details != "" {
				warnings = append(warnings, fmt.Sprintf("'%s' type has unexpected details field", child.Type))
			}

			failureMessage += child.Name
		}
	}

	return failureMessage, warnings
}

func extractFailureMessageFromChildren(nodes []TestNode) (string, []string) {
	var warnings []string
	failureMessage := ""

	for _, node := range nodes {
		msg, wrns := extractFailureMessage(node)
		if msg != "" {
			failureMessage += fmt.Sprintf("%s: %s\n", node.Name, msg)
		}
		if wrns != nil {
			warnings = append(warnings, wrns...)
		}
	}

	return failureMessage, warnings
}

func extractDuration(text string) time.Duration {
	// Duration is in the format "123.456789s", "123,456789s" or "4m 34s", so we need to do some normalization. We
	// need to replace the comma with a dot and remove the space to be able to parse it with time.ParseDuration.
	normalized := strings.ReplaceAll(text, ",", ".")
	normalized = strings.ReplaceAll(normalized, " ", "")

	duration, err := time.ParseDuration(normalized)
	if err != nil {
		return 0
	}

	return duration
}

func extractRetries(testNode TestNode, fallbackName string) ([]TestCase, []string, error) {
	var retries []TestCase
	var warnings []string

	for _, child := range testNode.Children {
		if child.Type == TestNodeTypeRepetition {
			// Use the parent test node's identifier, instead of the repetition's identifier (1, 2, ...).
			retry, testCaseWarnings := extractTestCase(child, testNode.Identifier, testNode.Name, fallbackName)
			warnings = append(warnings, testCaseWarnings...)
			retries = append(retries, retry)
		}
	}

	return retries, warnings, nil
}

// extractArgumentRetries merges the arguments' repetitions into one attempt per run index.
func extractArgumentRetries(testNode TestNode, fallbackName string) ([]TestCase, bool, []string) {
	var attemptsByArgument [][]TestNode
	hasRepetitions := false
	retryMode := false

	for _, argument := range testNode.Children {
		if argument.Type != TestNodeTypeArguments {
			continue
		}

		var attempts []TestNode
		for _, repetition := range argument.Children {
			if repetition.Type != TestNodeTypeRepetition {
				continue
			}
			if strings.HasPrefix(repetition.Name, retryRepetitionPrefix) {
				retryMode = true
			}
			// Labels the attempt's failure message with the argument.
			repetition.Name = argument.Name
			attempts = append(attempts, repetition)
		}

		if len(attempts) > 0 {
			hasRepetitions = true
		} else {
			attempts = []TestNode{argument}
		}
		attemptsByArgument = append(attemptsByArgument, attempts)
	}

	if !hasRepetitions {
		return nil, false, nil
	}

	attemptCount := 0
	for _, attempts := range attemptsByArgument {
		attemptCount = max(attemptCount, len(attempts))
	}

	var retries []TestCase
	var warnings []string
	for i := range attemptCount {
		var attempt []TestNode
		for _, attempts := range attemptsByArgument {
			if i < len(attempts) {
				attempt = append(attempt, attempts[i])
			}
		}

		retry, attemptWarnings := mergeArgumentAttempt(testNode, attempt, fallbackName)
		warnings = append(warnings, attemptWarnings...)
		retries = append(retries, retry)
	}

	return retries, retryMode, warnings
}

func mergeArgumentAttempt(testNode TestNode, attempt []TestNode, fallbackName string) (TestCase, []string) {
	message, warnings := extractFailureMessageFromChildren(attempt)

	// Arguments run in parallel, so the slowest one is the time of the attempt.
	var duration time.Duration
	var results []TestResult
	for _, node := range attempt {
		duration = max(duration, extractDuration(node.Duration))
		results = append(results, node.Result)
	}

	return TestCase{
		Name:       testNode.Name,
		ClassName:  extractClassName(testNode.Identifier, fallbackName),
		Identifier: testNode.Identifier,
		Time:       duration,
		Result:     mergeResults(results),
		Message:    message,
	}, warnings
}

func mergeResults(results []TestResult) TestResult {
	if slices.Contains(results, TestResultFailed) {
		return TestResultFailed
	}
	mixed := slices.ContainsFunc(results, func(result TestResult) bool {
		return result != results[0]
	})
	if mixed {
		return TestResultPassed
	}

	return results[0]
}

// parametrizedResult replaces the test case's own result, which can stay failed after a retry passes.
func parametrizedResult(retries []TestCase, retryMode bool) TestResult {
	last := retries[len(retries)-1].Result
	if retryMode {
		return last
	}

	failed := slices.ContainsFunc(retries, func(retry TestCase) bool {
		return retry.Result == TestResultFailed
	})
	if failed {
		return TestResultFailed
	}

	return last
}
