package certificateutil

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPrivateKey_redactsWhenRendered(t *testing.T) {
	_, key, err := GenerateTestCertificate(1, "TEAMID", "Acme Inc", "Apple Development: Jane Doe", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)

	t.Run("a key is masked by String and MarshalJSON", func(t *testing.T) {
		privateKey := NewPrivateKey(key)

		require.Equal(t, "[REDACTED]", privateKey.String())

		data, err := json.Marshal(privateKey)
		require.NoError(t, err)
		require.Equal(t, `"[REDACTED]"`, string(data))
	})

	t.Run("an absent key renders as empty and null", func(t *testing.T) {
		privateKey := NewPrivateKey(nil)

		require.Equal(t, "", privateKey.String())

		data, err := json.Marshal(privateKey)
		require.NoError(t, err)
		require.Equal(t, "null", string(data))
	})

	t.Run("Key returns the wrapped key unchanged", func(t *testing.T) {
		require.Equal(t, key, NewPrivateKey(key).Key())
		require.Nil(t, NewPrivateKey(nil).Key())
	})

	t.Run("an encoder without a marshaller sees an empty value, not the key", func(t *testing.T) {
		data, err := yaml.Marshal(NewPrivateKey(key))
		require.NoError(t, err)
		require.Equal(t, "{}\n", string(data))
	})
}

// TestCertificateInfoModel_neverRendersPrivateKey covers the ways key material could reach a build
// log: fmt verbs falling back to per-field formatting, and marshalling the whole model. All are
// asserted on a real CertificateInfoModel so the guarantee holds for any caller, not just for
// CertificatePrinter.
func TestCertificateInfoModel_neverRendersPrivateKey(t *testing.T) {
	cert, key, err := GenerateTestCertificate(1234, "TEAMID", "Acme Inc", "Apple Development: Jane Doe", time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)

	info := NewCertificateInfo(*cert, key)

	rsaKey, ok := info.PrivateKey.Key().(*rsa.PrivateKey)
	require.True(t, ok, "test fixture should hold an RSA key")
	secret := rsaKey.D.String()
	require.NotEmpty(t, secret)

	for _, verb := range []string{"%v", "%+v", "%s"} {
		t.Run("formatted with "+verb, func(t *testing.T) {
			rendered := fmt.Sprintf(verb, info)

			require.Contains(t, rendered, "[REDACTED]")
			require.NotContains(t, rendered, secret)
		})
	}

	t.Run("formatted with %#v", func(t *testing.T) {
		rendered := fmt.Sprintf("%#v", info)

		require.NotContains(t, rendered, secret)
		require.Contains(t, rendered, "PrivateKey{key:(*rsa.PrivateKey)")
	})

	t.Run("marshalled to JSON", func(t *testing.T) {
		data, err := json.Marshal(info)
		require.NoError(t, err)

		require.Contains(t, string(data), "[REDACTED]")
		require.NotContains(t, string(data), secret)
		require.NotContains(t, string(data), "Primes")
	})

	t.Run("marshalled to YAML", func(t *testing.T) {
		data, err := yaml.Marshal(info)
		require.NoError(t, err)

		require.NotContains(t, string(data), secret)
		require.NotContains(t, string(data), "primes")
		require.Contains(t, string(data), "privatekey: {}")
	})
}
