package xcresult3

import "github.com/bitrise-io/go-xcode/v2/testresult/xcresult3/model3"

// loadXCResultData loads the actions invocation record and test plan run summaries from an xcresult file.
func loadXCResultData(pth string, useLegacyFlag bool) (*actionsInvocationRecord, []actionTestPlanRunSummaries, error) {
	var r actionsInvocationRecord
	if err := xcresulttoolGet(pth, "", useLegacyFlag, &r); err != nil {
		return nil, nil, err
	}

	var summaries []actionTestPlanRunSummaries
	for _, action := range r.Actions.Values {
		refID := action.ActionResult.TestsRef.ID.Value
		var s actionTestPlanRunSummaries
		if err := xcresulttoolGet(pth, refID, useLegacyFlag, &s); err != nil {
			return nil, nil, err
		}
		summaries = append(summaries, s)
	}
	return &r, summaries, nil
}

// ParseTestResults parses the test results from the given xcresult file.
func ParseTestResults(pth string) (*model3.TestData, error) {
	var data model3.TestData
	if err := xcresulttoolGet(pth, "", false, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
