package model3

// TestNodeType identifies the kind of a node in the xcresulttool test tree.
type TestNodeType string

// These are all the types the xcresulttool (version 23500, format version 3.53) supports.
const (
	TestNodeTypeTestPlan       TestNodeType = "Test Plan"
	TestNodeTypeUnitTestBundle TestNodeType = "Unit test bundle"
	TestNodeTypeUITestBundle   TestNodeType = "UI test bundle"
	TestNodeTypeTestSuite      TestNodeType = "Test Suite"
	TestNodeTypeTestCase       TestNodeType = "Test Case"
	TestNodeTypeDevice         TestNodeType = "Device"
	TestNodeTypeTestPlanConfig TestNodeType = "Test Plan Configuration"
	TestNodeTypeArguments      TestNodeType = "Arguments"
	TestNodeTypeRepetition     TestNodeType = "Repetition"
	TestNodeTypeTestCaseRun    TestNodeType = "Test Case Run"
	TestNodeTypeFailureMessage TestNodeType = "Failure Message"
	TestNodeTypeSourceCodeRef  TestNodeType = "Source Code Reference"
	TestNodeTypeAttachment     TestNodeType = "Attachment"
	TestNodeTypeExpression     TestNodeType = "Expression"
	TestNodeTypeTestValue      TestNodeType = "Test Value"
)

// TestResult represents the outcome of a test case.
type TestResult string

// TestResult values reported by xcresulttool.
const (
	TestResultPassed          TestResult = "Passed"
	TestResultFailed          TestResult = "Failed"
	TestResultSkipped         TestResult = "Skipped"
	TestResultExpectedFailure TestResult = "Expected Failure"
	TestResultUnknown         TestResult = "unknown"
)

// TestData is the top-level structure returned by xcresulttool for a test run.
type TestData struct {
	Devices                []Devices       `json:"devices"`
	TestNodes              []TestNode      `json:"testNodes"`
	TestPlanConfigurations []Configuration `json:"testPlanConfigurations"`
}

// Devices describes a device that participated in the test run.
type Devices struct {
	Identifier   string `json:"deviceId"`
	Name         string `json:"deviceName"`
	Architecture string `json:"architecture"`
	ModelName    string `json:"modelName"`
	Platform     string `json:"platform"`
	OS           string `json:"osVersion"`
}

// TestNode is a node in the xcresulttool test tree (plan, bundle, suite, case, etc.).
type TestNode struct {
	Identifier string       `json:"nodeIdentifier"`
	Type       TestNodeType `json:"nodeType"`
	Name       string       `json:"name"`
	Details    string       `json:"details"`
	Duration   string       `json:"duration"`
	Result     TestResult   `json:"result"`
	Tags       []string     `json:"tags"`
	Children   []TestNode   `json:"children"`
}

// Configuration describes a test plan configuration used during the test run.
type Configuration struct {
	Identifier string `json:"configurationId"`
	Name       string `json:"configurationName"`
}
