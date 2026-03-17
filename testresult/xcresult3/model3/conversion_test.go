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
