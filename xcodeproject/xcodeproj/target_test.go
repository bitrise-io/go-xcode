package xcodeproj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarget_productPredicates(t *testing.T) {
	tests := []struct {
		name         string
		target       Target
		isExecutable bool
		isUITest     bool
		isAppClip    bool
		isTest       bool
		isNative     bool
	}{
		{
			name:         "application",
			target:       Target{kind: nativeTargetKind, productPath: "App.app", productType: "com.apple.product-type.application"},
			isExecutable: true,
			isNative:     true,
		},
		{
			name:         "app extension",
			target:       Target{kind: nativeTargetKind, productPath: "Today.appex", productType: "com.apple.product-type.app-extension"},
			isExecutable: true,
			isNative:     true,
		},
		{
			name:         "app clip",
			target:       Target{kind: nativeTargetKind, productPath: "Clip.app", productType: appClipProductType},
			isExecutable: true,
			isAppClip:    true,
			isNative:     true,
		},
		{
			name:     "unit test bundle",
			target:   Target{kind: nativeTargetKind, productPath: "Tests.xctest", productType: "com.apple.product-type.bundle.unit-test"},
			isTest:   true,
			isNative: true,
		},
		{
			name:     "ui test bundle",
			target:   Target{kind: nativeTargetKind, productPath: "UITests.xctest", productType: "com.apple.product-type.bundle.ui-testing"},
			isUITest: true,
			isTest:   true,
			isNative: true,
		},
		{
			name:     "OCTest bundle",
			target:   Target{kind: nativeTargetKind, productPath: "Legacy.octest", productType: "com.apple.product-type.bundle"},
			isTest:   true,
			isNative: true,
		},
		{
			name:   "aggregate target produces nothing",
			target: Target{kind: aggregateTargetKind},
		},
		{
			name:     "static library is not executable",
			target:   Target{kind: nativeTargetKind, productPath: "libThing.a", productType: "com.apple.product-type.library.static"},
			isNative: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isExecutable, tt.target.IsExecutableProduct(), "IsExecutableProduct")
			assert.Equal(t, tt.isUITest, tt.target.IsUITestProduct(), "IsUITestProduct")
			assert.Equal(t, tt.isAppClip, tt.target.IsAppClipProduct(), "IsAppClipProduct")
			assert.Equal(t, tt.isTest, tt.target.isTest(), "isTest")
			assert.Equal(t, tt.isNative, tt.target.isNativeTarget(), "isNativeTarget")
		})
	}
}

func TestTarget_DependsOn(t *testing.T) {
	target := Target{dependencyTargetIDs: []string{"A", "B"}}

	assert.True(t, target.DependsOn("A"))
	assert.True(t, target.DependsOn("B"))
	assert.False(t, target.DependsOn("C"))
	assert.False(t, Target{}.DependsOn("A"))
}
