package xcodeproj

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
)

// TargetType ...
type TargetType string

// ProductType ...
type ProductType string

// TargetTypes
const (
	NativeTargetType    TargetType = "PBXNativeTarget"
	AggregateTargetType TargetType = "PBXAggregateTarget"
	LegacyTargetType    TargetType = "PBXLegacyTarget"
)

// ProductType
const (
	StaticLibraryProductType                ProductType = "com.apple.product-type.library.static"
	DynamicLibraryProductType               ProductType = "com.apple.product-type.library.dynamic"
	CommandLineToolProductType              ProductType = "com.apple.product-type.tool"
	BundleProductType                       ProductType = "com.apple.product-type.bundle"
	FrameworkProductType                    ProductType = "com.apple.product-type.framework"
	StaticFrameworkProductType              ProductType = "com.apple.product-type.framework.static"
	XCFrameworkProductType                  ProductType = "com.apple.product-type.xcframework"
	ApplicationProductType                  ProductType = "com.apple.product-type.application"
	UnitTestProductType                     ProductType = "com.apple.product-type.bundle.unit-test"
	UIUnitTestProductType                   ProductType = "com.apple.product-type.bundle.ui-testing"
	OCUnitTestBundleProductType             ProductType = "com.apple.product-type.bundle.ocunit-test"
	InAppPurchaseContentProductType         ProductType = "com.apple.product-type.in-app-purchase-content"
	AppExtensionProductType                 ProductType = "com.apple.product-type.app-extension"
	XPCServiceProductType                   ProductType = "com.apple.product-type.xpc-service"
	Watch1AppProductType                    ProductType = "com.apple.product-type.application.watchapp"
	Watch2AppProductType                    ProductType = "com.apple.product-type.application.watchapp2"
	Watch1ExtensionProductType              ProductType = "com.apple.product-type.watchkit-extension"
	Watch2ExtensionProductType              ProductType = "com.apple.product-type.watchkit2-extension"
	Watch2AppContainerProductType           ProductType = "com.apple.product-type.application.watchapp2-container"
	TVAppExtensionProductType               ProductType = "com.apple.product-type.tv-app-extension"
	TVAppBroadcastExtensionProductType      ProductType = "com.apple.product-type.tv-broadcast-extension"
	IMessageExtensionProductType            ProductType = "com.apple.product-type.app-extension.messages"
	MessagesStickerPackExtensionProductType ProductType = "com.apple.product-type.app-extension.messages-sticker-pack"
	MessagesApplicationProductType          ProductType = "com.apple.product-type.application.messages"
	AppClipProductType                      ProductType = "com.apple.product-type.application.on-demand-install-capable"
	XcodeExtensionProductType               ProductType = "com.apple.product-type.xcode-extension"
	InstrumentsPackageProductType           ProductType = "com.apple.product-type.instruments-package"
	IntentsServiceExtensionProductType      ProductType = "com.apple.product-type.app-extension.intents-service"
	MetalLibraryProductType                 ProductType = "com.apple.product-type.metal-library"
	DriverExtensionProductType              ProductType = "com.apple.product-type.driver-extension"
	SystemExtensionProductType              ProductType = "com.apple.product-type.system-extension"
)

func (p ProductType) String() string {
	return string(p)
}

func (p ProductType) IsApplicationOrApplicationExtensions() bool {
	switch p {
	case ApplicationProductType, Watch1AppProductType, Watch2AppProductType, MessagesApplicationProductType, AppClipProductType:
		return true
	case Watch2AppContainerProductType:
		return true
	case Watch1ExtensionProductType, Watch2ExtensionProductType, AppExtensionProductType, IMessageExtensionProductType, MessagesStickerPackExtensionProductType, IntentsServiceExtensionProductType:
		return true
	default:
		return false
	}
}

// Target ...
type Target struct {
	Type                   TargetType
	ID                     string
	Name                   string
	BuildConfigurationList ConfigurationList
	Dependencies           []TargetDependency
	ProductReference       ProductReference
	ProductType            string
	buildPhaseIDs          []string
}

// DependentTargets ...
func (t Target) DependentTargets() []Target {
	var targets []Target
	for _, targetDependency := range t.Dependencies {
		childTarget := targetDependency.Target
		targets = append(targets, childTarget)

		childDependentTargets := childTarget.DependentTargets()
		targets = append(targets, childDependentTargets...)
	}

	return targets
}

// DependesOn ...
func (t Target) DependesOn(targetID string) bool {
	for _, targetDependency := range t.Dependencies {
		childTarget := targetDependency.Target
		if childTarget.ID == targetID {
			return true
		}
	}
	return false
}

// DependentExecutableProductTargets ...
func (t Target) DependentExecutableProductTargets() []Target {
	var targets []Target
	for _, targetDependency := range t.Dependencies {
		childTarget := targetDependency.Target
		if !childTarget.IsExecutableProduct() {
			continue
		}

		targets = append(targets, childTarget)

		childDependentTargets := childTarget.DependentExecutableProductTargets()
		targets = append(targets, childDependentTargets...)
	}

	return targets
}

// IsAppProduct ...
func (t Target) IsAppProduct() bool {
	return filepath.Ext(t.ProductReference.Path) == ".app"
}

// IsAppExtensionProduct ...
func (t Target) IsAppExtensionProduct() bool {
	return filepath.Ext(t.ProductReference.Path) == ".appex"
}

// IsExecutableProduct ...
func (t Target) IsExecutableProduct() bool {
	return t.IsAppProduct() || t.IsAppExtensionProduct()
}

// IsTest identifies test targets
// Based on https://github.com/CocoaPods/Xcodeproj/blob/907c81763a7660978fda93b2f38f05de0cbb51ad/lib/xcodeproj/project/object/native_target.rb#L470
func (t Target) IsTest() bool {
	return t.IsTestProduct() ||
		t.IsUITestProduct() ||
		t.ProductType == "com.apple.product-type.bundle" // OCTest bundle
}

// IsTestProduct ...
func (t Target) IsTestProduct() bool {
	return filepath.Ext(t.ProductType) == ".unit-test"
}

// IsUITestProduct ...
func (t Target) IsUITestProduct() bool {
	return filepath.Ext(t.ProductType) == ".ui-testing"
}

func (t Target) isAppClipProduct() bool {
	return t.ProductType == AppClipProductType.String()
}

// CanExportAppClip ...
func (t Target) CanExportAppClip() bool {
	deps := t.DependentTargets()
	for _, target := range deps {
		if target.isAppClipProduct() {
			return true
		}
	}

	return false
}

func parseTarget(id string, objects serialized.Object) (Target, error) {
	rawTarget, err := objects.Object(id)
	if err != nil {
		return Target{}, err
	}

	isa, err := rawTarget.String("isa")
	if err != nil {
		return Target{}, err
	}

	var targetType TargetType
	switch isa {
	case "PBXNativeTarget":
		targetType = NativeTargetType
	case "PBXAggregateTarget":
		targetType = AggregateTargetType
	case "PBXLegacyTarget":
		targetType = LegacyTargetType
	default:
		return Target{}, fmt.Errorf("unknown target type: %s", isa)
	}

	name, err := rawTarget.String("name")
	if err != nil {
		return Target{}, err
	}

	productType, err := rawTarget.String("productType")
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		return Target{}, err
	}

	buildConfigurationListID, err := rawTarget.String("buildConfigurationList")
	if err != nil {
		return Target{}, err
	}

	buildConfigurationList, err := parseConfigurationList(buildConfigurationListID, objects)
	if err != nil {
		return Target{}, err
	}

	var dependencies []TargetDependency

	// Filter for applications and extensions product type only
	if ProductType(productType).IsApplicationOrApplicationExtensions() {
		dependencyIDs, err := rawTarget.StringSlice("dependencies")
		if err != nil {
			return Target{}, err
		}

		for _, dependencyID := range dependencyIDs {
			dependency, err := parseTargetDependency(dependencyID, objects)
			if err != nil {
				// KeyNotFoundError can be only raised if the 'target' property not found on the raw target dependency object
				// we only care about target dependency, which points to a target
				if serialized.IsKeyNotFoundError(err) {
					continue
				} else {
					return Target{}, err
				}
			}

			dependencies = append(dependencies, dependency)
		}
	}

	var productReference ProductReference
	productReferenceID, err := rawTarget.String("productReference")
	if err != nil {
		if !serialized.IsKeyNotFoundError(err) {
			return Target{}, err
		}
	} else {
		productReference, err = parseProductReference(productReferenceID, objects)
		if err != nil {
			return Target{}, err
		}
	}

	buildPhaseIDs, err := rawTarget.StringSlice("buildPhases")
	if err != nil {
		return Target{}, err
	}

	return Target{
		Type:                   targetType,
		ID:                     id,
		Name:                   name,
		BuildConfigurationList: buildConfigurationList,
		Dependencies:           dependencies,
		ProductReference:       productReference,
		ProductType:            productType,
		buildPhaseIDs:          buildPhaseIDs,
	}, nil
}
