// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	provider "github.com/skip-mev/connect-mmu/store/provider"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

type Store_Expecter struct {
	mock *mock.Mock
}

func (_m *Store) EXPECT() *Store_Expecter {
	return &Store_Expecter{mock: &_m.Mock}
}

// AddAssetInfo provides a mock function with given fields: ctx, params
func (_m *Store) AddAssetInfo(ctx context.Context, params provider.CreateAssetInfoParams) (provider.AssetInfo, error) {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for AddAssetInfo")
	}

	var r0 provider.AssetInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, provider.CreateAssetInfoParams) (provider.AssetInfo, error)); ok {
		return rf(ctx, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, provider.CreateAssetInfoParams) provider.AssetInfo); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Get(0).(provider.AssetInfo)
	}

	if rf, ok := ret.Get(1).(func(context.Context, provider.CreateAssetInfoParams) error); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store_AddAssetInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddAssetInfo'
type Store_AddAssetInfo_Call struct {
	*mock.Call
}

// AddAssetInfo is a helper method to define mock.On call
//   - ctx context.Context
//   - params provider.CreateAssetInfoParams
func (_e *Store_Expecter) AddAssetInfo(ctx interface{}, params interface{}) *Store_AddAssetInfo_Call {
	return &Store_AddAssetInfo_Call{Call: _e.mock.On("AddAssetInfo", ctx, params)}
}

func (_c *Store_AddAssetInfo_Call) Run(run func(ctx context.Context, params provider.CreateAssetInfoParams)) *Store_AddAssetInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(provider.CreateAssetInfoParams))
	})
	return _c
}

func (_c *Store_AddAssetInfo_Call) Return(_a0 provider.AssetInfo, _a1 error) *Store_AddAssetInfo_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Store_AddAssetInfo_Call) RunAndReturn(run func(context.Context, provider.CreateAssetInfoParams) (provider.AssetInfo, error)) *Store_AddAssetInfo_Call {
	_c.Call.Return(run)
	return _c
}

// AddProviderMarket provides a mock function with given fields: ctx, params
func (_m *Store) AddProviderMarket(ctx context.Context, params provider.CreateProviderMarketParams) (provider.ProviderMarket, error) {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for AddProviderMarket")
	}

	var r0 provider.ProviderMarket
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, provider.CreateProviderMarketParams) (provider.ProviderMarket, error)); ok {
		return rf(ctx, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, provider.CreateProviderMarketParams) provider.ProviderMarket); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Get(0).(provider.ProviderMarket)
	}

	if rf, ok := ret.Get(1).(func(context.Context, provider.CreateProviderMarketParams) error); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store_AddProviderMarket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddProviderMarket'
type Store_AddProviderMarket_Call struct {
	*mock.Call
}

// AddProviderMarket is a helper method to define mock.On call
//   - ctx context.Context
//   - params provider.CreateProviderMarketParams
func (_e *Store_Expecter) AddProviderMarket(ctx interface{}, params interface{}) *Store_AddProviderMarket_Call {
	return &Store_AddProviderMarket_Call{Call: _e.mock.On("AddProviderMarket", ctx, params)}
}

func (_c *Store_AddProviderMarket_Call) Run(run func(ctx context.Context, params provider.CreateProviderMarketParams)) *Store_AddProviderMarket_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(provider.CreateProviderMarketParams))
	})
	return _c
}

func (_c *Store_AddProviderMarket_Call) Return(_a0 provider.ProviderMarket, _a1 error) *Store_AddProviderMarket_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Store_AddProviderMarket_Call) RunAndReturn(run func(context.Context, provider.CreateProviderMarketParams) (provider.ProviderMarket, error)) *Store_AddProviderMarket_Call {
	_c.Call.Return(run)
	return _c
}

// GetProviderMarkets provides a mock function with given fields: ctx, params
func (_m *Store) GetProviderMarkets(ctx context.Context, params provider.GetFilteredProviderMarketsParams) ([]provider.GetFilteredProviderMarketsRow, error) {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for GetProviderMarkets")
	}

	var r0 []provider.GetFilteredProviderMarketsRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, provider.GetFilteredProviderMarketsParams) ([]provider.GetFilteredProviderMarketsRow, error)); ok {
		return rf(ctx, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, provider.GetFilteredProviderMarketsParams) []provider.GetFilteredProviderMarketsRow); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]provider.GetFilteredProviderMarketsRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, provider.GetFilteredProviderMarketsParams) error); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store_GetProviderMarkets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProviderMarkets'
type Store_GetProviderMarkets_Call struct {
	*mock.Call
}

// GetProviderMarkets is a helper method to define mock.On call
//   - ctx context.Context
//   - params provider.GetFilteredProviderMarketsParams
func (_e *Store_Expecter) GetProviderMarkets(ctx interface{}, params interface{}) *Store_GetProviderMarkets_Call {
	return &Store_GetProviderMarkets_Call{Call: _e.mock.On("GetProviderMarkets", ctx, params)}
}

func (_c *Store_GetProviderMarkets_Call) Run(run func(ctx context.Context, params provider.GetFilteredProviderMarketsParams)) *Store_GetProviderMarkets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(provider.GetFilteredProviderMarketsParams))
	})
	return _c
}

func (_c *Store_GetProviderMarkets_Call) Return(_a0 []provider.GetFilteredProviderMarketsRow, _a1 error) *Store_GetProviderMarkets_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Store_GetProviderMarkets_Call) RunAndReturn(run func(context.Context, provider.GetFilteredProviderMarketsParams) ([]provider.GetFilteredProviderMarketsRow, error)) *Store_GetProviderMarkets_Call {
	_c.Call.Return(run)
	return _c
}

// WriteToPath provides a mock function with given fields: ctx, path
func (_m *Store) WriteToPath(ctx context.Context, path string) error {
	ret := _m.Called(ctx, path)

	if len(ret) == 0 {
		panic("no return value specified for WriteToPath")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store_WriteToPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteToPath'
type Store_WriteToPath_Call struct {
	*mock.Call
}

// WriteToPath is a helper method to define mock.On call
//   - ctx context.Context
//   - path string
func (_e *Store_Expecter) WriteToPath(ctx interface{}, path interface{}) *Store_WriteToPath_Call {
	return &Store_WriteToPath_Call{Call: _e.mock.On("WriteToPath", ctx, path)}
}

func (_c *Store_WriteToPath_Call) Run(run func(ctx context.Context, path string)) *Store_WriteToPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Store_WriteToPath_Call) Return(_a0 error) *Store_WriteToPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Store_WriteToPath_Call) RunAndReturn(run func(context.Context, string) error) *Store_WriteToPath_Call {
	_c.Call.Return(run)
	return _c
}

// NewStore creates a new instance of Store. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *Store {
	mock := &Store{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
