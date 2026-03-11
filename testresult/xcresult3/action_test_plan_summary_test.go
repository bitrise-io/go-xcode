package xcresult3

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/pretty"
)

func TestActionTestPlanRunSummaries_tests(t *testing.T) {
	testCases := []struct {
		name      string
		summaries actionTestPlanRunSummaries
		want      map[string][]actionTestSummaryGroup
	}{
		{
			name: "single test with status",
			summaries: actionTestPlanRunSummaries{
				Summaries: summaries{
					Values: []summary{
						{
							TestableSummaries: testableSummaries{
								Values: []actionTestableSummary{
									{
										Name: name{Value: "test case 1"},
										Tests: tests{
											Values: []actionTestSummaryGroup{
												{TestStatus: testStatus{Value: "success"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: map[string][]actionTestSummaryGroup{
				"test case 1": {
					{TestStatus: testStatus{Value: "success"}},
				},
			},
		},
		{
			name: "single test with status + subtests with status",
			summaries: actionTestPlanRunSummaries{
				Summaries: summaries{
					Values: []summary{
						{
							TestableSummaries: testableSummaries{
								Values: []actionTestableSummary{
									{
										Name: name{Value: "test case 1"},
										Tests: tests{
											Values: []actionTestSummaryGroup{
												{TestStatus: testStatus{Value: "success"}},
											},
										},
									},
									{
										Name: name{Value: "test case 2"},
										Tests: tests{
											Values: []actionTestSummaryGroup{
												{
													Subtests: subtests{
														Values: []actionTestSummaryGroup{
															{TestStatus: testStatus{Value: "success"}},
															{TestStatus: testStatus{Value: "success"}},
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
			},
			want: map[string][]actionTestSummaryGroup{
				"test case 1": {
					{TestStatus: testStatus{Value: "success"}},
				},
				"test case 2": {
					{TestStatus: testStatus{Value: "success"}},
					{TestStatus: testStatus{Value: "success"}},
				},
			},
		},
		{
			name:      "no test with status",
			summaries: actionTestPlanRunSummaries{},
			want:      map[string][]actionTestSummaryGroup{},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := tt.summaries.tests(); !reflect.DeepEqual(got, tt.want) {
				fmt.Println("want: ", pretty.Object(tt.want))
				fmt.Println("got: ", pretty.Object(got))
				t.Errorf("actionTestPlanRunSummaries.tests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActionTestPlanRunSummaries_failuresCount(t *testing.T) {
	testCases := []struct {
		name                string
		summaries           actionTestPlanRunSummaries
		testableSummaryName string
		wantFailure         int
	}{
		{
			name: "single failure",
			summaries: actionTestPlanRunSummaries{
				Summaries: summaries{
					Values: []summary{
						{
							TestableSummaries: testableSummaries{
								Values: []actionTestableSummary{
									{
										Name: name{Value: "test case"},
										Tests: tests{
											Values: []actionTestSummaryGroup{
												{TestStatus: testStatus{Value: "Failure"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			testableSummaryName: "test case",
			wantFailure:         1,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if gotFailure := tt.summaries.failuresCount(tt.testableSummaryName); gotFailure != tt.wantFailure {
				t.Errorf("actionTestPlanRunSummaries.failuresCount() = %v, want %v", gotFailure, tt.wantFailure)
			}
		})
	}
}

func TestActionTestPlanRunSummaries_totalTime(t *testing.T) {
	testCases := []struct {
		name                string
		summaries           actionTestPlanRunSummaries
		testableSummaryName string
		wantTime            float64
	}{
		{
			name: "single test",
			summaries: actionTestPlanRunSummaries{
				Summaries: summaries{
					Values: []summary{
						{
							TestableSummaries: testableSummaries{
								Values: []actionTestableSummary{
									{
										Name: name{Value: "test case"},
										Tests: tests{
											Values: []actionTestSummaryGroup{
												{
													Duration:   duration{Value: "10"},
													TestStatus: testStatus{Value: "Failure"},
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
			testableSummaryName: "test case",
			wantTime:            10,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if gotTime := tt.summaries.totalTime(tt.testableSummaryName); gotTime != tt.wantTime {
				t.Errorf("actionTestPlanRunSummaries.totalTime() = %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}
