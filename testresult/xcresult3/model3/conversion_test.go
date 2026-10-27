package model3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConversion(t *testing.T) {
	tests := []struct {
		name string
		data *TestData
		want *TestSummary
	}{
		{
			name: "Simple case with multiple test bundles, suites, and cases",
			data: &TestData{
				TestNodes: []TestNode{
					{
						Type: TestNodeTypeTestPlan,
						Name: "TP1",
						Children: []TestNode{
							{
								Type: TestNodeTypeUnitTestBundle,
								Name: "TB1",
								Children: []TestNode{
									{
										Type: TestNodeTypeTestSuite,
										Name: "TS1",
										Children: []TestNode{
											{
												Type:     TestNodeTypeTestCase,
												Name:     "TC1",
												Result:   TestResultPassed,
												Duration: "0.5s",
											},
											{
												Type:     TestNodeTypeTestCase,
												Name:     "TC2",
												Result:   TestResultFailed,
												Duration: "1m 11s",
											},
										},
									},
									{
										Type: TestNodeTypeTestSuite,
										Name: "TS2",
										Children: []TestNode{
											{
												Type:     TestNodeTypeTestCase,
												Name:     "TC3",
												Result:   TestResultSkipped,
												Duration: "66s",
											},
											{
												Type:     TestNodeTypeTestCase,
												Name:     "TC4",
												Result:   TestResultPassed,
												Duration: "0.03s",
											},
										},
									},
								},
							},
							{
								Type: TestNodeTypeUITestBundle,
								Name: "TB2",
								Children: []TestNode{
									{
										Type: TestNodeTypeTestSuite,
										Name: "TS3",
										Children: []TestNode{
											{
												Type:     TestNodeTypeTestCase,
												Name:     "TC5",
												Result:   TestResultPassed,
												Duration: "15s",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: &TestSummary{
				TestPlans: []TestPlan{
					{
						Name: "TP1",
						TestBundles: []TestBundle{
							{
								Name: "TB1",
								TestSuites: []TestSuite{
									{
										Name: "TS1",
										TestCases: []TestCaseWithRetries{
											{
												TestCase: TestCase{
													Name:      "TC1",
													ClassName: "TS1",
													Time:      500 * time.Millisecond,
													Result:    "Passed",
												},
											},
											{
												TestCase: TestCase{
													Name:      "TC2",
													ClassName: "TS1",
													Time:      71 * time.Second,
													Result:    "Failed",
												},
											},
										},
									},
									{
										Name: "TS2",
										TestCases: []TestCaseWithRetries{
											{
												TestCase: TestCase{
													Name:      "TC3",
													ClassName: "TS2",
													Time:      66 * time.Second,
													Result:    "Skipped",
												},
											},
											{
												TestCase: TestCase{
													Name:      "TC4",
													ClassName: "TS2",
													Time:      30 * time.Millisecond,
													Result:    "Passed",
												},
											},
										},
									},
								},
							},
							{
								Name: "TB2",
								TestSuites: []TestSuite{
									{
										Name: "TS3",
										TestCases: []TestCaseWithRetries{
											{
												TestCase: TestCase{
													Name:      "TC5",
													ClassName: "TS3",
													Time:      15 * time.Second,
													Result:    "Passed",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Test cases with node identifiers, one of them reported with a display name",
			data: &TestData{
				TestNodes: []TestNode{
					{
						Type: TestNodeTypeTestPlan,
						Name: "TP1",
						Children: []TestNode{
							{
								Type: TestNodeTypeUnitTestBundle,
								Name: "BullsEyeTests",
								Children: []TestNode{
									{
										Type: TestNodeTypeTestSuite,
										Name: "BullsEyeSwiftTestingTests",
										Children: []TestNode{
											{
												Type:       TestNodeTypeTestCase,
												Name:       "Score is computed when the guess matches the target",
												Identifier: "BullsEyeSwiftTestingTests/scoreIsComputedWhenGuessMatchesTarget()",
												Result:     TestResultPassed,
												Duration:   "0.1s",
											},
											{
												Type:       TestNodeTypeTestCase,
												Name:       "scoreIsComputedWhenGuessIsHigherThanTarget()",
												Identifier: "BullsEyeSwiftTestingTests/scoreIsComputedWhenGuessIsHigherThanTarget()",
												Result:     TestResultPassed,
												Duration:   "0.2s",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: &TestSummary{
				TestPlans: []TestPlan{
					{
						Name: "TP1",
						TestBundles: []TestBundle{
							{
								Name: "BullsEyeTests",
								TestSuites: []TestSuite{
									{
										Name: "BullsEyeSwiftTestingTests",
										TestCases: []TestCaseWithRetries{
											{
												TestCase: TestCase{
													Name:       "Score is computed when the guess matches the target",
													ClassName:  "BullsEyeSwiftTestingTests",
													Identifier: "BullsEyeSwiftTestingTests/scoreIsComputedWhenGuessMatchesTarget()",
													Time:       100 * time.Millisecond,
													Result:     "Passed",
												},
											},
											{
												TestCase: TestCase{
													Name:       "scoreIsComputedWhenGuessIsHigherThanTarget()",
													ClassName:  "BullsEyeSwiftTestingTests",
													Identifier: "BullsEyeSwiftTestingTests/scoreIsComputedWhenGuessIsHigherThanTarget()",
													Time:       200 * time.Millisecond,
													Result:     "Passed",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Test bundle without test suites",
			data: &TestData{
				TestNodes: []TestNode{
					{
						Type: TestNodeTypeTestPlan,
						Name: "TP1",
						Children: []TestNode{
							{
								Type: TestNodeTypeUnitTestBundle,
								Name: "TB1",
								Children: []TestNode{
									{
										Type:     TestNodeTypeTestCase,
										Name:     "TC1",
										Result:   TestResultPassed,
										Duration: "0.5s",
									},
								},
							},
						},
					},
				},
			},
			want: &TestSummary{
				TestPlans: []TestPlan{
					{
						Name: "TP1",
						TestBundles: []TestBundle{
							{
								Name: "TB1",
								TestSuites: []TestSuite{
									{
										Name: "TB1",
										TestCases: []TestCaseWithRetries{
											{
												TestCase: TestCase{
													Name:      "TC1",
													ClassName: "TB1",
													Time:      500 * time.Millisecond,
													Result:    "Passed",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _, err := Convert(test.data)
			require.NoError(t, err)

			require.Equal(t, test.want, got)
		})
	}
}

func TestConversion_repetitions(t *testing.T) {
	const (
		identifier = "ProductTests/measurements(asset:)"
		name       = "measurements(asset:)"
		className  = "ProductTests"
	)

	tests := []struct {
		name     string
		testCase TestNode
		want     TestCaseWithRetries
	}{
		{
			name: "retried test case without arguments",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultPassed, Duration: "3s",
				Children: []TestNode{
					repetition("First Run", TestResultFailed, "1s", "boom"),
					repetition("Retry 1", TestResultPassed, "2s", ""),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultPassed, 3*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultFailed, time.Second, "boom"),
					testCase(TestResultPassed, 2*time.Second, ""),
				},
			},
		},
		{
			name: "parametrized test passes on retry",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "5s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "",
						repetition("First Run", TestResultFailed, "1s", "shirts failed"),
						repetition("Retry 1", TestResultPassed, "2s", ""),
					),
					arguments(".dresses", TestResultPassed, "",
						repetition("First Run", TestResultPassed, "3s", ""),
					),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultPassed, 5*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultFailed, 3*time.Second, ".shirts: shirts failed\n"),
					testCase(TestResultPassed, 2*time.Second, ""),
				},
			},
		},
		{
			name: "parametrized test fails on every retry",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "3s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "",
						repetition("First Run", TestResultFailed, "1s", "first"),
						repetition("Retry 1", TestResultFailed, "1s", "second"),
					),
					arguments(".dresses", TestResultPassed, "",
						repetition("First Run", TestResultPassed, "1s", ""),
					),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultFailed, 3*time.Second, ".shirts: second\n"),
				Retries: []TestCase{
					testCase(TestResultFailed, time.Second, ".shirts: first\n"),
					testCase(TestResultFailed, time.Second, ".shirts: second\n"),
				},
			},
		},
		{
			name: "parametrized test run a fixed number of times fails when any run fails",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "2s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "",
						repetition("Repetition 1", TestResultFailed, "1s", "boom"),
						repetition("Repetition 2", TestResultPassed, "1s", ""),
					),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultFailed, 2*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultFailed, time.Second, ".shirts: boom\n"),
					testCase(TestResultPassed, time.Second, ""),
				},
			},
		},
		{
			name: "parametrized test without repetitions keeps its own result",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "2s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "boom"),
					arguments(".dresses", TestResultPassed, ""),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultFailed, 2*time.Second, ".shirts: boom\n"),
			},
		},
		{
			name: "argument without repetitions joins the first attempt",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "6s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "",
						repetition("First Run", TestResultFailed, "1s", "boom"),
						repetition("Retry 1", TestResultPassed, "1s", ""),
					),
					{Type: TestNodeTypeArguments, Name: ".dresses", Result: TestResultPassed, Duration: "4s"},
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultPassed, 6*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultFailed, 4*time.Second, ".shirts: boom\n"),
					testCase(TestResultPassed, time.Second, ""),
				},
			},
		},
		{
			name: "failed argument without repetitions fails the retried test",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "6s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "",
						repetition("First Run", TestResultFailed, "1s", "shirts failed"),
						repetition("Retry 1", TestResultPassed, "1s", ""),
					),
					{
						Type: TestNodeTypeArguments, Name: ".dresses", Result: TestResultFailed, Duration: "4s",
						Children: []TestNode{{Type: TestNodeTypeFailureMessage, Name: "dresses failed"}},
					},
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultFailed, 6*time.Second, ".dresses: dresses failed\n"),
				Retries: []TestCase{
					testCase(TestResultFailed, 4*time.Second, ".shirts: shirts failed\n.dresses: dresses failed\n"),
					testCase(TestResultFailed, time.Second, ".dresses: dresses failed\n"),
				},
			},
		},
		{
			name: "arguments retried a different number of times",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultFailed, Duration: "9s",
				Children: []TestNode{
					arguments(".shirts", TestResultFailed, "",
						repetition("First Run", TestResultFailed, "1s", "shirts first"),
						repetition("Retry 1", TestResultFailed, "1s", "shirts second"),
						repetition("Retry 2", TestResultPassed, "1s", ""),
					),
					arguments(".dresses", TestResultFailed, "",
						repetition("First Run", TestResultFailed, "2s", "dresses first"),
						repetition("Retry 1", TestResultPassed, "2s", ""),
					),
					arguments(".shoes", TestResultPassed, "",
						repetition("First Run", TestResultPassed, "3s", ""),
					),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultPassed, 9*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultFailed, 3*time.Second, ".shirts: shirts first\n.dresses: dresses first\n"),
					testCase(TestResultFailed, 2*time.Second, ".shirts: shirts second\n"),
					testCase(TestResultPassed, time.Second, ""),
				},
			},
		},
		{
			name: "attempt with passed and skipped arguments is passed",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultPassed, Duration: "2s",
				Children: []TestNode{
					arguments(".shirts", TestResultPassed, "",
						repetition("Repetition 1", TestResultPassed, "1s", ""),
					),
					arguments(".dresses", TestResultSkipped, "",
						repetition("Repetition 1", TestResultSkipped, "1s", ""),
					),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultPassed, 2*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultPassed, time.Second, ""),
				},
			},
		},
		{
			name: "attempt with only skipped arguments is skipped",
			testCase: TestNode{
				Type: TestNodeTypeTestCase, Identifier: identifier, Name: name, Result: TestResultSkipped, Duration: "2s",
				Children: []TestNode{
					arguments(".shirts", TestResultSkipped, "",
						repetition("Repetition 1", TestResultSkipped, "1s", ""),
					),
					arguments(".dresses", TestResultSkipped, "",
						repetition("Repetition 1", TestResultSkipped, "1s", ""),
					),
				},
			},
			want: TestCaseWithRetries{
				TestCase: testCase(TestResultSkipped, 2*time.Second, ""),
				Retries: []TestCase{
					testCase(TestResultSkipped, time.Second, ""),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := &TestData{
				TestNodes: []TestNode{
					{
						Type: TestNodeTypeTestPlan,
						Name: "TP1",
						Children: []TestNode{
							{
								Type: TestNodeTypeUnitTestBundle,
								Name: "TB1",
								Children: []TestNode{
									{Type: TestNodeTypeTestSuite, Name: className, Children: []TestNode{test.testCase}},
								},
							},
						},
					},
				},
			}

			got, _, err := Convert(data)
			require.NoError(t, err)

			require.Equal(t, []TestCaseWithRetries{test.want}, got.TestPlans[0].TestBundles[0].TestSuites[0].TestCases)
		})
	}
}

func testCase(result TestResult, duration time.Duration, message string) TestCase {
	return TestCase{
		Name:       "measurements(asset:)",
		ClassName:  "ProductTests",
		Identifier: "ProductTests/measurements(asset:)",
		Time:       duration,
		Result:     result,
		Message:    message,
	}
}

func arguments(name string, result TestResult, failureMessage string, repetitions ...TestNode) TestNode {
	node := TestNode{Type: TestNodeTypeArguments, Name: name, Result: result, Children: repetitions}
	if failureMessage != "" {
		node.Children = append(node.Children, TestNode{Type: TestNodeTypeFailureMessage, Name: failureMessage})
	}

	return node
}

func repetition(name string, result TestResult, duration, failureMessage string) TestNode {
	node := TestNode{Type: TestNodeTypeRepetition, Name: name, Result: result, Duration: duration}
	if failureMessage != "" {
		node.Children = []TestNode{{Type: TestNodeTypeFailureMessage, Name: failureMessage}}
	}

	return node
}
