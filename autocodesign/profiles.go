package autocodesign

import (
	"fmt"
	"strings"
)

// NonmatchingProfileError is returned when a profile/bundle ID does not match project requirements
// It is not a fatal error, as the profile can be regenerated
type NonmatchingProfileError struct {
	Reason string
}

func (e NonmatchingProfileError) Error() string {
	return fmt.Sprintf("provisioning profile does not match requirements: %s", e.Reason)
}

// AppIDName ...
func AppIDName(bundleID string) string {
	prefix := ""
	if strings.HasSuffix(bundleID, ".*") {
		prefix = "Wildcard "
	}
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ", "*", " ")
	return prefix + "Bitrise " + r.Replace(bundleID)
}
