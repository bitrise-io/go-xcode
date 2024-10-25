package main

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-xcode/v2/autocodesign"
	"github.com/bitrise-io/go-xcode/v2/codesign"
)

// Config holds the step inputs
type Config struct {
	Distribution  string `env:"distribution_method,opt[development,app-store,ad-hoc,enterprise]"`
	ProjectPath   string `env:"project_path,dir"`
	Scheme        string `env:"scheme,required"`
	Configuration string `env:"configuration"`

	BitriseConnection string `env:"apple_service_connection,opt[api-key,apple-id]"`

	RegisterTestDevices bool   `env:"register_test_devices,opt[yes,no]"`
	MinProfileDaysValid int    `env:"min_profile_validity,required"`
	SignUITestTargets   bool   `env:"sign_uitest_targets,opt[yes,no]"`
	TeamID              string `env:"apple_team_id"`

	CertificateURLList        string          `env:"certificate_url_list,required"`
	CertificatePassphraseList stepconf.Secret `env:"passphrase_list"`
	KeychainPath              string          `env:"keychain_path,required"`
	KeychainPassword          stepconf.Secret `env:"keychain_password,required"`
	BuildAPIToken             string          `env:"build_api_token"`
	BuildURL                  string          `env:"build_url"`
	APIKeyPath                stepconf.Secret `env:"api_key_path"`
	APIKeyID                  string          `env:"api_key_id"`
	APIKeyIssuerID            string          `env:"api_key_issuer_id"`

	VerboseLog bool `env:"verbose_log,opt[no,yes]"`
}

// DistributionType ...
func (c Config) DistributionType() autocodesign.DistributionType {
	return autocodesign.DistributionType(c.Distribution)
}

func parseAuthType(bitriseConnection string) (codesign.AuthType, error) {
	switch bitriseConnection {
	case "api-key":
		return codesign.APIKeyAuth, nil
	case "apple-id":
		return codesign.AppleIDAuth, nil
	default:
		return 0, fmt.Errorf("invalid connection input: %s", bitriseConnection)
	}
}
