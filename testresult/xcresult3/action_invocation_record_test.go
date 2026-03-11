package xcresult3

import (
	"testing"

	"github.com/bitrise-io/go-steputils/v2/testreport"
)

func TestTestFailureSummary_fileAndLineNumber(t *testing.T) {
	testCases := []struct {
		name     string
		summary  testFailureSummary
		wantFile string
		wantLine string
	}{
		{
			name: "",
			summary: testFailureSummary{
				DocumentLocationInCreatingWorkspace: documentLocationInCreatingWorkspace{
					URL: url{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
				},
			},
			wantFile: "file:/Xcode11TestUITests2.swift",
			wantLine: "CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotLine := tt.summary.fileAndLineNumber()
			if gotFile != tt.wantFile {
				t.Errorf("testFailureSummary.fileAndLineNumber() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if gotLine != tt.wantLine {
				t.Errorf("testFailureSummary.fileAndLineNumber() gotLine = %v, want %v", gotLine, tt.wantLine)
			}
		})
	}
}

func TestActionsInvocationRecord_failure(t *testing.T) {
	testCases := []struct {
		name   string
		record actionsInvocationRecord
		test   actionTestSummaryGroup
		want   string
	}{
		{
			name: "Simple test",
			record: actionsInvocationRecord{
				Issues: issues{
					TestFailureSummaries: testFailureSummaries{
						Values: []testFailureSummary{
							{
								ProducingTarget: producingTarget{Value: "Xcode11TestUITests2"},
								TestCaseName:    testCaseName{Value: "Xcode11TestUITests2.testFail()"},
								Message:         message{Value: "XCTAssertEqual failed: (\"1\") is not equal to (\"0\")"},
								DocumentLocationInCreatingWorkspace: documentLocationInCreatingWorkspace{
									URL: url{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
								},
							},
						},
					},
				},
			},
			test: actionTestSummaryGroup{
				Identifier: identifier{Value: "Xcode11TestUITests2/testFail()"},
			},
			want: `file:/Xcode11TestUITests2.swift:CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33 - XCTAssertEqual failed: ("1") is not equal to ("0")`,
		},
		{
			name: "class inherited test",
			record: actionsInvocationRecord{
				Issues: issues{
					TestFailureSummaries: testFailureSummaries{
						Values: []testFailureSummary{
							{
								ProducingTarget: producingTarget{Value: "Xcode11TestUITests2"},
								TestCaseName:    testCaseName{Value: "SomethingDifferentClass.testFail()"},
								Message:         message{Value: "XCTAssertEqual failed: (\"1\") is not equal to (\"0\")"},
								DocumentLocationInCreatingWorkspace: documentLocationInCreatingWorkspace{
									URL: url{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
								},
							},
						},
					},
				},
			},
			test: actionTestSummaryGroup{
				Identifier: identifier{Value: "SomethingDifferentClass/testFail()"},
			},
			want: `file:/Xcode11TestUITests2.swift:CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33 - XCTAssertEqual failed: ("1") is not equal to ("0")`,
		},
		{
			name: "inner class test",
			record: actionsInvocationRecord{
				Issues: issues{
					TestFailureSummaries: testFailureSummaries{
						Values: []testFailureSummary{
							{
								ProducingTarget: producingTarget{Value: "Xcode11TestUITests2"},
								TestCaseName:    testCaseName{Value: "-[SomethingDifferentClass testFail]"},
								Message:         message{Value: "XCTAssertEqual failed: (\"1\") is not equal to (\"0\")"},
								DocumentLocationInCreatingWorkspace: documentLocationInCreatingWorkspace{
									URL: url{Value: "file:/Xcode11TestUITests2.swift#CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33"},
								},
							},
						},
					},
				},
			},
			test: actionTestSummaryGroup{
				Identifier: identifier{Value: "SomethingDifferentClass/testFail"},
			},
			want: `file:/Xcode11TestUITests2.swift:CharacterRangeLen=0&EndingLineNumber=33&StartingLineNumber=33 - XCTAssertEqual failed: ("1") is not equal to ("0")`,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.failure(tt.test, testreport.TestSuite{Name: "Xcode11TestUITests2"}); got != tt.want {
				t.Errorf("actionsInvocationRecord.failure() = %v, want %v", got, tt.want)
			}
		})
	}
}
