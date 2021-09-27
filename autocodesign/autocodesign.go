package autocodesign

import (
	"fmt"
	"math/big"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
)

// Profile ...
type Profile interface {
	ID() string
	Attributes() appstoreconnect.ProfileAttributes
	CertificateIDs() (map[string]bool, error)
	DeviceIDs() (map[string]bool, error)
	BundleID() (appstoreconnect.BundleID, error)
	Entitlements() (serialized.Object, error)
}

// AppCodesignAssets ...
type AppCodesignAssets struct {
	ArchivableTargetProfilesByBundleID map[string]Profile
	UITestTargetProfilesByBundleID     map[string]Profile
	Certificate                        certificateutil.CertificateInfoModel
}

// Platform ...
type Platform string

// Const
const (
	IOS   Platform = "iOS"
	TVOS  Platform = "tvOS"
	MacOS Platform = "macOS"
)

// DistributionType ...
type DistributionType string

// DistributionTypes ...
var (
	Development DistributionType = "development"
	AppStore    DistributionType = "app-store"
	AdHoc       DistributionType = "ad-hoc"
	Enterprise  DistributionType = "enterprise"
)

// Entitlement ...
type Entitlement serialized.Object

// Certificate is certificate present on Apple App Store Connect API, could match a local certificate
type Certificate struct {
	Certificate certificateutil.CertificateInfoModel
	ID          string
}

// DevPortalClient ...
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
	CreateBundleID(bundleIDIdentifier, appIDName string) (*appstoreconnect.BundleID, error)
}

// AssetWriter ...
type AssetWriter interface {
	Write(codesignAssetsByDistributionType map[DistributionType]AppCodesignAssets) error
}

// AppLayout ...
type AppLayout struct {
	TeamID                                 string
	Platform                               Platform
	ArchivableTargetBundleIDToEntitlements map[string]serialized.Object
	UITestTargetBundleIDs                  []string
}

// CertificateProvider ...
type CertificateProvider interface {
	GetCertificates() ([]certificateutil.CertificateInfoModel, error)
}

// CodesignAssetsOpts ...
type CodesignAssetsOpts struct {
	DistributionType       DistributionType
	BitriseTestDevices     []devportalservice.TestDevice
	MinProfileValidityDays int
	VerboseLog             bool
}

// CodesignAssetManager ...
type CodesignAssetManager interface {
	EnsureCodesignAssets(appLayout AppLayout, opts CodesignAssetsOpts) (map[DistributionType]AppCodesignAssets, error)
}

type codesignAssetManager struct {
	devPortalClient     DevPortalClient
	certificateProvider CertificateProvider
	assetWriter         AssetWriter
}

// NewCodesignAssetManager ...
func NewCodesignAssetManager(devPortalClient DevPortalClient, certificateProvider CertificateProvider, assetWriter AssetWriter) CodesignAssetManager {
	return codesignAssetManager{
		devPortalClient:     devPortalClient,
		certificateProvider: certificateProvider,
		assetWriter:         assetWriter,
	}
}

// EnsureCodesignAssets ...
func (m codesignAssetManager) EnsureCodesignAssets(appLayout AppLayout, opts CodesignAssetsOpts) (map[DistributionType]AppCodesignAssets, error) {
	fmt.Println()
	log.Infof("Downloading certificates")

	certs, err := m.certificateProvider.GetCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to download certificates: %w", err)
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
		return nil, fmt.Errorf("%w", err)
	}

	var devPortalDeviceIDs []string
	if distributionTypeRequiresDeviceList(distrTypes) {
		var err error
		devPortalDeviceIDs, err = ensureTestDevices(m.devPortalClient, opts.BitriseTestDevices, appLayout.Platform)
		if err != nil {
			return nil, fmt.Errorf("Failed to ensure test devices: %w", err)
		}
	}

	// Ensure Profiles
	codesignAssetsByDistributionType, err := ensureProfiles(m.devPortalClient, distrTypes, certsByType, appLayout, devPortalDeviceIDs, opts.MinProfileValidityDays)
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure profiles: %w", err)
	}

	// Install certificates and profiles
	fmt.Println()
	log.Infof("Install certificates and profiles")
	if err := m.assetWriter.Write(codesignAssetsByDistributionType); err != nil {
		return nil, fmt.Errorf("Failed to install codesigning files: %s", err)
	}

	return codesignAssetsByDistributionType, nil
}
