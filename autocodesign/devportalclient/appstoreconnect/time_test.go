package appstoreconnect

import "testing"

func TestTime_UnmarshalJSON(t *testing.T) {

	tests := []struct {
		b       []byte
		name    string
		t       *Time
		wantErr bool
	}{
		{name: "without quotation mark", b: []byte("2021-05-19T08:07:47.000+0000"), wantErr: false},
		{name: "with quotation mark", b: []byte(`"2021-05-19T08:07:47.000+0000"`), wantErr: false},
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
