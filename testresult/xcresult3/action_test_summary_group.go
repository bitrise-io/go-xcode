package xcresult3

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ErrSummaryNotFound ...
var ErrSummaryNotFound = errors.New("no summaryRef.ID.Value found for test case")

// ActionTestSummaryGroup ...
type ActionTestSummaryGroup struct {
	Name       Name       `json:"name"`
	Identifier Identifier `json:"identifier"`
	Duration   Duration   `json:"duration"`
	TestStatus TestStatus `json:"testStatus"` // only the inner-most tests will have a status, the ones which don't have "subtests"
	SummaryRef SummaryRef `json:"summaryRef"` // only the inner-most tests will have a summaryRef, the ones which don't have "subtests"
	Subtests   Subtests   `json:"subtests"`
}

// Subtests ...
type Subtests struct {
	Values []ActionTestSummaryGroup `json:"_values"`
}

// ID ...
type ID struct {
	Value string `json:"_value"`
}

// SummaryRef ...
type SummaryRef struct {
	ID ID `json:"id"`
}

// TestStatus ...
type TestStatus struct {
	Value string `json:"_value"`
}

// Duration ...
type Duration struct {
	Value string `json:"_value"`
}

// Identifier ...
type Identifier struct {
	Value string `json:"_value"`
}

func (g ActionTestSummaryGroup) references() (class, method string) {
	// Xcode11TestUITests2/testFail()
	if g.Identifier.Value != "" {
		s := strings.Split(g.Identifier.Value, "/")
		if len(s) == 2 {
			return s[0], s[1]
		}
	}
	return
}

// testsWithStatus returns ActionTestSummaryGroup with TestStatus defined.
func (g ActionTestSummaryGroup) testsWithStatus() (tests []ActionTestSummaryGroup) {
	if g.TestStatus.Value != "" {
		tests = append(tests, g)
	}

	for _, subtest := range g.Subtests.Values {
		tests = append(tests, subtest.testsWithStatus()...)
	}
	return
}

// loadActionTestSummary ...
func (g ActionTestSummaryGroup) loadActionTestSummary(xcresultPath string, useLegacyFlag bool) (ActionTestSummary, error) {
	if g.SummaryRef.ID.Value == "" {
		return ActionTestSummary{}, ErrSummaryNotFound
	}

	var summary ActionTestSummary
	if err := xcresulttoolGet(xcresultPath, g.SummaryRef.ID.Value, useLegacyFlag, &summary); err != nil {
		return ActionTestSummary{}, fmt.Errorf("failed to load ActionTestSummary: %w", err)
	}
	return summary, nil
}

// exportScreenshots ...
func (g ActionTestSummaryGroup) exportScreenshots(resultPth, outputDir string, useLegacyFlag bool) error {
	if g.TestStatus.Value == "" {
		return nil
	}

	if g.SummaryRef.ID.Value == "" {
		return nil
	}

	var summary ActionTestSummary
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
