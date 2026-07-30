package certificateutil

// redacted is what a PrivateKey renders instead of its contents.
const redacted = "[REDACTED]"

// PrivateKey holds a certificate's private key. It deliberately never renders its contents:
// String and MarshalJSON both redact, so a CertificateInfoModel cannot leak key material through
// %s/%v formatting or JSON marshalling, however it is embedded or logged.
type PrivateKey struct {
	key any
}

// NewPrivateKey ...
func NewPrivateKey(key any) PrivateKey {
	return PrivateKey{key: key}
}

// Key returns the underlying private key. Certificates read from the Keychain have none, in which
// case this is nil.
func (k PrivateKey) Key() any {
	return k.key
}

// String implements fmt.Stringer, masking the key.
func (k PrivateKey) String() string {
	if k.key == nil {
		return ""
	}
	return redacted
}

// MarshalJSON implements json.Marshaler, masking the key.
func (k PrivateKey) MarshalJSON() ([]byte, error) {
	if k.key == nil {
		return []byte("null"), nil
	}
	return []byte(`"` + redacted + `"`), nil
}
