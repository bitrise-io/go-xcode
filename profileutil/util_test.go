package profileutil_test

import (
	"reflect"
	"testing"

	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/fullsailor/pkcs7"
	"github.com/stretchr/testify/require"
)

func TestInstalledProvisioningProfiles(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		profileType profileutil.ProfileType
		want        []*pkcs7.PKCS7
		wantErr     bool
	}{
		{
			name:        "InstalledProvisioningProfiles() succeeds with ProfileTypeIos",
			profileType: profileutil.ProfileTypeIos,
			want:        nil, // TODO: fill in expected value
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := profileutil.InstalledProvisioningProfiles(tt.profileType)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("InstalledProvisioningProfiles() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("InstalledProvisioningProfiles() succeeded unexpectedly")
			}

			for _, gotProfile := range got {
				p, err := profileutil.NewProvisioningProfileInfo(*gotProfile)
				if err != nil {
					t.Errorf("NewProvisioningProfileInfo() failed: %v", err)
				}

				t.Logf("name: %s", p.Name) // test that Name() doesn't panic
			}

			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("InstalledProvisioningProfiles() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFindProvisioningProfile(t *testing.T) {
	type args struct {
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		want    *pkcs7.PKCS7
		want1   string
		wantErr bool
	}{
		{
			name:  "FindProvisioningProfile() succeeds with valid uuid",
			args:  args{uuid: "ff532fbb-33ce-460b-97c3-bdc6d0e2d4e0"},
			want:  nil, // TODO: fill in expected value
			want1: "",  // TODO: fill in expected valueq
			// wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := profileutil.FindProvisioningProfile(tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProvisioningProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindProvisioningProfile() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FindProvisioningProfile() got1 = %v, want %v", got1, tt.want1)
			}
			p, err := profileutil.NewProvisioningProfileInfo(*got)
			require.NoError(t, err)
			require.Equal(t, "BitriseBot-Wildcard", p.Name)
		})
	}
}
