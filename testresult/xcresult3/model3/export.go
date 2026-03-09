package model3

import (
	"encoding/json"
	"time"
)

type TestAttachmentDetails struct {
	TestIdentifier string       `json:"testIdentifier"`
	Attachments    []Attachment `json:"attachments"`
}

type Timestamp time.Time

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
