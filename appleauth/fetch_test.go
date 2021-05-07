package appleauth

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/stretchr/testify/require"
)

func TestSelect(t *testing.T) {
	type args struct {
		devportalConnection *devportalservice.AppleDeveloperConnection
		authSources         []Source
		inputs              Inputs
	}
	tests := []struct {
		name        string
		args        args
		want        Credentials
		wantErr     bool
		wantErrType error
	}{
		{
			name: "No connection active (nil), no inputs",
			args: args{
				devportalConnection: nil,
				authSources:         []Source{&ConnectionAPIKeySource{}, &ConnectionAppleIDSource{}, &InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs:              Inputs{},
			},
			want:        Credentials{},
			wantErr:     true,
			wantErrType: &MissingAuthConfigError{},
		},
		{
			name: "No connection active (empty), no inputs",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{},
				authSources:         []Source{&ConnectionAPIKeySource{}, &ConnectionAppleIDSource{}, &InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs:              Inputs{},
			},
			want:        Credentials{},
			wantErr:     true,
			wantErrType: &MissingAuthConfigError{},
		},
		{
			name: "No connection active (empty, error), inputs (Apple ID)",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{},
				authSources:         []Source{&ConnectionAPIKeySource{}, &ConnectionAppleIDSource{}, &InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs: Inputs{
					Username: "a", Password: "b", AppSpecificPassword: "c",
					APIIssuer: "", APIKeyPath: "",
				},
			},
			want: Credentials{
				AppleID: &AppleID{
					Username: "a", Password: "b", AppSpecificPassword: "c", Session: "",
				},
				APIKey: nil,
			},
		},
		{
			name: "Connection active (API Key), inputs (Apple ID)",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					APIKeyConnection: &devportalservice.APIKeyConnection{
						KeyID: "x", IssuerID: "y", PrivateKey: "z",
					},
				},
				authSources: []Source{&ConnectionAPIKeySource{}, &ConnectionAppleIDSource{}, &InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs: Inputs{
					Username: "a", Password: "b", AppSpecificPassword: "c",
					APIIssuer: "", APIKeyPath: "",
				},
			},
			want: Credentials{
				AppleID: nil,
				APIKey: &devportalservice.APIKeyConnection{
					KeyID: "x", IssuerID: "y", PrivateKey: "z",
				},
			},
		},
		{
			name: "Connection active (API Key), inputs (Apple ID), connection not enabled",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					APIKeyConnection: &devportalservice.APIKeyConnection{
						KeyID: "x", IssuerID: "y", PrivateKey: "z",
					},
				},
				authSources: []Source{&InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs: Inputs{
					Username: "a", Password: "b", AppSpecificPassword: "c",
					APIIssuer: "", APIKeyPath: "",
				},
			},
			want: Credentials{
				AppleID: &AppleID{
					Username: "a", Password: "b", AppSpecificPassword: "c", Session: "",
				},
				APIKey: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Select(tt.args.devportalConnection, tt.args.authSources, tt.args.inputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, reflect.TypeOf(tt.wantErrType), reflect.TypeOf(err), "Select() error type")
			require.Equal(t, tt.want, got, "Select() =")
		})
	}
}
