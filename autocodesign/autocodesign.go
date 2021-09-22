package autocodesign

import (
	"math/big"

	"github.com/bitrise-io/go-xcode/devportalservice"

	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
)

// Certificate is certificate present on Apple App Store Connect API, could match a local certificate
type Certificate struct {
	Certificate certificateutil.CertificateInfoModel
	ID          string
}

// CertificateType ...
type CertificateType string

// CertificateTypes ...
const (
	Development              CertificateType = "DEVELOPMENT"
	Distribution             CertificateType = "DISTRIBUTION"
	IOSDevelopment           CertificateType = "IOS_DEVELOPMENT"
	IOSDistribution          CertificateType = "IOS_DISTRIBUTION"
	MacDistribution          CertificateType = "MAC_APP_DISTRIBUTION"
	MacInstallerDistribution CertificateType = "MAC_INSTALLER_DISTRIBUTION"
	MacDevelopment           CertificateType = "MAC_APP_DEVELOPMENT"
	DeveloperIDKext          CertificateType = "DEVELOPER_ID_KEXT"
	DeveloperIDApplication   CertificateType = "DEVELOPER_ID_APPLICATION"
)

type DevicePlatform appstoreconnect.DevicePlatform
type Device appstoreconnect.Device
type BitriseTestDevice devportalservice.TestDevice
type Profile interface {
	ID() string
	Attributes() appstoreconnect.ProfileAttributes
	CertificateIDs() (map[string]bool, error)
	DeviceIDs() (map[string]bool, error)
	BundleID() (appstoreconnect.BundleID, error)
}
type Entitlement serialized.Object
type BundleID appstoreconnect.BundleID
type ProfileType appstoreconnect.ProfileType

type DevPortalClient interface {
	QueryCertificateBySerial(*big.Int) (Certificate, error)
	QueryAllIOSCertificates() (map[CertificateType][]Certificate, error)

	ListDevices(udid string, platform DevicePlatform) ([]Device, error)
	RegisterDevice(testDevice BitriseTestDevice) (*Device, error)

	FindProfile(name string, profileType ProfileType) (Profile, error)
	DeleteProfile(id string) error
	CreateProfile(name string, profileType ProfileType, bundleID BundleID, certificateIDs []string, deviceIDs []string) (Profile, error)

	FindBundleID(bundleIDIdentifier string) (*BundleID, error)
	CheckBundleIDEntitlements(bundleID BundleID, projectEntitlements Entitlement) error
	SyncBundleID(bundleID BundleID, entitlements Entitlement) error
	CreateBundleID(bundleIDIdentifier string) (*BundleID, error)
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

// CodesignAssetManager

type CodesignAssetsOpts struct {
	DistributionType       string
	MinProfileValidityDays int
}

type CodesignAssetManager interface {
	EnsureCodesignAssets(appLayout AppLayout, opts CodesignAssetsOpts) error
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

func (m codesignAssetManager) EnsureCodesignAssets(appLayout AppLayout, opts CodesignAssetsOpts) error {

	return nil
}
