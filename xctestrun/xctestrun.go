package xctestrun

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-steputils/v2/testquarantine"
)

/*
ParseQuarantinedTests converts the Bitrise quarantined tests JSON input ($BITRISE_QUARANTINED_TESTS_JSON)
to xctestrun file's SkipTestIdentifiers format: TestClass/TestMethod (`()` suffix removed) mapped by TestTargets.
*/
func ParseQuarantinedTests(quarantinedTestsInput string) (map[string][]string, error) {
	quarantinedTests, err := testquarantine.ParseQuarantinedTests(quarantinedTestsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quarantined tests input: %w", err)
	}

	skippedTestsByTarget := map[string][]string{}
	for _, qt := range quarantinedTests {
		if len(qt.TestSuiteName) == 0 || qt.TestSuiteName[0] == "" || qt.ClassName == "" || qt.TestCaseName == "" {
			continue
		}

		testTarget := qt.TestSuiteName[0]
		testClass := qt.ClassName
		testMethod := strings.TrimSuffix(qt.TestCaseName, "()")

		skippedTests := skippedTestsByTarget[testTarget]
		skippedTests = append(skippedTests, fmt.Sprintf("%s/%s", testClass, testMethod))
		skippedTestsByTarget[testTarget] = skippedTests
	}

	return skippedTestsByTarget, nil
}

// AddQuarantinedTestsToXctestrun adds the given skipped tests to the xctestrun file's SkipTestIdentifiers.
func AddQuarantinedTestsToXctestrun(xctestrunPth string, skippedTestByTarget map[string][]string) error {
	xctestrun, plistFormat, err := parseXctestrun(xctestrunPth)
	if err != nil {
		return err
	}

	updatedXctestrun, err := addSkippedTestsToXctestrun(xctestrun, skippedTestByTarget)
	if err != nil {
		return err
	}

	if err := writeXctestrun(xctestrunPth, updatedXctestrun, plistFormat); err != nil {
		return err
	}

	return nil
}

func parseXctestrun(xctestrunPth string) (map[string]any, int, error) {
	xctestrunContent, err := os.ReadFile(xctestrunPth)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read xctestrun file: %w", err)
	}

	var xctestrun map[string]any
	format, err := plist.Unmarshal(xctestrunContent, &xctestrun)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal xctestrun plist: %w", err)
	}

	return xctestrun, format, nil
}

func writeXctestrun(xctestrunPth string, xctestrun map[string]any, format int) error {
	updatedXctestrunContent, err := plist.Marshal(xctestrun, format)
	if err != nil {
		return fmt.Errorf("failed to marshal xctestrun plist: %w", err)
	}

	if err := os.WriteFile(xctestrunPth, updatedXctestrunContent, 0644); err != nil {
		return fmt.Errorf("failed to write updated xctestrun file: %w", err)
	}

	return nil
}

func addSkippedTestsToXctestrun(xctestrun map[string]any, skippedTestByTarget map[string][]string) (map[string]any, error) {
	testConfigurationsRaw, ok := xctestrun["TestConfigurations"]
	if !ok {
		return nil, fmt.Errorf("TestConfigurations not found in xctestrun")
	}

	testConfigurationsSlice, ok := testConfigurationsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid TestConfigurations format in xctestrun")
	}

	for testConfigurationIdx, testConfigurationRaw := range testConfigurationsSlice {
		testConfiguration, ok := testConfigurationRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid test configuration format in xctestrun")
		}

		testTargetsRaw, ok := testConfiguration["TestTargets"]
		if !ok {
			return nil, fmt.Errorf("TestTargets not found in test configuration")
		}

		testTargetsSlice, ok := testTargetsRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid TestTargets format in test configuration")
		}

		for testTargetIdx, testTargetRaw := range testTargetsSlice {
			testTarget, ok := testTargetRaw.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid test target format in test configuration")
			}

			blueprintNameRaw, ok := testTarget["BlueprintName"]
			if !ok {
				return nil, fmt.Errorf("BlueprintName not found in test target")
			}

			blueprintName, ok := blueprintNameRaw.(string)
			if !ok {
				return nil, fmt.Errorf("invalid BlueprintName format in test target")
			}

			skippedTestsToAdd, ok := skippedTestByTarget[blueprintName]
			if !ok {
				continue
			}

			var skipTestIdentifiers []interface{}
			skipTestIdentifiersRaw, ok := testTarget["SkipTestIdentifiers"]
			if ok {
				skipTestIdentifiers, ok = skipTestIdentifiersRaw.([]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid SkipTestIdentifiers format in test target")
				}
			}

			for _, skippedTestsToAddItem := range skippedTestsToAdd {
				skipTestIdentifiers = append(skipTestIdentifiers, skippedTestsToAddItem)
			}

			testTarget["SkipTestIdentifiers"] = skipTestIdentifiers
			testTargetsSlice[testTargetIdx] = testTarget
		}

		testConfiguration["TestTargets"] = testTargetsSlice
		testConfigurationsSlice[testConfigurationIdx] = testConfiguration
	}

	xctestrun["TestConfigurations"] = testConfigurationsSlice

	return xctestrun, nil
}
