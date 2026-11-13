package model3

import "time"

// TestSummary holds the top-level test results grouped by test plans.
type TestSummary struct {
	TestPlans []TestPlan
}

type TestPlan struct {
	Name        string
	TestBundles []TestBundle
}

type TestBundle struct {
	Name       string
	TestSuites []TestSuite
}

type TestSuite struct {
	Name      string
	TestCases []TestCaseWithRetries
}

type TestCaseWithRetries struct {
	TestCase
	Retries []TestCase
}

type TestCase struct {
	Name      string
	ClassName string
	// Identifier is the test framework's own identifier of the test case, which Name does not
	// carry when the test framework reports a display name instead.
	Identifier string
	Time       time.Duration
	Result     TestResult
	Message    string
}
