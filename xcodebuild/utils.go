package xcodebuild

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

func ParseShowBuildSettingsCommandOutput(out string) (map[string]any, error) {
	settings := map[string]any{}
	var buffer bytes.Buffer
	reader := bufio.NewReader(strings.NewReader(out))

	for {
		b, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		lineFragment := string(b)
		buffer.WriteString(lineFragment)

		// isPrefix is set to false once a full line has been read
		if isPrefix == false {
			line := strings.TrimSpace(buffer.String())

			if split := strings.Split(line, "="); len(split) > 1 {
				key := strings.TrimSpace(split[0])
				value := strings.TrimSpace(strings.Join(split[1:], "="))
				value = strings.Trim(value, `"`)

				settings[key] = value
			}

			buffer.Reset()
		}
	}

	return settings, nil
}
