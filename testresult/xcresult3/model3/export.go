package model3

import (
	"encoding/json"
	"time"
)

// TestAttachmentDetails contains the test identifier and its list of attachments.
type TestAttachmentDetails struct {
	TestIdentifier string       `json:"testIdentifier"`
	Attachments    []Attachment `json:"attachments"`
}

// Timestamp is a time.Time that unmarshals from a Unix epoch float.
type Timestamp time.Time

// Attachment describes a single exported test attachment file.
type Attachment struct {
	ExportedFileName           string    `json:"exportedFileName"`
	SuggestedHumanReadableName string    `json:"suggestedHumanReadableName"`
	IsAssociatedWithFailure    bool      `json:"isAssociatedWithFailure"`
	Timestamp                  Timestamp `json:"timestamp"`
	ConfigurationName          string    `json:"configurationName"`
	DeviceName                 string    `json:"deviceName"`
	DeviceID                   string    `json:"deviceId"`
	RepetitionNumber           int       `json:"repetitionNumber"`
}

// UnmarshalJSON decodes a Unix epoch float into a Timestamp.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var timestamp float64
	if err := json.Unmarshal(b, &timestamp); err != nil {
		return err
	}

	// Extract seconds and nanoseconds separately to preserve fractional part
	seconds := int64(timestamp)
	nanoseconds := int64((timestamp - float64(seconds)) * 1e9)

	*t = Timestamp(time.Unix(seconds, nanoseconds))
	return nil
}
