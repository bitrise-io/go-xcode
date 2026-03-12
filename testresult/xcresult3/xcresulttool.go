package xcresult3

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	command2 "github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

func isXcresulttoolAvailable() bool {
	if _, err := exec.LookPath("xcrun"); err != nil {
		return false
	}
	return command.New("xcrun", "--find", "xcresulttool").Run() == nil
}

func xcresulttoolVersion() (int, error) {
	args := []string{"xcresulttool", "version"}
	cmd := command.New("xcrun", args...)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return 0, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
		}
		return 0, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), err)
	}
	// xcresulttool version 23025, format version 3.53 (current)
	versionRegexp := regexp.MustCompile("xcresulttool version ([0-9]+)")

	matches := versionRegexp.FindStringSubmatch(out)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no version matches found in output: %s", out)
	}

	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert version: %s", matches[1])
	}

	return version, nil
}

func isXcode16OrNewer() bool {
	version, err := xcresulttoolVersion()
	if err != nil {
		return false
	}

	return version >= 23_000 // Xcode 16 beta1 has version 23000
}

func supportsNewExtractionMethods() (bool, error) {
	version, err := xcresulttoolVersion()
	if err != nil {
		return false, err
	}

	return version >= 23_021, nil // Xcode 16 beta3 has version 23021
}

// xcresulttoolGet performs xcrun xcresulttool get with --id flag defined if id provided and marshals the output into v.
func xcresulttoolGet(xcresultPth, id string, useLegacyFlag bool, v interface{}) error {
	commandFactory := command2.NewFactory(env.NewRepository())
	logger := log.NewLogger()

	args := []string{"xcresulttool", "get"}

	supportsNewMethod, err := supportsNewExtractionMethods()
	if err != nil {
		return err
	}

	if supportsNewMethod && !useLegacyFlag {
		args = append(args, "test-results", "tests")
	} else {
		args = append(args, "--format", "json")

		if isXcode16OrNewer() && useLegacyFlag {
			args = append(args, "--legacy")
		}
	}

	args = append(args, "--path", xcresultPth)

	if id != "" {
		args = append(args, "--id", id)
	}

	var outBuffer, errBuffer, combinedBuffer bytes.Buffer
	outWriter := io.MultiWriter(&outBuffer, &combinedBuffer)
	errWriter := io.MultiWriter(&errBuffer, &combinedBuffer)

	cmd := commandFactory.Create("xcrun", args, &command2.Opts{
		Stdout:      outWriter,
		Stderr:      errWriter,
		Stdin:       nil,
		Env:         os.Environ(),
		Dir:         "",
		ErrorFinder: nil,
	})
	if err := cmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), combinedBuffer.String())
		}
		return fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), err)
	}
	if stdErr := errBuffer.String(); stdErr != "" {
		logger.Warnf("%s: %s", cmd.PrintableCommandArgs(), stdErr)
	}

	stdout := outBuffer.Bytes()
	if err := json.Unmarshal(stdout, v); err != nil {
		logger.Warnf("Failed to parse %s command output, first lines:\n%s", cmd.PrintableCommandArgs(), firstLines(string(stdout), 10))
		return err
	}
	return nil
}

func firstLines(out string, count int) string {
	if count < 1 {
		return ""
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) >= count {
			break
		}
	}
	return strings.Join(lines, "\n")
}

// xcresulttoolExport exports a file with the given id at the given output path.
func xcresulttoolExport(xcresultPth, id, outputPth string, useLegacyFlag bool) error {
	args := []string{"xcresulttool", "export"}

	supportsNewMethod, err := supportsNewExtractionMethods()
	if err != nil {
		return err
	}

	if supportsNewMethod && !useLegacyFlag {
		args = append(args, "attachments")
	} else {
		args = append(args, "--type", "file")

		if isXcode16OrNewer() && useLegacyFlag {
			args = append(args, "--legacy")
		}
	}

	args = append(args, "--path", xcresultPth)
	args = append(args, "--output-path", outputPth)

	if id != "" {
		args = append(args, "--id", id)
	}

	cmd := command.New("xcrun", args...)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
		}
		return fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), err)
	}
	return nil
}
