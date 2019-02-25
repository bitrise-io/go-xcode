package xcarchive

import "strings"

func escapeGlobPath(path string) string {
	path = strings.Replace(path, "\\", "\\\\", -1)
	for _, char := range []string{"[", "*", "?"} {
		path = strings.Replace(path, char, "\\"+char, -1)
	}
	return path
}
