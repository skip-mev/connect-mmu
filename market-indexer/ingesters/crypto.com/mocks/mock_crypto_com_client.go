// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	crypto_com "github.com/skip-mev/connect-mmu/market-indexer/ingesters/crypto.com"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

type Client_Expecter struct {
	mock *mock.Mock
}

func (_m *Client) EXPECT() *Client_Expecter {
	return &Client_Expecter{mock: &_m.Mock}
}

// Instruments provides a mock function with given fields: _a0
func (_m *Client) Instruments(_a0 context.Context) (crypto_com.InstrumentsResponse, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Instruments")
	}

	var r0 crypto_com.InstrumentsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (crypto_com.InstrumentsResponse, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) crypto_com.InstrumentsResponse); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(crypto_com.InstrumentsResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Instruments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Instruments'
type Client_Instruments_Call struct {
	*mock.Call
}

// Instruments is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *Client_Expecter) Instruments(_a0 interface{}) *Client_Instruments_Call {
	return &Client_Instruments_Call{Call: _e.mock.On("Instruments", _a0)}
}

func (_c *Client_Instruments_Call) Run(run func(_a0 context.Context)) *Client_Instruments_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Client_Instruments_Call) Return(_a0 crypto_com.InstrumentsResponse, _a1 error) *Client_Instruments_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Instruments_Call) RunAndReturn(run func(context.Context) (crypto_com.InstrumentsResponse, error)) *Client_Instruments_Call {
	_c.Call.Return(run)
	return _c
}

// Tickers provides a mock function with given fields: _a0
func (_m *Client) Tickers(_a0 context.Context) (crypto_com.TickersResponse, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Tickers")
	}

	var r0 crypto_com.TickersResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (crypto_com.TickersResponse, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) crypto_com.TickersResponse); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(crypto_com.TickersResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Client_Tickers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Tickers'
type Client_Tickers_Call struct {
	*mock.Call
}

// Tickers is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *Client_Expecter) Tickers(_a0 interface{}) *Client_Tickers_Call {
	return &Client_Tickers_Call{Call: _e.mock.On("Tickers", _a0)}
}

func (_c *Client_Tickers_Call) Run(run func(_a0 context.Context)) *Client_Tickers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Client_Tickers_Call) Return(_a0 crypto_com.TickersResponse, _a1 error) *Client_Tickers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Client_Tickers_Call) RunAndReturn(run func(context.Context) (crypto_com.TickersResponse, error)) *Client_Tickers_Call {
	_c.Call.Return(run)
	return _c
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
