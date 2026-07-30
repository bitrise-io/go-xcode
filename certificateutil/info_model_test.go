package certificateutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewCertificateInfo_mapsFields(t *testing.T) {
	const (
		commonName = "Apple Development: Jane Doe"
		teamID     = "TEAMID"
		teamName   = "Acme Inc"
	)
	expiry := time.Now().AddDate(1, 0, 0)
	cert, privateKey, err := GenerateTestCertificate(42, teamID, teamName, commonName, expiry)
	require.NoError(t, err)

	info := NewCertificateInfo(*cert, privateKey)

	require.Equal(t, commonName, info.CommonName)
	require.Equal(t, teamID, info.TeamID)
	require.Equal(t, teamName, info.TeamName)
	require.Equal(t, "42", info.Serial)
	require.NotEmpty(t, info.SHA1Fingerprint)
	require.Equal(t, privateKey, info.PrivateKey.Key())
	require.WithinDuration(t, expiry, info.EndDate, time.Second)
}

func TestCheckValidityAt(t *testing.T) {
	cert, _, err := GenerateTestCertificate(1, "TEAM", "Team", "Name", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)

	tests := []struct {
		name    string
		at      time.Time
		wantErr bool
	}{
		{name: "before validity window", at: cert.NotBefore.Add(-time.Hour), wantErr: true},
		{name: "within validity window", at: cert.NotBefore.Add(time.Hour), wantErr: false},
		{name: "after validity window", at: cert.NotAfter.Add(time.Hour), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkValidityAt(tt.at, *cert)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCertificateInfoModel_String(t *testing.T) {
	t.Run("valid certificate renders its human-readable fields without an error", func(t *testing.T) {
		cert, _, err := GenerateTestCertificate(1234, "TEAMID", "Acme Inc", "Apple Development: Jane Doe", time.Now().AddDate(1, 0, 0))
		require.NoError(t, err)

		got := NewCertificateInfo(*cert, nil).String()

		require.Contains(t, got, "Serial: 1234")
		require.Contains(t, got, "Name: Apple Development: Jane Doe")
		require.Contains(t, got, "Team: Acme Inc (TEAMID)")
		require.Contains(t, got, "Expiry: ")
		require.NotContains(t, got, "error:")
	})

	t.Run("expired certificate appends the validity error", func(t *testing.T) {
		cert, _, err := GenerateTestCertificate(2, "TEAMID", "Acme Inc", "Apple Development: Jane Doe", time.Now().AddDate(-1, 0, 0))
		require.NoError(t, err)

		got := NewCertificateInfo(*cert, nil).String()

		require.Contains(t, got, "Serial: 2")
		require.Contains(t, got, "Name: Apple Development: Jane Doe")
		require.Contains(t, got, "error:")
	})
}
