// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	xcodeversion "github.com/bitrise-io/go-xcode/v2/xcodeversion"
	mock "github.com/stretchr/testify/mock"
)

// Reader is an autogenerated mock type for the Reader type
type Reader struct {
	mock.Mock
}

// GetVersion provides a mock function with no fields
func (_m *Reader) GetVersion() (xcodeversion.Version, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetVersion")
	}

	var r0 xcodeversion.Version
	var r1 error
	if rf, ok := ret.Get(0).(func() (xcodeversion.Version, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() xcodeversion.Version); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(xcodeversion.Version)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewXcodeVersionReader creates a new instance of Reader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewXcodeVersionReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *Reader {
	mock := &Reader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
