package xcodeproj

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

// Target isa values.
const (
	nativeTargetISA    = "PBXNativeTarget"
	aggregateTargetISA = "PBXAggregateTarget"
	legacyTargetISA    = "PBXLegacyTarget"
)

const appClipProductType = "com.apple.product-type.application.on-demand-install-capable"

// Target is a build target of the project: an app, an app extension, a test bundle, a library.
type Target struct {
	ID   string
	Name string

	// BuildConfigurations are the target's own configurations, in project file order.
	BuildConfigurations []BuildConfiguration
	// DefaultConfigurationName is empty when the target's configuration list does not name one.
	DefaultConfigurationName string

	isa                 string
	productType         string
	productPath         string
	dependencyTargetIDs []string
	buildPhaseIDs       []string
}

// DependsOn reports whether the target declares a direct dependency on the target with the given
// ID. It does not consider transitive dependencies; use XcodeProj.DependentTargetsOfTarget for
// those.
func (t Target) DependsOn(targetID string) bool {
	for _, id := range t.dependencyTargetIDs {
		if id == targetID {
			return true
		}
	}
	return false
}

// IsExecutableProduct reports whether the target produces an .app or an .appex bundle, meaning it
// is a product that gets code signed and embedded in an archive.
func (t Target) IsExecutableProduct() bool {
	return t.isAppProduct() || t.isAppExtensionProduct()
}

// IsUITestProduct reports whether the target is a UI test bundle.
func (t Target) IsUITestProduct() bool {
	return filepath.Ext(t.productType) == ".ui-testing"
}

// IsAppClipProduct reports whether the target produces an App Clip.
func (t Target) IsAppClipProduct() bool {
	return t.productType == appClipProductType
}

func (t Target) isAppProduct() bool {
	return filepath.Ext(t.productPath) == ".app"
}

func (t Target) isAppExtensionProduct() bool {
	return filepath.Ext(t.productPath) == ".appex"
}

func (t Target) isNativeTarget() bool {
	return t.isa == nativeTargetISA
}

// isTest identifies any flavour of test target.
// Based on https://github.com/CocoaPods/Xcodeproj/blob/907c81763a7660978fda93b2f38f05de0cbb51ad/lib/xcodeproj/project/object/native_target.rb#L470
func (t Target) isTest() bool {
	return t.isTestProduct() ||
		t.IsUITestProduct() ||
		t.productType == "com.apple.product-type.bundle" // OCTest bundle
}

func (t Target) isTestProduct() bool {
	return filepath.Ext(t.productType) == ".unit-test"
}

func parseTarget(id string, objects serialized.Object) (Target, error) {
	rawTarget, ok := objects.Object(id)
	if !ok {
		return Target{}, fmt.Errorf("target %s not found", id)
	}

	isa, ok := rawTarget.String("isa")
	if !ok {
		return Target{}, fmt.Errorf("target %s has no isa", id)
	}

	switch isa {
	case nativeTargetISA, aggregateTargetISA, legacyTargetISA:
	default:
		return Target{}, fmt.Errorf("unknown target type: %s", isa)
	}

	name, ok := rawTarget.String("name")
	if !ok {
		return Target{}, fmt.Errorf("target %s has no name", id)
	}

	// Only native targets have a product type.
	productType, _ := rawTarget.String("productType")

	buildConfigurationListID, ok := rawTarget.String("buildConfigurationList")
	if !ok {
		return Target{}, fmt.Errorf("target %s has no buildConfigurationList", id)
	}

	buildConfigurations, defaultConfigurationName, err := parseConfigurationList(buildConfigurationListID, objects)
	if err != nil {
		return Target{}, fmt.Errorf("target %s: %w", id, err)
	}

	// dependencies is optional: a target with none omits the key entirely.
	dependencyIDs, _ := rawTarget.StringSlice("dependencies")

	var dependencyTargetIDs []string
	for _, dependencyID := range dependencyIDs {
		// A PBXTargetDependency can reference a target either directly or through a proxy. Only
		// the direct form carries a "target" key, and only that form interests us.
		targetID, ok := targetIDOfDependency(dependencyID, objects)
		if !ok {
			continue
		}
		dependencyTargetIDs = append(dependencyTargetIDs, targetID)
	}

	// productReference is absent for targets that build no product, such as aggregate targets.
	var productPath string
	if productReferenceID, ok := rawTarget.String("productReference"); ok {
		productReference, ok := objects.Object(productReferenceID)
		if !ok {
			return Target{}, fmt.Errorf("target %s references missing product %s", id, productReferenceID)
		}
		productPath, _ = productReference.String("path")
	}

	// buildPhases is optional in the same way as dependencies.
	buildPhaseIDs, _ := rawTarget.StringSlice("buildPhases")

	return Target{
		ID:                       id,
		Name:                     name,
		BuildConfigurations:      buildConfigurations,
		DefaultConfigurationName: defaultConfigurationName,
		isa:                      isa,
		productType:              productType,
		productPath:              productPath,
		dependencyTargetIDs:      dependencyTargetIDs,
		buildPhaseIDs:            buildPhaseIDs,
	}, nil
}

func targetIDOfDependency(dependencyID string, objects serialized.Object) (string, bool) {
	rawDependency, ok := objects.Object(dependencyID)
	if !ok {
		return "", false
	}
	return rawDependency.String("target")
}
