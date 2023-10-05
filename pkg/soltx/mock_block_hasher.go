// Code generated by mockery v2.14.0. DO NOT EDIT.

package soltx

import (
	context "context"

	rpc "github.com/olegfomenko/solana-go/rpc"
	mock "github.com/stretchr/testify/mock"
)

// mockBlockHasher is an autogenerated mock type for the blockHasher type
type mockBlockHasher struct {
	mock.Mock
}

// GetRecentBlockhash provides a mock function with given fields: _a0, _a1
func (_m *mockBlockHasher) GetRecentBlockhash(_a0 context.Context, _a1 rpc.CommitmentType) (*rpc.GetRecentBlockhashResult, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *rpc.GetRecentBlockhashResult
	if rf, ok := ret.Get(0).(func(context.Context, rpc.CommitmentType) *rpc.GetRecentBlockhashResult); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rpc.GetRecentBlockhashResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, rpc.CommitmentType) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTnewMockBlockHasher interface {
	mock.TestingT
	Cleanup(func())
}

// newMockBlockHasher creates a new instance of mockBlockHasher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockBlockHasher(t mockConstructorTestingTnewMockBlockHasher) *mockBlockHasher {
	mock := &mockBlockHasher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
