package autocodesign

import (
	cs "github.com/bitrise-io/go-xcode/autocodesign/codesignmodels"
	"github.com/bitrise-io/go-xcode/autocodesign/devportalclient/appstoreconnect"
)

// CertificateTypeByDistribution ...
var CertificateTypeByDistribution = map[cs.DistributionType]appstoreconnect.CertificateType{
	cs.Development: appstoreconnect.IOSDevelopment,
	cs.AppStore:    appstoreconnect.IOSDistribution,
	cs.AdHoc:       appstoreconnect.IOSDistribution,
	cs.Enterprise:  appstoreconnect.IOSDistribution,
}

// ProfileTypeToPlatform ...
var ProfileTypeToPlatform = map[appstoreconnect.ProfileType]cs.Platform{
	appstoreconnect.IOSAppDevelopment: cs.IOS,
	appstoreconnect.IOSAppStore:       cs.IOS,
	appstoreconnect.IOSAppAdHoc:       cs.IOS,
	appstoreconnect.IOSAppInHouse:     cs.IOS,

	appstoreconnect.TvOSAppDevelopment: cs.TVOS,
	appstoreconnect.TvOSAppStore:       cs.TVOS,
	appstoreconnect.TvOSAppAdHoc:       cs.TVOS,
	appstoreconnect.TvOSAppInHouse:     cs.TVOS,
}

// ProfileTypeToDistribution ...
var ProfileTypeToDistribution = map[appstoreconnect.ProfileType]cs.DistributionType{
	appstoreconnect.IOSAppDevelopment: cs.Development,
	appstoreconnect.IOSAppStore:       cs.AppStore,
	appstoreconnect.IOSAppAdHoc:       cs.AdHoc,
	appstoreconnect.IOSAppInHouse:     cs.Enterprise,

	appstoreconnect.TvOSAppDevelopment: cs.Development,
	appstoreconnect.TvOSAppStore:       cs.AppStore,
	appstoreconnect.TvOSAppAdHoc:       cs.AdHoc,
	appstoreconnect.TvOSAppInHouse:     cs.Enterprise,
}

// PlatformToProfileTypeByDistribution ...
var PlatformToProfileTypeByDistribution = map[cs.Platform]map[cs.DistributionType]appstoreconnect.ProfileType{
	cs.IOS: {
		cs.Development: appstoreconnect.IOSAppDevelopment,
		cs.AppStore:    appstoreconnect.IOSAppStore,
		cs.AdHoc:       appstoreconnect.IOSAppAdHoc,
		cs.Enterprise:  appstoreconnect.IOSAppInHouse,
	},
	cs.TVOS: {
		cs.Development: appstoreconnect.TvOSAppDevelopment,
		cs.AppStore:    appstoreconnect.TvOSAppStore,
		cs.AdHoc:       appstoreconnect.TvOSAppAdHoc,
		cs.Enterprise:  appstoreconnect.TvOSAppInHouse,
	},
}
