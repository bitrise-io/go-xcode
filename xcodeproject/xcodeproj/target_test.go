package xcodeproj

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			target:       Target{isa: nativeTargetISA, productPath: "App.app", productType: "com.apple.product-type.application"},
			isExecutable: true,
			isNative:     true,
		},
		{
			name:         "app extension",
			target:       Target{isa: nativeTargetISA, productPath: "Today.appex", productType: "com.apple.product-type.app-extension"},
			isExecutable: true,
			isNative:     true,
		},
		{
			name:         "app clip",
			target:       Target{isa: nativeTargetISA, productPath: "Clip.app", productType: appClipProductType},
			isExecutable: true,
			isAppClip:    true,
			isNative:     true,
		},
		{
			name:     "unit test bundle",
			target:   Target{isa: nativeTargetISA, productPath: "Tests.xctest", productType: "com.apple.product-type.bundle.unit-test"},
			isTest:   true,
			isNative: true,
		},
		{
			name:     "ui test bundle",
			target:   Target{isa: nativeTargetISA, productPath: "UITests.xctest", productType: "com.apple.product-type.bundle.ui-testing"},
			isUITest: true,
			isTest:   true,
			isNative: true,
		},
		{
			name:     "OCTest bundle",
			target:   Target{isa: nativeTargetISA, productPath: "Legacy.octest", productType: "com.apple.product-type.bundle"},
			isTest:   true,
			isNative: true,
		},
		{
			name:   "aggregate target produces nothing",
			target: Target{isa: aggregateTargetISA},
		},
		{
			name:     "static library is not executable",
			target:   Target{isa: nativeTargetISA, productPath: "libThing.a", productType: "com.apple.product-type.library.static"},
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

func TestParseTarget_fromFixture(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	target, ok := project.TargetByName("XcodeProj")
	require.True(t, ok)

	assert.Equal(t, appTargetID, target.ID)
	assert.Equal(t, nativeTargetISA, target.isa)
	assert.Equal(t, "com.apple.product-type.application", target.productType)
	assert.Equal(t, "XcodeProj.app", target.productPath)
	assert.Equal(t, []string{todayExtensionTarget}, target.dependencyTargetIDs)
	assert.Equal(t, "Release", target.DefaultConfigurationName)
	require.Len(t, target.BuildConfigurations, 2)
	assert.NotEmpty(t, target.buildPhaseIDs)
}

func TestParseTarget_unknownTargetType(t *testing.T) {
	objects := map[string]any{
		"T1": map[string]any{"isa": "PBXSomethingElse", "name": "X"},
	}

	_, err := parseTarget("T1", objects)
	require.ErrorContains(t, err, "unknown target type")
}
