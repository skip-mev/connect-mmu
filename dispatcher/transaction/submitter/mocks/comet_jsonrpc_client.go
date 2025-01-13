// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	mock "github.com/stretchr/testify/mock"

	types "github.com/cometbft/cometbft/types"
)

// CometJSONRPCClient is an autogenerated mock type for the CometJSONRPCClient type
type CometJSONRPCClient struct {
	mock.Mock
}

type CometJSONRPCClient_Expecter struct {
	mock *mock.Mock
}

func (_m *CometJSONRPCClient) EXPECT() *CometJSONRPCClient_Expecter {
	return &CometJSONRPCClient_Expecter{mock: &_m.Mock}
}

// BroadcastTxSync provides a mock function with given fields: ctx, tx
func (_m *CometJSONRPCClient) BroadcastTxSync(ctx context.Context, tx types.Tx) (*coretypes.ResultBroadcastTx, error) {
	ret := _m.Called(ctx, tx)

	if len(ret) == 0 {
		panic("no return value specified for BroadcastTxSync")
	}

	var r0 *coretypes.ResultBroadcastTx
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Tx) (*coretypes.ResultBroadcastTx, error)); ok {
		return rf(ctx, tx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.Tx) *coretypes.ResultBroadcastTx); ok {
		r0 = rf(ctx, tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTx)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.Tx) error); ok {
		r1 = rf(ctx, tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CometJSONRPCClient_BroadcastTxSync_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BroadcastTxSync'
type CometJSONRPCClient_BroadcastTxSync_Call struct {
	*mock.Call
}

// BroadcastTxSync is a helper method to define mock.On call
//   - ctx context.Context
//   - tx types.Tx
func (_e *CometJSONRPCClient_Expecter) BroadcastTxSync(ctx interface{}, tx interface{}) *CometJSONRPCClient_BroadcastTxSync_Call {
	return &CometJSONRPCClient_BroadcastTxSync_Call{Call: _e.mock.On("BroadcastTxSync", ctx, tx)}
}

func (_c *CometJSONRPCClient_BroadcastTxSync_Call) Run(run func(ctx context.Context, tx types.Tx)) *CometJSONRPCClient_BroadcastTxSync_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.Tx))
	})
	return _c
}

func (_c *CometJSONRPCClient_BroadcastTxSync_Call) Return(_a0 *coretypes.ResultBroadcastTx, _a1 error) *CometJSONRPCClient_BroadcastTxSync_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CometJSONRPCClient_BroadcastTxSync_Call) RunAndReturn(run func(context.Context, types.Tx) (*coretypes.ResultBroadcastTx, error)) *CometJSONRPCClient_BroadcastTxSync_Call {
	_c.Call.Return(run)
	return _c
}

// Tx provides a mock function with given fields: ctx, hash, prove
func (_m *CometJSONRPCClient) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	ret := _m.Called(ctx, hash, prove)

	if len(ret) == 0 {
		panic("no return value specified for Tx")
	}

	var r0 *coretypes.ResultTx
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte, bool) (*coretypes.ResultTx, error)); ok {
		return rf(ctx, hash, prove)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []byte, bool) *coretypes.ResultTx); ok {
		r0 = rf(ctx, hash, prove)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultTx)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []byte, bool) error); ok {
		r1 = rf(ctx, hash, prove)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CometJSONRPCClient_Tx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Tx'
type CometJSONRPCClient_Tx_Call struct {
	*mock.Call
}

// Tx is a helper method to define mock.On call
//   - ctx context.Context
//   - hash []byte
//   - prove bool
func (_e *CometJSONRPCClient_Expecter) Tx(ctx interface{}, hash interface{}, prove interface{}) *CometJSONRPCClient_Tx_Call {
	return &CometJSONRPCClient_Tx_Call{Call: _e.mock.On("Tx", ctx, hash, prove)}
}

func (_c *CometJSONRPCClient_Tx_Call) Run(run func(ctx context.Context, hash []byte, prove bool)) *CometJSONRPCClient_Tx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte), args[2].(bool))
	})
	return _c
}

func (_c *CometJSONRPCClient_Tx_Call) Return(_a0 *coretypes.ResultTx, _a1 error) *CometJSONRPCClient_Tx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CometJSONRPCClient_Tx_Call) RunAndReturn(run func(context.Context, []byte, bool) (*coretypes.ResultTx, error)) *CometJSONRPCClient_Tx_Call {
	_c.Call.Return(run)
	return _c
}

// NewCometJSONRPCClient creates a new instance of CometJSONRPCClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCometJSONRPCClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *CometJSONRPCClient {
	mock := &CometJSONRPCClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
