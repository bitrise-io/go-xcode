package xcresult3

import (
	"testing"

	"github.com/bitrise-io/go-steputils/v2/testreport"
)

func TestTestFailureSummary_fileAndLineNumber(t *testing.T) {
	tests := []struct {
		name     string
		summary  TestFailureSummary
		wantFile string
		wantLine string
	}{
		{
			name: "",
			summary: TestFailureSummary{
				DocumentLocationInCreatingWorkspace: DocumentLocationInCreatingWorkspace{
					URL: URL{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
				},
			},
			wantFile: "file:/Xcode11TestUITests2.swift",
			wantLine: "CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotLine := tt.summary.fileAndLineNumber()
			if gotFile != tt.wantFile {
				t.Errorf("TestFailureSummary.fileAndLineNumber() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if gotLine != tt.wantLine {
				t.Errorf("TestFailureSummary.fileAndLineNumber() gotLine = %v, want %v", gotLine, tt.wantLine)
			}
		})
	}
}

func TestActionsInvocationRecord_failure(t *testing.T) {
	tests := []struct {
		name   string
		record ActionsInvocationRecord
		test   ActionTestSummaryGroup
		want   string
	}{
		{
			name: "Simple test",
			record: ActionsInvocationRecord{
				Issues: Issues{
					TestFailureSummaries: TestFailureSummaries{
						Values: []TestFailureSummary{
							TestFailureSummary{
								ProducingTarget: ProducingTarget{Value: "Xcode11TestUITests2"},
								TestCaseName:    TestCaseName{Value: "Xcode11TestUITests2.testFail()"},
								Message:         Message{Value: "XCTAssertEqual failed: (\"1\") is not equal to (\"0\")"},
								DocumentLocationInCreatingWorkspace: DocumentLocationInCreatingWorkspace{
									URL: URL{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
								},
							},
						},
					},
				},
			},
			test: ActionTestSummaryGroup{
				Identifier: Identifier{Value: "Xcode11TestUITests2/testFail()"},
			},
			want: `file:/Xcode11TestUITests2.swift:CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33 - XCTAssertEqual failed: ("1") is not equal to ("0")`,
		},
		{
			name: "class inherited test",
			record: ActionsInvocationRecord{
				Issues: Issues{
					TestFailureSummaries: TestFailureSummaries{
						Values: []TestFailureSummary{
							TestFailureSummary{
								ProducingTarget: ProducingTarget{Value: "Xcode11TestUITests2"},
								TestCaseName:    TestCaseName{Value: "SomethingDifferentClass.testFail()"},
								Message:         Message{Value: "XCTAssertEqual failed: (\"1\") is not equal to (\"0\")"},
								DocumentLocationInCreatingWorkspace: DocumentLocationInCreatingWorkspace{
									URL: URL{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
								},
							},
						},
					},
				},
			},
			test: ActionTestSummaryGroup{
				Identifier: Identifier{Value: "SomethingDifferentClass/testFail()"},
			},
			want: `file:/Xcode11TestUITests2.swift:CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33 - XCTAssertEqual failed: ("1") is not equal to ("0")`,
		},
		{
			name: "inner class test",
			record: ActionsInvocationRecord{
				Issues: Issues{
					TestFailureSummaries: TestFailureSummaries{
						Values: []TestFailureSummary{
							TestFailureSummary{
								ProducingTarget: ProducingTarget{Value: "Xcode11TestUITests2"},
								TestCaseName:    TestCaseName{Value: "-[SomethingDifferentClass testFail]"},
								Message:         Message{Value: "XCTAssertEqual failed: (\"1\") is not equal to (\"0\")"},
								DocumentLocationInCreatingWorkspace: DocumentLocationInCreatingWorkspace{
									URL: URL{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
								},
							},
						},
					},
				},
			},
			test: ActionTestSummaryGroup{
				Identifier: Identifier{Value: "SomethingDifferentClass/testFail"},
			},
			want: `file:/Xcode11TestUITests2.swift:CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33 - XCTAssertEqual failed: ("1") is not equal to ("0")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.failure(tt.test, testreport.TestSuite{Name: "Xcode11TestUITests2"}); got != tt.want {
				t.Errorf("ActionsInvocationRecord.failure() = %v, want %v", got, tt.want)
			}
		})
	}
}
