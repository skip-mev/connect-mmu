// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	cometbfttypes "github.com/cometbft/cometbft/types"
	client "github.com/cosmos/cosmos-sdk/client"

	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// SigningAgent is an autogenerated mock type for the SigningAgent type
type SigningAgent struct {
	mock.Mock
}

type SigningAgent_Expecter struct {
	mock *mock.Mock
}

func (_m *SigningAgent) EXPECT() *SigningAgent_Expecter {
	return &SigningAgent_Expecter{mock: &_m.Mock}
}

// GetSigningAccount provides a mock function with given fields: ctx
func (_m *SigningAgent) GetSigningAccount(ctx context.Context) (types.AccountI, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetSigningAccount")
	}

	var r0 types.AccountI
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (types.AccountI, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) types.AccountI); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.AccountI)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SigningAgent_GetSigningAccount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSigningAccount'
type SigningAgent_GetSigningAccount_Call struct {
	*mock.Call
}

// GetSigningAccount is a helper method to define mock.On call
//   - ctx context.Context
func (_e *SigningAgent_Expecter) GetSigningAccount(ctx interface{}) *SigningAgent_GetSigningAccount_Call {
	return &SigningAgent_GetSigningAccount_Call{Call: _e.mock.On("GetSigningAccount", ctx)}
}

func (_c *SigningAgent_GetSigningAccount_Call) Run(run func(ctx context.Context)) *SigningAgent_GetSigningAccount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *SigningAgent_GetSigningAccount_Call) Return(_a0 types.AccountI, _a1 error) *SigningAgent_GetSigningAccount_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SigningAgent_GetSigningAccount_Call) RunAndReturn(run func(context.Context) (types.AccountI, error)) *SigningAgent_GetSigningAccount_Call {
	_c.Call.Return(run)
	return _c
}

// Sign provides a mock function with given fields: ctx, txb
func (_m *SigningAgent) Sign(ctx context.Context, txb client.TxBuilder) (cometbfttypes.Tx, error) {
	ret := _m.Called(ctx, txb)

	if len(ret) == 0 {
		panic("no return value specified for Sign")
	}

	var r0 cometbfttypes.Tx
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, client.TxBuilder) (cometbfttypes.Tx, error)); ok {
		return rf(ctx, txb)
	}
	if rf, ok := ret.Get(0).(func(context.Context, client.TxBuilder) cometbfttypes.Tx); ok {
		r0 = rf(ctx, txb)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cometbfttypes.Tx)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, client.TxBuilder) error); ok {
		r1 = rf(ctx, txb)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SigningAgent_Sign_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sign'
type SigningAgent_Sign_Call struct {
	*mock.Call
}

// Sign is a helper method to define mock.On call
//   - ctx context.Context
//   - txb client.TxBuilder
func (_e *SigningAgent_Expecter) Sign(ctx interface{}, txb interface{}) *SigningAgent_Sign_Call {
	return &SigningAgent_Sign_Call{Call: _e.mock.On("Sign", ctx, txb)}
}

func (_c *SigningAgent_Sign_Call) Run(run func(ctx context.Context, txb client.TxBuilder)) *SigningAgent_Sign_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(client.TxBuilder))
	})
	return _c
}

func (_c *SigningAgent_Sign_Call) Return(_a0 cometbfttypes.Tx, _a1 error) *SigningAgent_Sign_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SigningAgent_Sign_Call) RunAndReturn(run func(context.Context, client.TxBuilder) (cometbfttypes.Tx, error)) *SigningAgent_Sign_Call {
	_c.Call.Return(run)
	return _c
}

// NewSigningAgent creates a new instance of SigningAgent. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSigningAgent(t interface {
	mock.TestingT
	Cleanup(func())
}) *SigningAgent {
	mock := &SigningAgent{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
