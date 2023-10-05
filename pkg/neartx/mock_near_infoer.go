// Code generated by mockery v2.14.0. DO NOT EDIT.

package neartx

import (
	"github.com/rarimo/near-go/common"
	client "github.com/rarimo/near-go/nearclient"

	context "context"


	mock "github.com/stretchr/testify/mock"
)

// MockNearInfoer is an autogenerated mock type for the NearInfoer type
type MockNearInfoer struct {
	mock.Mock
}

// AccessKeyView provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *MockNearInfoer) AccessKeyView(_a0 context.Context, _a1 string, _a2 common.Base58PublicKey, _a3 client.BlockCharacteristic) (common.AccessKeyView, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 common.AccessKeyView
	if rf, ok := ret.Get(0).(func(context.Context, string, common.Base58PublicKey, client.BlockCharacteristic) common.AccessKeyView); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(common.AccessKeyView)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, common.Base58PublicKey, client.BlockCharacteristic) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockDetails provides a mock function with given fields: _a0, _a1
func (_m *MockNearInfoer) BlockDetails(_a0 context.Context, _a1 client.BlockCharacteristic) (common.BlockView, error) {
	ret := _m.Called(_a0, _a1)

	var r0 common.BlockView
	if rf, ok := ret.Get(0).(func(context.Context, client.BlockCharacteristic) common.BlockView); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(common.BlockView)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, client.BlockCharacteristic) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockNearInfoer interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockNearInfoer creates a new instance of MockNearInfoer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockNearInfoer(t mockConstructorTestingTNewMockNearInfoer) *MockNearInfoer {
	mock := &MockNearInfoer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
