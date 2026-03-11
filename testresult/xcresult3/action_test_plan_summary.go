package xcresult3

import "strconv"

type actionTestPlanRunSummaries struct {
	Summaries summaries `json:"summaries"`
}

type summaries struct {
	Values []summary `json:"_values"`
}

type summary struct {
	TestableSummaries testableSummaries `json:"testableSummaries"`
}

type testableSummaries struct {
	Values []actionTestableSummary `json:"_values"`
}

type actionTestableSummary struct {
	Name  name  `json:"name"`
	Tests tests `json:"tests"`
}

type tests struct {
	Values []actionTestSummaryGroup `json:"_values"`
}

type name struct {
	Value string `json:"_value"`
}

// tests returns actionTestSummaryGroup mapped by the container testableSummary name.
func (s actionTestPlanRunSummaries) tests() ([]string, map[string][]actionTestSummaryGroup) {
	summaryGroupsByName := map[string][]actionTestSummaryGroup{}

	var testSuiteOrder []string
	for _, smry := range s.Summaries.Values {
		for _, testableSummary := range smry.TestableSummaries.Values {
			// test suite
			n := testableSummary.Name.Value
			if _, found := summaryGroupsByName[n]; !found {
				testSuiteOrder = append(testSuiteOrder, n)
			}

			var ts []actionTestSummaryGroup
			for _, test := range testableSummary.Tests.Values {
				ts = append(ts, test.testsWithStatus()...)
			}

			summaryGroupsByName[n] = ts
		}
	}

	return testSuiteOrder, summaryGroupsByName
}

func (s actionTestPlanRunSummaries) failuresCount(testableSummaryName string) (failure int) {
	_, testsByCase := s.tests()
	ts := testsByCase[testableSummaryName]
	for _, test := range ts {
		if test.TestStatus.Value == "Failure" {
			failure++
		}
	}
	return
}

func (s actionTestPlanRunSummaries) skippedCount(testableSummaryName string) (skipped int) {
	_, testsByCase := s.tests()
	ts := testsByCase[testableSummaryName]
	for _, test := range ts {
		if test.TestStatus.Value == "Skipped" {
			skipped++
		}
	}
	return
}

func (s actionTestPlanRunSummaries) totalTime(testableSummaryName string) (time float64) {
	_, testsByCase := s.tests()
	ts := testsByCase[testableSummaryName]
	for _, test := range ts {
		if test.Duration.Value != "" {
			d, err := strconv.ParseFloat(test.Duration.Value, 64)
			if err == nil {
				time += d
			}
		}
	}
	return
}
