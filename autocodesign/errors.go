package autocodesign

import (
	"fmt"

	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
)

// missingCertificateError ...
type missingCertificateError struct {
	Type   appstoreconnect.CertificateType
	TeamID string
}

func (e missingCertificateError) Error() string {
	return fmt.Sprintf("no valid %s type certificates uploaded with Team ID (%s)\n ", e.Type, e.TeamID)
}
