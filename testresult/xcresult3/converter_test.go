package xcresult3

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-steputils/v2/testreport"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
)

// copyTestdataToDir ...
// To populate the _tmp dir with sample data
// run `bitrise run download_sample_artifacts` before running tests here,
// which will download https://github.com/bitrise-io/sample-artifacts
// into the _tmp dir.
func copyTestdataToDir(pathInTestdataDir, dirPathToCopyInto string) (string, error) {
	err := command.CopyDir(
		filepath.Join("../../../_tmp/", pathInTestdataDir),
		dirPathToCopyInto,
		true,
	)
	return dirPathToCopyInto, err
}

func TestConverter_XML(t *testing.T) {
	t.Run("xcresult3-flaky-with-rerun.xcresult", func(t *testing.T) {
		fileName := "xcresult3-flaky-with-rerun.xcresult"
		rootDir, xcresultPath, err := setupTestData(fileName)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.RemoveAll(rootDir))
		}()

		t.Log("tempTestdataDir: ", rootDir)

		c := Converter{
			xcresultPth: xcresultPath,
		}
		junitXML, err := c.Convert()
		require.NoError(t, err)
		require.Equal(t, []testreport.TestSuite{
			{
				Name: "BullsEyeTests", Tests: 5, Failures: 0, Skipped: 0, Time: 0.9774,
				TestCases: []testreport.TestCase{
					{
						Name: "testStartNewRoundUsesRandomValueFromApiRequest()", ClassName: "BullsEyeFakeTests",
						Time: 0.014,
					},
					{
						Name: "testGameStyleCanBeChanged()", ClassName: "BullsEyeMockTests",
						Time: 0.0093,
					},
					{
						Name: "testScoreIsComputedPerformance()", ClassName: "BullsEyeTests",
						Time: 0.74,
					},
					{
						Name: "testScoreIsComputedWhenGuessIsHigherThanTarget()", ClassName: "BullsEyeTests",
						Time: 0.0041,
					},
					{
						Name: "testScoreIsComputedWhenGuessIsLowerThanTarget()", ClassName: "BullsEyeTests",
						Time: 0.21,
					},
				},
			},
			{
				Name: "BullsEyeSlowTests", Tests: 2, Failures: 0, Skipped: 0, Time: 0.53,
				TestCases: []testreport.TestCase{
					{
						Name: "testApiCallCompletes()", ClassName: "BullsEyeSlowTests",
						Time: 0.28,
					},
					{
						Name: "testValidApiCallGetsHTTPStatusCode200()", ClassName: "BullsEyeSlowTests",
						Time: 0.25,
					},
				},
			},
			{
				Name: "BullsEyeUITests", Tests: 1, Failures: 0, Skipped: 0, Time: 9,
				TestCases: []testreport.TestCase{
					{
						Name: "testGameStyleSwitch()", ClassName: "BullsEyeUITests",
						Time: 9,
						Properties: &testreport.Properties{
							Property: []testreport.Property{
								{
									Name:  "attachment_0",
									Value: "Screenshot 2022-02-10 at 02.57.39 PM_1644505059194999933.jpeg",
								},
								{
									Name:  "attachment_1",
									Value: "Screenshot 2022-02-10 at 02.57.39 PM_1644505059388999938.jpeg",
								},
								{
									Name:  "attachment_2",
									Value: "Screenshot 2022-02-10 at 02.57.44 PM_1644505064670000076.jpeg",
								},
								{
									Name:  "attachment_3",
									Value: "Screenshot 2022-02-10 at 02.57.47 PM_1644505067144000053.jpeg",
								},
								{
									Name:  "attachment_4",
									Value: "Screenshot 2022-02-10 at 02.57.47 PM_1644505067476000070.jpeg",
								},
								{
									Name:  "attachment_5",
									Value: "Screenshot 2022-02-10 at 02.57.47 PM_1644505067992000102.jpeg",
								},
							},
						},
					},
				},
			},
			{
				Name: "BullsEyeFlakyTests", Tests: 3, Failures: 1, Skipped: 1, Time: 0.226,
				TestCases: []testreport.TestCase{
					{
						Name: "testFlakyFeature()", ClassName: "BullsEyeFlakyTests", Time: 0.2,
						Failure: &testreport.Failure{
							Value: `BullsEyeFlakyTests.swift:43: XCTAssertEqual failed: ("1") is not equal to ("0") - Number is not even`,
						},
					},
					{
						Name: "testFlakyFeature()", ClassName: "BullsEyeFlakyTests", Time: 0.006,
					},
					{
						Name: "testFlakySkip()", ClassName: "BullsEyeSkippedTests", Time: 0.02,
						Skipped: &testreport.Skipped{},
					},
				},
			},
		}, junitXML.TestSuites)
	})

	t.Run("xcresults3 success-failed-skipped-tests.xcresult", func(t *testing.T) {
		fileName := "xcresult3-success-failed-skipped-tests.xcresult"
		rootDir, xcresultPath, err := setupTestData(fileName)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.RemoveAll(rootDir))
		}()

		t.Log("tempTestdataDir: ", rootDir)

		c := Converter{
			xcresultPth: xcresultPath,
		}
		junitXML, err := c.Convert()
		require.NoError(t, err)
		require.Equal(t, []testreport.TestSuite{
			{
				Name:     "testProjectUITests",
				Tests:    3,
				Failures: 1,
				Skipped:  1,
				Time:     0.435,
				TestCases: []testreport.TestCase{
					{
						Name:      "testFailure()",
						ClassName: "testProjectUITests",
						Time:      0.26,
						Failure: &testreport.Failure{
							Value: "testProjectUITests.swift:30: XCTAssertTrue failed",
						},
						Properties: &testreport.Properties{
							Property: []testreport.Property{
								{
									Name:  "attachment_0",
									Value: "Screenshot 2021-02-09 at 08.35.51 AM_1612859751989000082.jpeg",
								},
								{
									Name:  "attachment_1",
									Value: "Screenshot 2021-02-09 at 08.35.52 AM_1612859752052999973.jpeg",
								},
								{
									Name:  "attachment_2",
									Value: "Screenshot 2021-02-09 at 08.35.52 AM_1612859752052999973.jpeg",
								},
							},
						},
					},
					{
						Name:      "testSkip()",
						ClassName: "testProjectUITests",
						Time:      0.086,
						Skipped:   &testreport.Skipped{},
					},
					{
						Name:      "testSuccess()",
						ClassName: "testProjectUITests",
						Time:      0.089,
					},
				},
			},
		}, junitXML.TestSuites)
	})

	t.Run("xcresult3-multiple-test-plan-configurations.xcresult", func(t *testing.T) {
		fileName := "xcresult3-multiple-test-plan-configurations.xcresult"
		rootDir, xcresultPath, err := setupTestData(fileName)
		require.NoError(t, err)
		require.NotEmpty(t, rootDir)
		require.NotEmpty(t, xcresultPath)

		t.Log("tempTestdataDir: ", rootDir)

		c := Converter{xcresultPth: xcresultPath}
		junitXML, err := c.Convert()
		require.NoError(t, err)
		require.NotNil(t, junitXML)

		require.EqualValues(
			t,
			junitXML.TestSuites[0].TestCases[0].Failure.Value,
			`English: swift_testingTests.swift:20: Expectation failed: true == false - // This test is intended to fail to demonstrate test failure reporting.
German: swift_testingTests.swift:20: Expectation failed: true == false - // This test is intended to fail to demonstrate test failure reporting.
`,
		)

	})
}

func BenchmarkConverter_XML(b *testing.B) {
	fileName := "xcresult3-flaky-with-rerun.xcresult"
	rootDir, xcresultPath, err := setupTestData(fileName)
	require.NoError(b, err)

	defer func() {
		require.NoError(b, os.RemoveAll(rootDir))
	}()

	b.Log("tempTestdataDir: ", rootDir)

	c := Converter{
		xcresultPth: xcresultPath,
	}
	_, err = c.Convert()
	require.NoError(b, err)
}

func setupTestData(fileName string) (string, string, error) {
	tempTestdataDir, err := pathutil.NormalizedOSTempDirPath("test")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	tempXCResultPath, err := copyTestdataToDir(fmt.Sprintf("./xcresults/%s", fileName), filepath.Join(tempTestdataDir, fileName))
	if err != nil {
		return "", "", err
	}

	return tempTestdataDir, tempXCResultPath, nil
}
