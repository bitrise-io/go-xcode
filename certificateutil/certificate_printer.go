package certificateutil

import (
	"encoding/json"
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

// CertificateLabel returns the shortest identifying form of the certificate
func (printer *CertificatePrinter) CertificateLabel(info CertificateInfoModel) string {
	return fmt.Sprintf("%s (%s)", info.CommonName, info.Serial)
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

// PrintableCertificate returns a verbose, indented-JSON description of the certificate, for debug
// logging alongside profileutil's PrintableProfile.
//
// The certificate's raw x509 form and its private key are deliberately never included.
func (printer *CertificatePrinter) PrintableCertificate(info CertificateInfoModel) string {
	printable := map[string]any{}
	printable["name"] = info.CommonName
	printable["serial"] = info.Serial
	printable["team"] = fmt.Sprintf("%s (%s)", info.TeamName, info.TeamID)
	printable["sha1_fingerprint"] = info.SHA1Fingerprint
	printable["start_date"] = info.StartDate.String()
	printable["expiry"] = info.EndDate.String()

	var errs []string
	if err := checkValidityAt(printer.timeProvider.Now(), info.Certificate); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		printable["errors"] = errs
	}

	data, err := json.MarshalIndent(printable, "", "\t")
	if err != nil {
		printer.logger.Errorf("Failed to marshal: %v, error: %s", printable, err)
		return ""
	}

	return string(data)
}
