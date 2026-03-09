package xcresult3

import (
	"crypto/md5"
	"encoding/hex"
)

// Attachment ...
type Attachment struct {
	Filename struct {
		Value string `json:"_value"`
	} `json:"filename"`

	PayloadRef struct {
		ID struct {
			Value string `json:"_value"`
		}
	} `json:"payloadRef"`
}

// Attachments ...
type Attachments struct {
	Values []Attachment `json:"_values"`
}

// ActionTestActivitySummary ...
type ActionTestActivitySummary struct {
	Attachments Attachments `json:"attachments"`
}

// ActivitySummaries ...
type ActivitySummaries struct {
	Values []ActionTestActivitySummary `json:"_values"`
}

// ActionTestFailureSummary ...
type ActionTestFailureSummary struct {
	Message struct {
		Value string `json:"_value"`
	} `json:"message"`

	FileName struct {
		Value string `json:"_value"`
	} `json:"fileName"`

	LineNumber struct {
		Value string `json:"_value"`
	} `json:"lineNumber"`
}

// FailureSummaries ...
type FailureSummaries struct {
	Values []ActionTestFailureSummary `json:"_values"`
}

// Configuration ...
type Configuration struct {
	Hash string
}

// UnmarshalJSON ...
func (c *Configuration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	hash := md5.Sum(data)
	c.Hash = hex.EncodeToString(hash[:])

	return nil
}

// ActionTestSummary ...
type ActionTestSummary struct {
	ActivitySummaries ActivitySummaries `json:"activitySummaries"`
	FailureSummaries  FailureSummaries  `json:"failureSummaries"`
	Configuration     Configuration     `json:"configuration"`
}
