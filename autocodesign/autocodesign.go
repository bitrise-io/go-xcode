package autocodesign

import (
	"fmt"
	"math/big"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/keychain"
)

// Certificate is certificate present on Apple App Store Connect API, could match a local certificate
type Certificate struct {
	Certificate certificateutil.CertificateInfoModel
	ID          string
}

type Profile interface {
	ID() string
	Attributes() appstoreconnect.ProfileAttributes
	CertificateIDs() (map[string]bool, error)
	DeviceIDs() (map[string]bool, error)
	BundleID() (appstoreconnect.BundleID, error)
}
type Entitlement serialized.Object

type DevPortalClient interface {
	QueryCertificateBySerial(*big.Int) (Certificate, error)
	QueryAllIOSCertificates() (map[appstoreconnect.CertificateType][]Certificate, error)

	ListDevices(udid string, platform appstoreconnect.DevicePlatform) ([]appstoreconnect.Device, error)
	RegisterDevice(testDevice devportalservice.TestDevice) (*appstoreconnect.Device, error)

	FindProfile(name string, profileType appstoreconnect.ProfileType) (Profile, error)
	DeleteProfile(id string) error
	CreateProfile(name string, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (Profile, error)

	FindBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error)
	CheckBundleIDEntitlements(bundleID appstoreconnect.BundleID, projectEntitlements Entitlement) error
	SyncBundleID(bundleID appstoreconnect.BundleID, entitlements Entitlement) error
	CreateBundleID(bundleIDIdentifier string) (*appstoreconnect.BundleID, error)
}

type AppLayout struct {
	TeamID                                 string
	Platform                               string
	ArchivableTargetBundleIDToEntitlements map[string]serialized.Object
	UITestTargetBundleIDs                  []string
}

type CertificateProvider interface {
	GetCertificates() ([]certificateutil.CertificateInfoModel, error)
}

// DistributionType ...
type DistributionType string

// DistributionTypes ...
var (
	Development DistributionType = "development"
	AppStore    DistributionType = "app-store"
	AdHoc       DistributionType = "ad-hoc"
	Enterprise  DistributionType = "enterprise"
)

// CodesignAssetManager
type CodesignAssetsOpts struct {
	DistributionType       DistributionType
	MinProfileValidityDays int
	Keychain               keychain.Keychain
	VerboseLog             bool
}

// AppCodesignAssets ...
type AppCodesignAssets struct {
	ArchivableTargetProfilesByBundleID map[string]Profile
	UITestTargetProfilesByBundleID     map[string]Profile
	Certificate                        certificateutil.CertificateInfoModel
}

type CodesignAssetManager interface {
	EnsureCodesignAssets(appLayout AppLayout, opts CodesignAssetsOpts) (map[DistributionType]AppCodesignAssets, error)
}

type codesignAssetManager struct {
	devPortalClient     DevPortalClient
	certificateProvider CertificateProvider
}

func NewCodesignAssetManager(devPortalClient DevPortalClient, certificateProvider CertificateProvider) CodesignAssetManager {
	return codesignAssetManager{
		devPortalClient:     devPortalClient,
		certificateProvider: certificateProvider,
	}
}

func (m codesignAssetManager) EnsureCodesignAssets(appLayout AppLayout, opts CodesignAssetsOpts) (map[DistributionType]AppCodesignAssets, error) {
	fmt.Println()
	log.Infof("Downloading certificates")

	certs, err := m.certificateProvider.GetCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to download certificates: %s", err)
	}
	log.Printf("%d certificates downloaded:", len(certs))
	for _, cert := range certs {
		log.Printf("- %s", cert.CommonName)
	}

	signUITestTargets := len(appLayout.UITestTargetBundleIDs) > 0
	certsByType, distrTypes, err := selectCertificatesAndDistributionTypes(
		m.devPortalClient,
		certs,
		opts.DistributionType,
		appLayout.TeamID,
		signUITestTargets,
		opts.VerboseLog,
	)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// Ensure devices
	var devPortalDeviceIDs []string
	if distributionTypeRequiresDeviceList(distrTypes) {
		var err error
		devPortalDeviceIDs, err = ensureTestDevices(m.devportalClient.DeviceClient, m.testDevices, project.Platform)
		if err != nil {
			return nil, fmt.Errorf("Failed to ensure test devices: %s", err)
		}
	}

	// Ensure Profiles
	codesignAssetsByDistributionType, err := ensureProfiles(m.devportalClient.ProfileClient, distrTypes, certsByType, project, devPortalDeviceIDs, minProfileDaysValid)
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure profiles: %s", err)
	}

	// Install certificates and profiles
	if err := installCodesigningFiles(codesignAssetsByDistributionType, keychain); err != nil {
		return nil, fmt.Errorf("Failed to install codesigning files: %s", err)
	}

	return codesignAssetsByDistributionType, nil
}
