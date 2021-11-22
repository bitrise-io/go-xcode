package codesign

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/autocodesign/certdownloader"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient"
	"github.com/bitrise-io/go-xcode/autocodesign/projectmanager"
	"github.com/bitrise-io/go-xcode/devportalservice"
)

type AuthType int

const (
	NoAuth AuthType = iota
	APIKeyAuth
	AppleIDAuth
)

type codeSigningStrategy int

const (
	noCodeSign codeSigningStrategy = iota
	codeSigningXcode
	codeSigningBitriseAPIKey
	codeSigningBitriseAppleID
)

type Opts struct {
	AuthType                  AuthType
	IsXcodeCodeSigningEnabled bool

	ProjectPath       string
	Scheme            string
	Configuration     string
	ExportMethod      string
	XcodeMajorVersion int

	CertificatesAndPassphrases []certdownloader.CertificateAndPassphrase

	AppleServiceConnection devportalservice.AppleDeveloperConnection
	RegisterTestDevices    bool
	SignUITests            bool
	KeychainPath           string
	KeychainPassword       stepconf.Secret
}

type Result struct {
	XcodebuildAuthParams *devportalservice.APIKeyConnection
}

type Manager interface {
	FetchAndApplyCodesignAssets(opts Opts) (Result, error)
}

type manager struct{}

type defaultManagerFactory struct {
	projectHelperFactory projectmanager.ProjectHelperFactory
	projectFactory       projectmanager.Factory
}

func NewFactory() defaultManagerFactory {
	return defaultManagerFactory{}
}

func (m *manager) FetchAndApplyCodesignAssets(opts Opts) (Result, error) {

	return Result{}, nil
}

func selectCredentials(authType AuthType, teamID string, conn devportalservice.AppleDeveloperConnection) (*appleauth.Credentials, error) {
	var authSource appleauth.Source

	switch authType {
	case NoAuth:
		return nil, nil
	case APIKeyAuth:
		authSource = &appleauth.ConnectionAPIKeySource{}
	case AppleIDAuth:
		authSource = &appleauth.ConnectionAppleIDFastlaneSource{}
	default:
		panic("missing implementation")
	}

	authConfig, err := appleauth.Select(&conn, []appleauth.Source{authSource}, appleauth.Inputs{})
	if err != nil {
		if conn.APIKeyConnection == nil && conn.AppleIDConnection == nil {
			fmt.Println()
			log.Warnf("%s", devportalclient.NotConnectedWarning)
		}

		return nil, fmt.Errorf("could not configure Apple service authentication: %v", err)
	}

	if authConfig.APIKey != nil {
		log.Donef("Using Apple service connection with API key.")
	} else if authConfig.AppleID != nil {
		log.Donef("Using Apple service connection with Apple ID.")
	} else {
		panic("No Apple authentication credentials found.")
	}

	return &authConfig, nil
}
