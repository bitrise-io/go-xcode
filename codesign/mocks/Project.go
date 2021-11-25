// Code generated by mockery 2.9.4. DO NOT EDIT.

package mocks

import (
	autocodesign "github.com/bitrise-io/go-xcode/autocodesign"

	mock "github.com/stretchr/testify/mock"
)

// Project is an autogenerated mock type for the Project type
type Project struct {
	mock.Mock
}

// ForceCodesignAssets provides a mock function with given fields: distribution, codesignAssetsByDistributionType
func (_m *Project) ForceCodesignAssets(distribution autocodesign.DistributionType, codesignAssetsByDistributionType map[autocodesign.DistributionType]autocodesign.AppCodesignAssets) error {
	ret := _m.Called(distribution, codesignAssetsByDistributionType)

	var r0 error
	if rf, ok := ret.Get(0).(func(autocodesign.DistributionType, map[autocodesign.DistributionType]autocodesign.AppCodesignAssets) error); ok {
		r0 = rf(distribution, codesignAssetsByDistributionType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAppLayout provides a mock function with given fields: uiTestTargets
func (_m *Project) GetAppLayout(uiTestTargets bool) (autocodesign.AppLayout, error) {
	ret := _m.Called(uiTestTargets)

	var r0 autocodesign.AppLayout
	if rf, ok := ret.Get(0).(func(bool) autocodesign.AppLayout); ok {
		r0 = rf(uiTestTargets)
	} else {
		r0 = ret.Get(0).(autocodesign.AppLayout)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool) error); ok {
		r1 = rf(uiTestTargets)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsSigningManagedAutomatically provides a mock function with given fields:
func (_m *Project) IsSigningManagedAutomatically() (bool, error) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Platform provides a mock function with given fields:
func (_m *Project) Platform() (autocodesign.Platform, error) {
	ret := _m.Called()

	var r0 autocodesign.Platform
	if rf, ok := ret.Get(0).(func() autocodesign.Platform); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(autocodesign.Platform)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
