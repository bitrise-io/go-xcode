package xcresult3

import (
	"reflect"
	"testing"
)

func TestActionTestSummaryGroup_references(t *testing.T) {
	tests := []struct {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ActionTestSummaryGroup{}
			g.Identifier.Value = tt.identifier

			gotClass, gotMethod := g.references()
			if gotClass != tt.wantClass {
				t.Errorf("ActionTestSummaryGroup.references() gotClass = %v, want %v", gotClass, tt.wantClass)
			}
			if gotMethod != tt.wantMethod {
				t.Errorf("ActionTestSummaryGroup.references() gotMethod = %v, want %v", gotMethod, tt.wantMethod)
			}
		})
	}
}

func TestActionTestSummaryGroup_testsWithStatus(t *testing.T) {

	tests := []struct {
		name       string
		group      ActionTestSummaryGroup
		subtests   []ActionTestSummaryGroup
		wantGroups []ActionTestSummaryGroup
	}{
		{
			name: "status in the root ActionTestSummaryGroup",
			group: ActionTestSummaryGroup{
				TestStatus: TestStatus{Value: "success"},
			},
			wantGroups: []ActionTestSummaryGroup{ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}}},
		},
		{
			name: "status in a sub ActionTestSummaryGroup",
			group: ActionTestSummaryGroup{
				Subtests: Subtests{
					Values: []ActionTestSummaryGroup{
						ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}},
					},
				},
			},
			wantGroups: []ActionTestSummaryGroup{ActionTestSummaryGroup{TestStatus: TestStatus{Value: "success"}}},
		},
		{
			name:       "no status",
			group:      ActionTestSummaryGroup{},
			wantGroups: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGroups := tt.group.testsWithStatus()
			if !reflect.DeepEqual(gotGroups, tt.wantGroups) {
				t.Errorf("ActionTestSummaryGroup.testsWithStatus() gotTarget = %v, want %v", gotGroups, tt.wantGroups)
			}
		})
	}
}
