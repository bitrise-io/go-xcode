package xcresult3

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bitrise-io/go-utils/pretty"
)

func TestActionTestPlanRunSummaries_tests(t *testing.T) {
	tests := []struct {
		name      string
		summaries ActionTestPlanRunSummaries
		want      map[string][]ActionTestSummaryGroup
	}{
		{
			name: "single test with status",
			summaries: ActionTestPlanRunSummaries{
				Summaries: Summaries{
					Values: []Summary{
						Summary{
							TestableSummaries: TestableSummaries{
								Values: []ActionTestableSummary{
									ActionTestableSummary{
										Name: Name{Value: "test case 1"},
										Tests: Tests{
											Values: []ActionTestSummaryGroup{
												ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: map[string][]ActionTestSummaryGroup{
				"test case 1": []ActionTestSummaryGroup{
					ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
				},
			},
		},
		{
			name: "single test with status + subtests with status",
			summaries: ActionTestPlanRunSummaries{
				Summaries: Summaries{
					Values: []Summary{
						Summary{
							TestableSummaries: TestableSummaries{
								Values: []ActionTestableSummary{
									ActionTestableSummary{
										Name: Name{Value: "test case 1"},
										Tests: Tests{
											Values: []ActionTestSummaryGroup{
												ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
											},
										},
									},
									ActionTestableSummary{
										Name: Name{Value: "test case 2"},
										Tests: Tests{
											Values: []ActionTestSummaryGroup{
												ActionTestSummaryGroup{
													Subtests: Subtests{
														Values: []ActionTestSummaryGroup{
															ActionTestSummaryGroup{
																TestStatus: TestStatus{Value: "success"},
															},
															ActionTestSummaryGroup{
																TestStatus: TestStatus{Value: "success"},
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
			},
			want: map[string][]ActionTestSummaryGroup{
				"test case 1": []ActionTestSummaryGroup{
					ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
				},
				"test case 2": []ActionTestSummaryGroup{
					ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
					ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
				},
			},
		},
		{
			name:      "no test with status",
			summaries: ActionTestPlanRunSummaries{},
			want:      map[string][]ActionTestSummaryGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := tt.summaries.tests(); !reflect.DeepEqual(got, tt.want) {
				fmt.Println("want: ", pretty.Object(tt.want))
				fmt.Println("got: ", pretty.Object(got))
				t.Errorf("ActionTestPlanRunSummaries.tests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActionTestPlanRunSummaries_failuresCount(t *testing.T) {
	tests := []struct {
		name                string
		summaries           ActionTestPlanRunSummaries
		testableSummaryName string
		wantFailure         int
	}{
		{
			name: "single failure",
			summaries: ActionTestPlanRunSummaries{
				Summaries: Summaries{
					Values: []Summary{
						Summary{
							TestableSummaries: TestableSummaries{
								Values: []ActionTestableSummary{
									ActionTestableSummary{
										Name: Name{Value: "test case"},
										Tests: Tests{
											Values: []ActionTestSummaryGroup{
												ActionTestSummaryGroup{TestStatus: TestStatus{Value: "Failure"}},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFailure := tt.summaries.failuresCount(tt.testableSummaryName); gotFailure != tt.wantFailure {
				t.Errorf("ActionTestPlanRunSummaries.failuresCount() = %v, want %v", gotFailure, tt.wantFailure)
			}
		})
	}
}

func TestActionTestPlanRunSummaries_totalTime(t *testing.T) {
	tests := []struct {
		name                string
		summaries           ActionTestPlanRunSummaries
		testableSummaryName string
		wantTime            float64
	}{
		{
			name: "single test",
			summaries: ActionTestPlanRunSummaries{
				Summaries: Summaries{
					Values: []Summary{
						Summary{
							TestableSummaries: TestableSummaries{
								Values: []ActionTestableSummary{
									ActionTestableSummary{
										Name: Name{Value: "test case"},
										Tests: Tests{
											Values: []ActionTestSummaryGroup{
												ActionTestSummaryGroup{
													Duration:   Duration{Value: "10"},
													TestStatus: TestStatus{Value: "Failure"},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTime := tt.summaries.totalTime(tt.testableSummaryName); gotTime != tt.wantTime {
				t.Errorf("ActionTestPlanRunSummaries.totalTime() = %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}
