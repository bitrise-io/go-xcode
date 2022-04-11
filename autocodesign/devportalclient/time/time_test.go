package time

import (
	"testing"
)

func TestTime_UnmarshalJSON(t *testing.T) {

	tests := []struct {
		b       []byte
		name    string
		t       *Time
		wantErr bool
	}{
		{name: "without quotation mark", b: []byte("2021-05-19T08:07:47.000+00:00"), wantErr: false},
		{name: "with quotation mark", b: []byte(`"2021-05-19T08:07:47.000+00:00"`), wantErr: false},
		{name: "Positive offset", b: []byte("2021-05-19T08:07:47.000+05:30"), wantErr: false},
		{name: "Alternate positive offset", b: []byte("2021-05-19T08:07:47.000+0530"), wantErr: false},
		{name: "Alternate negative offset", b: []byte("2021-05-19T08:07:47.000-0400"), wantErr: false},
		{name: "Single hour positive offset", b: []byte("2021-05-19T08:07:47.000+03"), wantErr: false},
		{name: "Single hour negative offset", b: []byte("2021-05-19T08:07:47.000-02"), wantErr: false},
		{name: "Zero offset UTC", b: []byte("2021-05-19T08:07:47.000Z"), wantErr: false},
		{name: "Custom spaceship time format", b: []byte("2022-04-01 12:45:25 UTC"), wantErr: false},
		{name: "unsupported format", b: []byte("2021-12-17T10:44:00Z00:00"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time := &Time{}
			if err := time.UnmarshalJSON(tt.b); (err != nil) != tt.wantErr {
				t.Errorf("Time.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
