package codesign

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/autocodesign/codesignasset"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient"
	"github.com/bitrise-io/go-xcode/autocodesign/keychain"
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

	ExportMethod      string
	XcodeMajorVersion int

	RegisterTestDevices bool
	SignUITests         bool
	MinProfileValidity  int
	IsVerboseLog        bool
}

type Result struct {
	XcodebuildAuthParams *devportalservice.APIKeyConnection
}

type Manager interface {
	FetchAndApplyCodesignAssets(opts Opts) (Result, error)
}

type manager struct {
	bitriseConnection      devportalservice.AppleDeveloperConnection
	devPortalClientFactory devportalclient.Factory
	certDownloader         autocodesign.CertificateProvider
	keychain               keychain.Keychain
	assetWriter            codesignasset.Writer

	projectFactory projectmanager.Factory
	project        Project

	logger log.Logger
}

func New(logger log.Logger,
	connection devportalservice.AppleDeveloperConnection,
	clientFactory devportalclient.Factory,
	certDownloader autocodesign.CertificateProvider,
	keychain keychain.Keychain,
	assetWriter codesignasset.Writer,
	projectFactory projectmanager.Factory,
) manager {
	return manager{
		bitriseConnection:      connection,
		devPortalClientFactory: clientFactory,
		certDownloader:         certDownloader,
		keychain:               keychain,
		assetWriter:            assetWriter,
		projectFactory:         projectFactory,
		logger:                 logger,
	}
}

type Project interface {
	IsSigningManagedAutomatically() (bool, error)
	Platform() (autocodesign.Platform, error)
	GetAppLayout(uiTestTargets bool) (autocodesign.AppLayout, error)
	ForceCodesignAssets(distribution autocodesign.DistributionType, codesignAssetsByDistributionType map[autocodesign.DistributionType]autocodesign.AppCodesignAssets) error
}

func (m *manager) getProject() (Project, error) {
	if m.project == nil {
		var err error
		m.project, err = m.projectFactory.Create()
		if err != nil {
			return nil, fmt.Errorf("failed to open project: %s", err)
		}
	}

	return m.project, nil
}

func (m *manager) FetchAndApplyCodesignAssets(opts Opts) (Result, error) {
	if opts.AuthType == NoAuth {
		m.logger.Println()
		m.logger.Infof("Skip downloading any Code Signing assets")

		return Result{}, nil
	}

	credentials, err := m.selectCredentials(opts.AuthType, m.bitriseConnection)
	if err != nil {
		return Result{}, err
	}

	strategy, err := m.selectCodeSigningStrategy(credentials, opts.XcodeMajorVersion)
	if err != nil {
		return Result{}, err
	}

	switch strategy {
	case noCodeSign:
		m.logger.Infof("Skip downloading any Code Signing assets")
		return Result{}, nil
	case codeSigningXcode:
		{
			m.logger.Println()
			m.logger.Infof("Xcode-managed Code Signing selected")

			m.logger.Infof("Downloading certificates from Bitrise")
			if err := m.downloadAndInstallCertificates(); err != nil {
				return Result{}, err
			}

			if opts.RegisterTestDevices && len(m.bitriseConnection.TestDevices) != 0 &&
				autocodesign.DistributionTypeRequiresDeviceList([]autocodesign.DistributionType{autocodesign.DistributionType(opts.ExportMethod)}) {
				if err := m.registerTestDevices(credentials); err != nil {
					return Result{}, err
				}
			}

			return Result{
				XcodebuildAuthParams: credentials.APIKey,
			}, nil
		}
	case codeSigningBitriseAPIKey:
		{
			m.logger.Println()
			m.logger.Infof("Bitrise Code Signing with Apple API key")
			if err := m.manageCodeSigningBitrise(credentials, opts); err != nil {
				return Result{}, err
			}

			return Result{}, nil
		}
	case codeSigningBitriseAppleID:
		{
			m.logger.Println()
			m.logger.Infof("Bitrise Code Signing with Apple ID")
			if err := m.manageCodeSigningBitrise(credentials, opts); err != nil {
				return Result{}, err
			}

			return Result{}, nil
		}
	}

	return Result{}, nil
}

func (m *manager) selectCredentials(authType AuthType, conn devportalservice.AppleDeveloperConnection) (appleauth.Credentials, error) {
	var authSource appleauth.Source

	switch authType {
	case APIKeyAuth:
		authSource = &appleauth.ConnectionAPIKeySource{}
	case AppleIDAuth:
		authSource = &appleauth.ConnectionAppleIDFastlaneSource{}
	case NoAuth:
		panic("not supported")
	default:
		panic("missing implementation")
	}

	authConfig, err := appleauth.Select(&conn, []appleauth.Source{authSource}, appleauth.Inputs{})
	if err != nil {
		if conn.APIKeyConnection == nil && conn.AppleIDConnection == nil {
			fmt.Println()
			m.logger.Warnf("%s", devportalclient.NotConnectedWarning)
		}

		return appleauth.Credentials{}, fmt.Errorf("could not configure Apple service authentication: %w", err)
	}

	if authConfig.APIKey != nil {
		m.logger.Donef("Using Apple service connection with API key.")
	} else if authConfig.AppleID != nil {
		m.logger.Donef("Using Apple service connection with Apple ID.")
	} else {
		panic("No Apple authentication credentials found.")
	}

	return authConfig, nil
}

func (m *manager) selectCodeSigningStrategy(credentials appleauth.Credentials, XcodeMajorVersion int) (codeSigningStrategy, error) {
	if credentials.AppleID != nil {
		return codeSigningBitriseAppleID, nil
	}

	if credentials.APIKey != nil {
		if XcodeMajorVersion < 13 {
			return codeSigningBitriseAPIKey, nil
		}

		project, err := m.getProject()
		if err != nil {
			return codeSigningXcode, err
		}

		managedSigning, err := project.IsSigningManagedAutomatically()
		if err != nil {
			return codeSigningXcode, err
		}

		if managedSigning {
			return codeSigningXcode, nil
		}

		return codeSigningBitriseAPIKey, nil
	}

	return noCodeSign, nil
}

func (m *manager) downloadAndInstallCertificates() error {
	certificates, err := m.certDownloader.GetCertificates()
	if err != nil {
		return fmt.Errorf("failed to download certificates: %s", err)
	}

	m.logger.Infof("Installing downloaded certificates:")
	for _, cert := range certificates {
		// Empty passphrase provided, as already parsed certificate + private key
		if err := m.keychain.InstallCertificate(cert, ""); err != nil {
			return err
		}

		m.logger.Infof("- %s (serial: %s", cert.CommonName, cert.Serial)
	}

	return nil
}

func (m *manager) registerTestDevices(credentials appleauth.Credentials) error {
	project, err := m.getProject()
	if err != nil {
		return err
	}

	platform, err := project.Platform()
	if err != nil {
		return fmt.Errorf("failed to read platform from project: %s", err)
	}

	// No Team ID required for API key client
	devPortalClient, err := m.devPortalClientFactory.Create(credentials, "")
	if err != nil {
		return err
	}

	if _, err = autocodesign.EnsureTestDevices(devPortalClient, m.bitriseConnection.TestDevices, autocodesign.Platform(platform)); err != nil {
		return fmt.Errorf("failed to ensure test devices: %w", err)
	}

	return nil
}

func (m *manager) manageCodeSigningBitrise(credentials appleauth.Credentials, opts Opts) error {
	// Analyze project
	fmt.Println()
	m.logger.Infof("Analyzing project")
	project, err := m.getProject()
	if err != nil {
		return err
	}

	appLayout, err := project.GetAppLayout(opts.SignUITests)
	if err != nil {
		return err
	}

	devPortalClient, err := m.devPortalClientFactory.Create(credentials, appLayout.TeamID)
	if err != nil {
		return err
	}

	manager := autocodesign.NewCodesignAssetManager(devPortalClient, m.certDownloader, m.assetWriter)

	// Fetch and apply codesigning assets
	distribution := autocodesign.DistributionType(opts.ExportMethod)
	testDevices := []devportalservice.TestDevice{}
	if opts.RegisterTestDevices {
		testDevices = m.bitriseConnection.TestDevices
	}
	codesignAssetsByDistributionType, err := manager.EnsureCodesignAssets(appLayout, autocodesign.CodesignAssetsOpts{
		DistributionType:       distribution,
		BitriseTestDevices:     testDevices,
		MinProfileValidityDays: opts.MinProfileValidity,
		VerboseLog:             opts.IsVerboseLog,
	})
	if err != nil {
		return err
	}

	if err := project.ForceCodesignAssets(distribution, codesignAssetsByDistributionType); err != nil {
		return fmt.Errorf("failed to force codesign settings: %s", err)
	}

	return nil
}
