package xcresult3

import "strconv"

// ActionTestPlanRunSummaries ...
type ActionTestPlanRunSummaries struct {
	Summaries Summaries `json:"summaries"`
}

// Summaries ...
type Summaries struct {
	Values []Summary `json:"_values"`
}

// Summary ...
type Summary struct {
	TestableSummaries TestableSummaries `json:"testableSummaries"`
}

// TestableSummaries ...
type TestableSummaries struct {
	Values []ActionTestableSummary `json:"_values"`
}

// ActionTestableSummary ...
type ActionTestableSummary struct {
	Name  Name  `json:"name"`
	Tests Tests `json:"tests"`
}

// Tests ...
type Tests struct {
	Values []ActionTestSummaryGroup `json:"_values"`
}

// Name ...
type Name struct {
	Value string `json:"_value"`
}

// tests returns ActionTestSummaryGroup mapped by the container TestableSummary name.
func (s ActionTestPlanRunSummaries) tests() ([]string, map[string][]ActionTestSummaryGroup) {
	summaryGroupsByName := map[string][]ActionTestSummaryGroup{}

	var testSuiteOrder []string
	for _, summary := range s.Summaries.Values {
		for _, testableSummary := range summary.TestableSummaries.Values {
			// test suite
			name := testableSummary.Name.Value
			if _, found := summaryGroupsByName[name]; !found {
				testSuiteOrder = append(testSuiteOrder, name)
			}

			var tests []ActionTestSummaryGroup
			for _, test := range testableSummary.Tests.Values {
				tests = append(tests, test.testsWithStatus()...)
			}

			summaryGroupsByName[name] = tests
		}
	}

	return testSuiteOrder, summaryGroupsByName
}

func (s ActionTestPlanRunSummaries) failuresCount(testableSummaryName string) (failure int) {
	_, testsByCase := s.tests()
	tests := testsByCase[testableSummaryName]
	for _, test := range tests {
		if test.TestStatus.Value == "Failure" {
			failure++
		}
	}
	return
}

func (s ActionTestPlanRunSummaries) skippedCount(testableSummaryName string) (skipped int) {
	_, testsByCase := s.tests()
	tests := testsByCase[testableSummaryName]
	for _, test := range tests {
		if test.TestStatus.Value == "Skipped" {
			skipped++
		}
	}
	return
}

func (s ActionTestPlanRunSummaries) totalTime(testableSummaryName string) (time float64) {
	_, testsByCase := s.tests()
	tests := testsByCase[testableSummaryName]
	for _, test := range tests {
		if test.Duration.Value != "" {
			d, err := strconv.ParseFloat(test.Duration.Value, 64)
			if err == nil {
				time += d
			}
		}
	}
	return
}
