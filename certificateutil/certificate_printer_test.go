package certificateutil

import (
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/require"
)

type fakeTimeProvider struct {
	now time.Time
}

func (p fakeTimeProvider) Now() time.Time {
	return p.now
}

// printerAt returns a printer whose clock is fixed at the given time, so validity-dependent output
// can be exercised without generating expired fixtures or depending on the wall clock.
func printerAt(now time.Time) *CertificatePrinter {
	return NewCertificatePrinter(log.NewLogger(), fakeTimeProvider{now: now})
}

func TestCertificatePrinter_CertificateLabel(t *testing.T) {
	cert, _, err := GenerateTestCertificate(1234, "TEAMID", "Acme Inc", "Apple Development: Jane Doe", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	info := NewCertificateInfo(*cert, nil)

	t.Run("renders the common name and serial", func(t *testing.T) {
		got := printerAt(cert.NotBefore.Add(time.Hour)).CertificateLabel(info)

		require.Equal(t, "Apple Development: Jane Doe (1234)", got)
	})

	t.Run("does not report validity, so it is unaffected by the clock", func(t *testing.T) {
		expired := printerAt(cert.NotAfter.Add(time.Hour)).CertificateLabel(info)

		require.Equal(t, "Apple Development: Jane Doe (1234)", expired)
		require.NotContains(t, expired, "error")
	})
}

func TestCertificatePrinter_CertificateSummary(t *testing.T) {
	cert, _, err := GenerateTestCertificate(1234, "TEAMID", "Acme Inc", "Apple Development: Jane Doe", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	info := NewCertificateInfo(*cert, nil)

	t.Run("valid certificate renders its human-readable fields without an error", func(t *testing.T) {
		got := printerAt(cert.NotBefore.Add(time.Hour)).CertificateSummary(info)

		require.Contains(t, got, "Serial: 1234")
		require.Contains(t, got, "Name: Apple Development: Jane Doe")
		require.Contains(t, got, "Team: Acme Inc (TEAMID)")
		require.Contains(t, got, "Expiry: ")
		require.NotContains(t, got, "error:")
	})

	t.Run("expired certificate appends the validity error", func(t *testing.T) {
		got := printerAt(cert.NotAfter.Add(time.Hour)).CertificateSummary(info)

		require.Contains(t, got, "Serial: 1234")
		require.Contains(t, got, "Name: Apple Development: Jane Doe")
		require.Contains(t, got, "error:")
		require.Contains(t, got, "not valid anymore")
	})

	t.Run("not yet valid certificate appends the validity error", func(t *testing.T) {
		got := printerAt(cert.NotBefore.Add(-time.Hour)).CertificateSummary(info)

		require.Contains(t, got, "Serial: 1234")
		require.Contains(t, got, "error:")
		require.Contains(t, got, "not yet valid")
	})
}
