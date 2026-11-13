package xcodeproj

import (
	"path/filepath"
	"slices"
)

// targetKind is what builds a target. The zero value is no kind.
type targetKind int

const (
	nativeTargetKind targetKind = iota + 1
	aggregateTargetKind
	externalBuildSystemTargetKind
)

const appClipProductType = "com.apple.product-type.application.on-demand-install-capable"

type Target struct {
	ID   string
	Name string

	BuildConfigurations      []BuildConfiguration
	DefaultConfigurationName string

	kind        targetKind
	productType string
	// productPath is the path of the product file; only its base name, such as App.app, is used.
	productPath         string
	dependencyTargetIDs []string
}

// DependsOn reports whether the target directly depends on the target with the given ID.
func (t Target) DependsOn(targetID string) bool {
	return slices.Contains(t.dependencyTargetIDs, targetID)
}

// IsExecutableProduct reports whether the target produces an .app or .appex bundle.
func (t Target) IsExecutableProduct() bool {
	return t.isAppProduct() || t.isAppExtensionProduct()
}

func (t Target) IsUITestProduct() bool {
	return filepath.Ext(t.productType) == ".ui-testing"
}

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
	return t.kind == nativeTargetKind
}

// Based on https://github.com/CocoaPods/Xcodeproj/blob/907c81763a7660978fda93b2f38f05de0cbb51ad/lib/xcodeproj/project/object/native_target.rb#L470
func (t Target) isTest() bool {
	return t.isTestProduct() ||
		t.IsUITestProduct() ||
		t.productType == "com.apple.product-type.bundle" // OCTest bundle
}

func (t Target) isTestProduct() bool {
	return filepath.Ext(t.productType) == ".unit-test"
}
