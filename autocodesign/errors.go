package autocodesign

import (
	"fmt"

	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
)

// missingCertificateError ...
type missingCertificateError struct {
	Type   appstoreconnect.CertificateType
	TeamID string
}

func (e missingCertificateError) Error() string {
	return fmt.Sprintf("no valid %s type certificates uploaded with Team ID (%s)\n ", e.Type, e.TeamID)
}

// NonmatchingProfileError is returned when a profile/bundle ID does not match project requirements
// It is not a fatal error, as the profile can be regenerated
type NonmatchingProfileError struct {
	Reason string
}

func (e NonmatchingProfileError) Error() string {
	return fmt.Sprintf("provisioning profile does not match requirements: %s", e.Reason)
}
