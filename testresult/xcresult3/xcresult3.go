package xcresult3

import "github.com/bitrise-io/go-xcode/v2/testresult/xcresult3/model3"

// Parse parses the given xcresult file's ActionsInvocationRecord and the list of ActionTestPlanRunSummaries.
func Parse(pth string, useLegacyFlag bool) (*ActionsInvocationRecord, []ActionTestPlanRunSummaries, error) {
	var r ActionsInvocationRecord
	if err := xcresulttoolGet(pth, "", useLegacyFlag, &r); err != nil {
		return nil, nil, err
	}

	var summaries []ActionTestPlanRunSummaries
	for _, action := range r.Actions.Values {
		refID := action.ActionResult.TestsRef.ID.Value
		var s ActionTestPlanRunSummaries
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
