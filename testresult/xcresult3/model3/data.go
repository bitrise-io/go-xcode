package model3

import "time"

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
	Time      time.Duration
	Result    TestResult
	Message   string
}
