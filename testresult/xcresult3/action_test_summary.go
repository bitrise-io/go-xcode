package xcresult3

import (
	"crypto/md5"
	"encoding/hex"
)

type attachment struct {
	Filename struct {
		Value string `json:"_value"`
	} `json:"filename"`

	PayloadRef struct {
		ID struct {
			Value string `json:"_value"`
		}
	} `json:"payloadRef"`
}

type attachments struct {
	Values []attachment `json:"_values"`
}

type actionTestActivitySummary struct {
	Attachments attachments `json:"attachments"`
}

type activitySummaries struct {
	Values []actionTestActivitySummary `json:"_values"`
}

type actionTestFailureSummary struct {
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

type failureSummaries struct {
	Values []actionTestFailureSummary `json:"_values"`
}

type configuration struct {
	Hash string
}

func (c *configuration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	hash := md5.Sum(data)
	c.Hash = hex.EncodeToString(hash[:])

	return nil
}

type actionTestSummary struct {
	ActivitySummaries activitySummaries `json:"activitySummaries"`
	FailureSummaries  failureSummaries  `json:"failureSummaries"`
	Configuration     configuration     `json:"configuration"`
}
