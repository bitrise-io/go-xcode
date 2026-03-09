package model3

import (
	"fmt"
	"strings"
	"time"
)

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

		testCases = append(testCases, TestCaseWithRetries{
			TestCase: testCase,
			Retries:  retries,
		})
	}

	return testCases, warnings, nil
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

func extractFailureMessageFromTestPlanConfigs(testPlanConfigNodes []TestNode) (string, []string) {
	var warnings []string
	failureMessage := ""

	for _, testPlanConfigNode := range testPlanConfigNodes {
		msg, wrns := extractFailureMessage(testPlanConfigNode)
		if msg != "" {
			failureMessage += fmt.Sprintf("%s: %s\n", testPlanConfigNode.Name, msg)
		}
		if wrns != nil {
			warnings = append(warnings, wrns...)
		}
	}

	return failureMessage, warnings
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
	if lastChild.Type == TestNodeTypeTestPlanConfig {
		return extractFailureMessageFromTestPlanConfigs(testNode.Children)
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

	className := strings.Split(nodeIdentifier, "/")[0]
	if className == "" {
		// In rare cases the identifier is an empty string, so we need to use the test suite name which is the
		// same as the first part of the identifier in normal cases.
		className = fallbackClassName
	}

	message, failureMessageWarnings := extractFailureMessage(testNode)
	if len(failureMessageWarnings) > 0 {
		warnings = append(warnings, failureMessageWarnings...)
	}

	return TestCase{
		Name:      name,
		ClassName: className,
		Time:      extractDuration(testNode.Duration),
		Result:    testNode.Result,
		Message:   message,
	}, warnings
}
