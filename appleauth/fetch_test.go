package appleauth

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/stretchr/testify/require"
)

var (
	argInput = Inputs{
		Username: "input_username", Password: "input_password", AppSpecificPassword: "input_appspecificpassword",
		APIIssuer: "", APIKeyPath: "",
	}

	argAppleIDConnection = devportalservice.AppleIDConnection{
		AppleID: "connection_appleid", Password: "connection_password", AppSpecificPassword: "connection_appspecificpassword",
	}

	argAppleIDConnectionMissingPassword = devportalservice.AppleIDConnection{
		AppleID: "connection_appleid", Password: "connection_password", AppSpecificPassword: "",
	}

	argAPIKeyConnection = devportalservice.APIKeyConnection{
		KeyID: "keyconnection_keyID", IssuerID: "keyconnection_issuerID", PrivateKey: "keyconnection_PrivateKey",
	}
)

var (
	expectedAppleIDWithArgInput = AppleID{
		Username:            argInput.Username,
		Password:            argInput.Password,
		AppSpecificPassword: argInput.AppSpecificPassword,
		Session:             "",
	}

	expectedAppleIDWithArgConnection = AppleID{
		Username:            argAppleIDConnection.AppleID,
		Password:            argAppleIDConnection.Password,
		AppSpecificPassword: argAppleIDConnection.AppSpecificPassword,
		Session:             "",
	}

	expectedAppleIDWithAPIKeyConnection = devportalservice.APIKeyConnection{
		KeyID:      argAPIKeyConnection.KeyID,
		IssuerID:   argAPIKeyConnection.IssuerID,
		PrivateKey: argAPIKeyConnection.PrivateKey,
	}
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
				inputs:              argInput,
			},
			want: Credentials{
				AppleID: &expectedAppleIDWithArgInput,
				APIKey:  nil,
			},
		},
		{
			name: "Connection active (Apple ID), inputs (Apple ID) with ConnectionAppleIDSource",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					AppleIDConnection: &argAppleIDConnection,
				},
				authSources: []Source{&ConnectionAppleIDSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: &expectedAppleIDWithArgConnection,
				APIKey:  nil,
			},
		},
		{
			name: "Connection active (Apple ID), inputs (Apple ID) with InputAppleIDSource",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					AppleIDConnection: &argAppleIDConnection,
				},
				authSources: []Source{&InputAppleIDSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: &expectedAppleIDWithArgInput,
				APIKey:  nil,
			},
		},
		{
			name: "Connection active (Apple ID), inputs (Apple ID) with ConnectionAppleIDFastlaneSource",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					AppleIDConnection: &argAppleIDConnection,
				},
				authSources: []Source{&ConnectionAppleIDFastlaneSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: &expectedAppleIDWithArgConnection,
				APIKey:  nil,
			},
		},
		{
			name: "Connection active but missing password (Apple ID), inputs (Apple ID) with ConnectionAppleIDFastlaneSource",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					AppleIDConnection: &argAppleIDConnectionMissingPassword,
				},
				authSources: []Source{&ConnectionAppleIDFastlaneSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: &AppleID{
					Username:            argAppleIDConnection.AppleID,
					Password:            argAppleIDConnection.Password,
					AppSpecificPassword: argInput.AppSpecificPassword,
					Session:             "",
				},
				APIKey: nil,
			},
		},
		{
			name: "Connection active (Apple ID), inputs (Apple ID) with InputAppleIDFastlaneSource",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					AppleIDConnection: &argAppleIDConnection,
				},
				authSources: []Source{&InputAppleIDFastlaneSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: &expectedAppleIDWithArgInput,
				APIKey:  nil,
			},
		},
		{
			name: "Connection active (API Key), inputs (Apple ID)",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					APIKeyConnection: &argAPIKeyConnection,
				},
				authSources: []Source{&ConnectionAPIKeySource{}, &ConnectionAppleIDSource{}, &InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: nil,
				APIKey:  &expectedAppleIDWithAPIKeyConnection,
			},
		},
		{
			name: "Connection active (API Key), inputs (Apple ID), connection not enabled",
			args: args{
				devportalConnection: &devportalservice.AppleDeveloperConnection{
					APIKeyConnection: &argAPIKeyConnection,
				},
				authSources: []Source{&InputAPIKeySource{}, &InputAppleIDSource{}},
				inputs:      argInput,
			},
			want: Credentials{
				AppleID: &expectedAppleIDWithArgInput,
				APIKey:  nil,
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
