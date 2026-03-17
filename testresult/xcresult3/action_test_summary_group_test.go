package xcresult3

import (
	"reflect"
	"testing"
)

func TestActionTestSummaryGroup_references(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		wantClass  string
		wantMethod string
	}{
		{
			name:       "simple test",
			identifier: "Xcode11TestUITests2/testFail()",
			wantClass:  "Xcode11TestUITests2",
			wantMethod: "testFail()",
		},
		{
			name:       "invalid format",
			identifier: "Xcode11TestUITests2testFail()",
			wantMethod: "",
			wantClass:  "",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			g := actionTestSummaryGroup{}
			g.Identifier.Value = tt.identifier

			gotClass, gotMethod := g.references()
			if gotClass != tt.wantClass {
				t.Errorf("actionTestSummaryGroup.references() gotClass = %v, want %v", gotClass, tt.wantClass)
			}
			if gotMethod != tt.wantMethod {
				t.Errorf("actionTestSummaryGroup.references() gotMethod = %v, want %v", gotMethod, tt.wantMethod)
			}
		})
	}
}

func TestActionTestSummaryGroup_testsWithStatus(t *testing.T) {
	testCases := []struct {
		name       string
		group      actionTestSummaryGroup
		wantGroups []actionTestSummaryGroup
	}{
		{
			name: "status in the root actionTestSummaryGroup",
			group: actionTestSummaryGroup{
				TestStatus: testStatus{Value: "success"},
			},
			wantGroups: []actionTestSummaryGroup{{TestStatus: testStatus{Value: "success"}}},
		},
		{
			name: "status in a sub actionTestSummaryGroup",
			group: actionTestSummaryGroup{
				Subtests: subtests{
					Values: []actionTestSummaryGroup{
						{TestStatus: testStatus{Value: "success"}},
					},
				},
			},
			wantGroups: []actionTestSummaryGroup{{TestStatus: testStatus{Value: "success"}}},
		},
		{
			name:       "no status",
			group:      actionTestSummaryGroup{},
			wantGroups: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotGroups := tt.group.testsWithStatus()
			if !reflect.DeepEqual(gotGroups, tt.wantGroups) {
				t.Errorf("actionTestSummaryGroup.testsWithStatus() gotTarget = %v, want %v", gotGroups, tt.wantGroups)
			}
		})
	}
}
