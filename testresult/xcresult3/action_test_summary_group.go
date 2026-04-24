package xcresult3

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var errSummaryNotFound = errors.New("no summaryRef.ID.Value found for test case")

type actionTestSummaryGroup struct {
	Name       name       `json:"name"`
	Identifier identifier `json:"identifier"`
	Duration   duration   `json:"duration"`
	TestStatus testStatus `json:"testStatus"` // only the inner-most tests will have a status, the ones which don't have "subtests"
	SummaryRef summaryRef `json:"summaryRef"` // only the inner-most tests will have a summaryRef, the ones which don't have "subtests"
	Subtests   subtests   `json:"subtests"`
}

type subtests struct {
	Values []actionTestSummaryGroup `json:"_values"`
}

type id struct {
	Value string `json:"_value"`
}

type summaryRef struct {
	ID id `json:"id"`
}

type testStatus struct {
	Value string `json:"_value"`
}

type duration struct {
	Value string `json:"_value"`
}

type identifier struct {
	Value string `json:"_value"`
}

func (g actionTestSummaryGroup) references() (class, method string) {
	// Xcode11TestUITests2/testFail()
	if g.Identifier.Value != "" {
		s := strings.Split(g.Identifier.Value, "/")
		if len(s) == 2 {
			return s[0], s[1]
		}
	}
	return
}

// testsWithStatus returns actionTestSummaryGroup entries with TestStatus set.
func (g actionTestSummaryGroup) testsWithStatus() (result []actionTestSummaryGroup) {
	if g.TestStatus.Value != "" {
		result = append(result, g)
	}

	for _, subtest := range g.Subtests.Values {
		result = append(result, subtest.testsWithStatus()...)
	}
	return
}

func (g actionTestSummaryGroup) loadActionTestSummary(xcresultPath string, useLegacyFlag bool) (actionTestSummary, error) {
	if g.SummaryRef.ID.Value == "" {
		return actionTestSummary{}, errSummaryNotFound
	}

	var summary actionTestSummary
	if err := xcresulttoolGet(xcresultPath, g.SummaryRef.ID.Value, useLegacyFlag, &summary); err != nil {
		return actionTestSummary{}, fmt.Errorf("failed to load action test summary: %w", err)
	}
	return summary, nil
}

func (g actionTestSummaryGroup) exportScreenshots(resultPth, outputDir string, useLegacyFlag bool) error {
	if g.TestStatus.Value == "" {
		return nil
	}

	if g.SummaryRef.ID.Value == "" {
		return nil
	}

	var summary actionTestSummary
	if err := xcresulttoolGet(resultPth, g.SummaryRef.ID.Value, useLegacyFlag, &summary); err != nil {
		return err
	}

	exported := map[string]bool{}
	for _, summary := range summary.ActivitySummaries.Values {
		for _, value := range summary.Attachments.Values {
			if value.Filename.Value != "" && value.PayloadRef.ID.Value != "" {
				if exported[value.PayloadRef.ID.Value] {
					continue
				}

				pth := filepath.Join(outputDir, value.Filename.Value)
				if err := xcresulttoolExport(resultPth, value.PayloadRef.ID.Value, pth, useLegacyFlag); err != nil {
					return err
				}
				exported[value.PayloadRef.ID.Value] = true
			}
		}
	}

	return nil
}
