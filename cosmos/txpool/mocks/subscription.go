// Code generated by mockery v2.35.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Subscription is an autogenerated mock type for the Subscription type
type Subscription struct {
	mock.Mock
}

type Subscription_Expecter struct {
	mock *mock.Mock
}

func (_m *Subscription) EXPECT() *Subscription_Expecter {
	return &Subscription_Expecter{mock: &_m.Mock}
}

// Err provides a mock function with given fields:
func (_m *Subscription) Err() <-chan error {
	ret := _m.Called()

	var r0 <-chan error
	if rf, ok := ret.Get(0).(func() <-chan error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan error)
		}
	}

	return r0
}

// Subscription_Err_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Err'
type Subscription_Err_Call struct {
	*mock.Call
}

// Err is a helper method to define mock.On call
func (_e *Subscription_Expecter) Err() *Subscription_Err_Call {
	return &Subscription_Err_Call{Call: _e.mock.On("Err")}
}

func (_c *Subscription_Err_Call) Run(run func()) *Subscription_Err_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Subscription_Err_Call) Return(_a0 <-chan error) *Subscription_Err_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Subscription_Err_Call) RunAndReturn(run func() <-chan error) *Subscription_Err_Call {
	_c.Call.Return(run)
	return _c
}

// Unsubscribe provides a mock function with given fields:
func (_m *Subscription) Unsubscribe() {
	_m.Called()
}

// Subscription_Unsubscribe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Unsubscribe'
type Subscription_Unsubscribe_Call struct {
	*mock.Call
}

// Unsubscribe is a helper method to define mock.On call
func (_e *Subscription_Expecter) Unsubscribe() *Subscription_Unsubscribe_Call {
	return &Subscription_Unsubscribe_Call{Call: _e.mock.On("Unsubscribe")}
}

func (_c *Subscription_Unsubscribe_Call) Run(run func()) *Subscription_Unsubscribe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Subscription_Unsubscribe_Call) Return() *Subscription_Unsubscribe_Call {
	_c.Call.Return()
	return _c
}

func (_c *Subscription_Unsubscribe_Call) RunAndReturn(run func()) *Subscription_Unsubscribe_Call {
	_c.Call.Return(run)
	return _c
}

// NewSubscription creates a new instance of Subscription. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSubscription(t interface {
	mock.TestingT
	Cleanup(func())
}) *Subscription {
	mock := &Subscription{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
