package model3

import "time"

// TestSummary holds the top-level test results grouped by test plans.
type TestSummary struct {
	TestPlans []TestPlan
}

// TestPlan represents a test plan and its test bundles.
type TestPlan struct {
	Name        string
	TestBundles []TestBundle
}

// TestBundle represents a test bundle and its test suites.
type TestBundle struct {
	Name       string
	TestSuites []TestSuite
}

// TestSuite represents a test suite and its test cases.
type TestSuite struct {
	Name      string
	TestCases []TestCaseWithRetries
}

// TestCaseWithRetries holds a test case along with any retry attempts.
type TestCaseWithRetries struct {
	TestCase
	Retries []TestCase
}

// TestCase represents a single test case execution result.
type TestCase struct {
	Name      string
	ClassName string
	Time      time.Duration
	Result    TestResult
	Message   string
}
