package model3

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

type TestResult string

const (
	TestResultPassed          TestResult = "Passed"
	TestResultFailed          TestResult = "Failed"
	TestResultSkipped         TestResult = "Skipped"
	TestResultExpectedFailure TestResult = "Expected Failure"
	TestResultUnknown         TestResult = "unknown"
)

type TestData struct {
	Devices                []Devices       `json:"devices"`
	TestNodes              []TestNode      `json:"testNodes"`
	TestPlanConfigurations []Configuration `json:"testPlanConfigurations"`
}

type Devices struct {
	Identifier   string `json:"deviceId"`
	Name         string `json:"deviceName"`
	Architecture string `json:"architecture"`
	ModelName    string `json:"modelName"`
	Platform     string `json:"platform"`
	OS           string `json:"osVersion"`
}

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

type Configuration struct {
	Identifier string `json:"configurationId"`
	Name       string `json:"configurationName"`
}
