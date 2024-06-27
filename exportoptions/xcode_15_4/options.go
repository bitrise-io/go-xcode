package xcode_15_4

type Destination string

const (
	DestinationExport Destination = "export"
	DestinationUpload Destination = "upload"
)

type ICloudContainerEnvironment string

const (
	ICloudContainerEnvironmentDevelopment ICloudContainerEnvironment = "Development"
	ICloudContainerEnvironmentProduction  ICloudContainerEnvironment = "Production"
)

type Method string

const (
	MethodAppStoreConnect Method = "app-store-connect"
	MethodReleaseTesting  Method = "release-testing"
	MethodEnterprise      Method = "enterprise"
	MethodDebugging       Method = "debugging"
)

type SigningStyle string

const (
	SigningStyleManual    SigningStyle = "manual"
	SigningStyleAutomatic SigningStyle = "automatic"
)

type Thinning string

const (
	ThinningNone               Thinning = "<none>"
	ThinningThinForAllVariants Thinning = "<thin-for-all-variants>"
	// or a specific device (e.g. "iPhone7,1"), e.g.: Thinning("iPhone7,1")
)
