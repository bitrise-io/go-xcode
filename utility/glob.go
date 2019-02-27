package utility

func escapeGlobPath(path string) string {
	var escaped string
	for _, ch := range path {
		if ch == '[' || ch == ']' || ch == '-' || ch == '*' || ch == '?' || ch == '\\' {
			escaped += "\\"
		}
		escaped += string(ch)
	}
	return escaped
}
