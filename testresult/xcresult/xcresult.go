package xcresult

import (
	"path/filepath"
	"strings"
	"unicode"

	"howett.net/plist"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-steputils/v2/testreport"
)

// Converter ...
type Converter struct {
	files                  []string
	testSummariesPlistPath string
}

// Setup configures the converter.
func (c *Converter) Setup(_ bool) {}

// Detect ...
func (c *Converter) Detect(files []string) bool {
	c.files = files
	for _, file := range c.files {
		if filepath.Ext(file) == ".xcresult" {
			testSummariesPlistPath := filepath.Join(file, "TestSummaries.plist")
			if exist, err := pathutil.IsPathExists(testSummariesPlistPath); err != nil || !exist {
				continue
			}

			c.testSummariesPlistPath = testSummariesPlistPath
			return true
		}
	}
	return false
}

// by one of our issue reports, need to replace backspace char (U+0008) as it is an invalid character for xml unmarshaller
// the legal character ranges are here: https://www.w3.org/TR/REC-xml/#charsets
// so the exclusion will be:
/*
	\u0000 - \u0008
	\u000B
	\u000C
	\u000E - \u001F
	\u007F - \u0084
	\u0086 - \u009F
	\uD800 - \uDFFF

	Unicode range D800–DFFF is used as surrogate pair. Unicode and ISO/IEC 10646 do not assign characters to any of the code points in the D800–DFFF range, so an individual code value from a surrogate pair does not represent a character. (A couple of code points — the first from the high surrogate area (D800–DBFF), and the second from the low surrogate area (DC00–DFFF) — are used in UTF-16 to represent a character in supplementary planes)
	\uFDD0 - \uFDEF; \uFFFE; \uFFFF
*/
// These are non-characters in the standard, not assigned to anything; and have no meaning.
func filterIllegalChars(data []byte) (filtered []byte) {
	illegalCharFilter := func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}
	filtered = []byte(strings.Map(illegalCharFilter, string(data)))
	return
}

// Convert returns the test report parsed from the xcresult file.
func (c *Converter) Convert() (testreport.TestReport, error) {
	data, err := fileutil.ReadBytesFromFile(c.testSummariesPlistPath)
	if err != nil {
		return testreport.TestReport{}, err
	}

	data = filterIllegalChars(data)

	var plistData testSummaryPlist
	if _, err := plist.Unmarshal(data, &plistData); err != nil {
		return testreport.TestReport{}, err
	}

	var xmlData testreport.TestReport
	keyOrder, tests := plistData.tests()
	for _, testID := range keyOrder {
		tests := tests[testID]
		testSuite := testreport.TestSuite{
			Name:     testID,
			Tests:    len(tests),
			Failures: tests.failuresCount(),
			Skipped:  tests.skippedCount(),
			Time:     tests.totalTime(),
		}

		for _, test := range tests {
			failureMessage := test.failure()

			var failure *testreport.Failure
			if len(failureMessage) > 0 {
				failure = &testreport.Failure{
					Value: failureMessage,
				}
			}

			var skipped *testreport.Skipped
			if test.skipped() {
				skipped = &testreport.Skipped{}
			}

			testSuite.TestCases = append(testSuite.TestCases, testreport.TestCase{
				Name:      test.TestName,
				ClassName: testID,
				Failure:   failure,
				Skipped:   skipped,
				Time:      test.Duration,
			})
		}

		xmlData.TestSuites = append(xmlData.TestSuites, testSuite)
	}

	return xmlData, nil
}
