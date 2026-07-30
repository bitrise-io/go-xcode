package certificateutil

import (
	"fmt"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/timeutil"
)

// CertificatePrinter renders certificate infos for logging.
type CertificatePrinter struct {
	logger       log.Logger
	timeProvider timeutil.TimeProvider
}

// NewCertificatePrinter ...
func NewCertificatePrinter(logger log.Logger, timeProvider timeutil.TimeProvider) *CertificatePrinter {
	return &CertificatePrinter{
		logger:       logger,
		timeProvider: timeProvider,
	}
}

// CertificateSummary returns a single-line description of the certificate, including error string
// if the cert is not valid
func (printer *CertificatePrinter) CertificateSummary(info CertificateInfoModel) string {
	team := fmt.Sprintf("%s (%s)", info.TeamName, info.TeamID)
	summary := fmt.Sprintf("Serial: %s, Name: %s, Team: %s, Expiry: %s", info.Serial, info.CommonName, team, info.EndDate)

	if err := checkValidityAt(printer.timeProvider.Now(), info.Certificate); err != nil {
		summary += fmt.Sprintf(", error: %s", err)
	}

	return summary
}
