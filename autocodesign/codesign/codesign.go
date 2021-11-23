package codesign

import (
	"errors"
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

// AuthType ...
type AuthType int

const (
	// NoAuth ...
	NoAuth AuthType = iota
	// APIKeyAuth ...
	APIKeyAuth
	// AppleIDAuth ...
	AppleIDAuth
	// AnyAuth ...
	AnyAuth
)

type codeSigningStrategy int

const (
	codeSigningXcode codeSigningStrategy = iota
	codeSigningBitriseAPIKey
	codeSigningBitriseAppleID
)

// Opts ...
type Opts struct {
	AuthType                  AuthType
	IsXcodeCodeSigningEnabled bool

	ExportMethod      autocodesign.DistributionType
	XcodeMajorVersion int

	RegisterTestDevices bool
	SignUITests         bool
	MinProfileValidity  int
	IsVerboseLog        bool
}

// Result ...
type Result struct {
	XcodebuildAuthParams *devportalservice.APIKeyConnection
}

// Manager ...
type Manager struct {
	appleAuthCredentials   appleauth.Credentials
	bitriseConnection      *devportalservice.AppleDeveloperConnection
	devPortalClientFactory devportalclient.Factory
	certDownloader         autocodesign.CertificateProvider
	keychain               keychain.Keychain
	assetWriter            codesignasset.Writer

	projectFactory projectmanager.Factory
	project        Project

	logger log.Logger
}

// New ...
func New(logger log.Logger,
	appleAuth appleauth.Credentials,
	connection *devportalservice.AppleDeveloperConnection,
	clientFactory devportalclient.Factory,
	certDownloader autocodesign.CertificateProvider,
	keychain keychain.Keychain,
	assetWriter codesignasset.Writer,
	projectFactory projectmanager.Factory,
) Manager {
	return Manager{
		appleAuthCredentials:   appleAuth,
		bitriseConnection:      connection,
		devPortalClientFactory: clientFactory,
		certDownloader:         certDownloader,
		keychain:               keychain,
		assetWriter:            assetWriter,
		projectFactory:         projectFactory,
		logger:                 logger,
	}
}

// Project ...
type Project interface {
	IsSigningManagedAutomatically() (bool, error)
	Platform() (autocodesign.Platform, error)
	GetAppLayout(uiTestTargets bool) (autocodesign.AppLayout, error)
	ForceCodesignAssets(distribution autocodesign.DistributionType, codesignAssetsByDistributionType map[autocodesign.DistributionType]autocodesign.AppCodesignAssets) error
}

func (m *Manager) getProject() (Project, error) {
	if m.project == nil {
		var err error
		m.project, err = m.projectFactory.Create()
		if err != nil {
			return nil, fmt.Errorf("failed to open project: %s", err)
		}
	}

	return m.project, nil
}

// PrepareCodesigning ...
func (m *Manager) PrepareCodesigning(opts Opts) (Result, error) {
	if opts.AuthType == NoAuth {
		m.logger.Println()
		m.logger.Infof("Skip downloading any Code Signing assets")

		return Result{}, nil
	}

	strategy, reason, err := m.selectCodeSigningStrategy(m.appleAuthCredentials, opts.IsXcodeCodeSigningEnabled, opts.XcodeMajorVersion)
	if err != nil {
		m.logger.Warnf("%s", err)
	}

	switch strategy {
	case codeSigningXcode:
		{
			m.logger.Println()
			m.logger.Infof("Preparing for Xcode-managed code-signing")
			m.logger.Printf(reason)
			m.logger.Println()
			m.logger.Infof("Downloading certificates from Bitrise")
			if err := m.downloadAndInstallCertificates(); err != nil {
				return Result{}, err
			}

			if opts.RegisterTestDevices && m.bitriseConnection != nil && len(m.bitriseConnection.TestDevices) != 0 &&
				autocodesign.DistributionTypeRequiresDeviceList([]autocodesign.DistributionType{opts.ExportMethod}) {
				if err := m.registerTestDevices(m.appleAuthCredentials, m.bitriseConnection.TestDevices); err != nil {
					return Result{}, err
				}
			}

			return Result{
				XcodebuildAuthParams: m.appleAuthCredentials.APIKey,
			}, nil
		}
	case codeSigningBitriseAPIKey:
		{
			m.logger.Println()
			m.logger.Infof("Bitrise-managed code-signing with Apple API key")
			log.Printf(reason)
			if err := m.manageCodeSigningBitrise(m.appleAuthCredentials, opts); err != nil {
				return Result{}, err
			}

			return Result{}, nil
		}
	case codeSigningBitriseAppleID:
		{
			m.logger.Println()
			m.logger.Infof("Bitrise-managed code-signing with Apple ID")
			m.logger.Printf(reason)
			if err := m.manageCodeSigningBitrise(m.appleAuthCredentials, opts); err != nil {
				return Result{}, err
			}

			return Result{}, nil
		}
	}

	return Result{}, nil
}

func SelectConnectionCredentials(authType AuthType, conn *devportalservice.AppleDeveloperConnection, logger log.Logger) (appleauth.Credentials, error) {
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

	authConfig, err := appleauth.Select(conn, []appleauth.Source{authSource}, appleauth.Inputs{})
	if err != nil {
		if conn != nil && conn.APIKeyConnection == nil && conn.AppleIDConnection == nil {
			fmt.Println()
			logger.Warnf("%s", devportalclient.NotConnectedWarning)
		}

		if errors.Is(err, &appleauth.MissingAuthConfigError{}) {
			if authType == AppleIDAuth {
				return appleauth.Credentials{}, fmt.Errorf("Apple ID authentication is selected in Step inputs, but Bitrise Apple service connection is unset")
			}

			return appleauth.Credentials{}, fmt.Errorf("API key authentication is selected in Step inputs, but Bitrise Apple service connection is unset")
		}

		return appleauth.Credentials{}, fmt.Errorf("could not select Apple authentication credentials: %w", err)
	}

	if authConfig.APIKey != nil {
		authConfig.AppleID = nil

		logger.Donef("Using Apple service connection with API key.")
	} else if authConfig.AppleID != nil {
		logger.Donef("Using Apple service connection with Apple ID.")
	} else {
		panic("No Apple authentication credentials found.")
	}

	return authConfig, nil
}

func (m *Manager) selectCodeSigningStrategy(credentials appleauth.Credentials, IsXcodeCodeSigningEnabled bool, XcodeMajorVersion int) (codeSigningStrategy, string, error) {
	const manualProfilesReason = "Using Bitrise-managed code-signing, as Xcode-managed code-signing requires automatically managed provisioning profiles"

	if credentials.AppleID != nil {
		return codeSigningBitriseAppleID, "Using Bitrise-managed code-signing, as Apple ID is not supported by Xcode-managed code-signing.", nil
	}

	if !IsXcodeCodeSigningEnabled {
		return codeSigningBitriseAPIKey, "", nil
	}

	if credentials.APIKey == nil {
		panic("No Apple authentication credentials found.")
	}

	if XcodeMajorVersion < 13 {
		return codeSigningBitriseAPIKey, "Using Bitrise-managed code-signing, as Xcode-managed code-signing requires at least Xcode 13.", nil
	}

	project, err := m.getProject()
	if err != nil {
		return codeSigningXcode, "Using Xcode-managed code-signing, as failed to parse project.", err
	}

	automaticProfiles, err := project.IsSigningManagedAutomatically()
	if err != nil {
		return codeSigningBitriseAPIKey, manualProfilesReason, err
	}

	if automaticProfiles {
		return codeSigningXcode, "Using Xcode-managed code-signing, as project uses automatically managed provisioning profiles.", nil
	}

	return codeSigningBitriseAPIKey, manualProfilesReason, nil
}

func (m *Manager) downloadAndInstallCertificates() error {
	certificates, err := m.certDownloader.GetCertificates()
	if err != nil {
		return fmt.Errorf("failed to download certificates: %s", err)
	}

	m.logger.Infof("Installing downloaded certificates:")
	for _, cert := range certificates {
		m.logger.Printf("- %s", cert)
		// Empty passphrase provided, as already parsed certificate + private key
		if err := m.keychain.InstallCertificate(cert, ""); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) registerTestDevices(credentials appleauth.Credentials, devices []devportalservice.TestDevice) error {
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

	if _, err = autocodesign.EnsureTestDevices(devPortalClient, devices, autocodesign.Platform(platform)); err != nil {
		return fmt.Errorf("failed to ensure test devices: %w", err)
	}

	return nil
}

func (m *Manager) manageCodeSigningBitrise(credentials appleauth.Credentials, opts Opts) error {
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
	if opts.RegisterTestDevices && m.bitriseConnection != nil {
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
