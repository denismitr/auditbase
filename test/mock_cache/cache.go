// Code generated by MockGen. DO NOT EDIT.
// Source: cache/cache.go

// Package mock_cache is a generated GoMock package.
package mock_cache

import (
	cache "github.com/denismitr/auditbase/cache"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockCacher is a mock of Cacher interface
type MockCacher struct {
	ctrl     *gomock.Controller
	recorder *MockCacherMockRecorder
}

// MockCacherMockRecorder is the mock recorder for MockCacher
type MockCacherMockRecorder struct {
	mock *MockCacher
}

// NewMockCacher creates a new mock instance
func NewMockCacher(ctrl *gomock.Controller) *MockCacher {
	mock := &MockCacher{ctrl: ctrl}
	mock.recorder = &MockCacherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCacher) EXPECT() *MockCacherMockRecorder {
	return m.recorder
}

// Remember mocks base method
func (m *MockCacher) Remember(arg0 cache.TargetParser) cache.RememberFunc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remember", arg0)
	ret0, _ := ret[0].(cache.RememberFunc)
	return ret0
}

// Remember indicates an expected call of Remember
func (mr *MockCacherMockRecorder) Remember(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remember", reflect.TypeOf((*MockCacher)(nil).Remember), arg0)
}

// Has mocks base method
func (m *MockCacher) Has(key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Has", key)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Has indicates an expected call of Has
func (mr *MockCacherMockRecorder) Has(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockCacher)(nil).Has), key)
}

// CreateKey mocks base method
func (m *MockCacher) CreateKey(key string, ttl time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateKey", key, ttl)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateKey indicates an expected call of CreateKey
func (mr *MockCacherMockRecorder) CreateKey(key, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateKey", reflect.TypeOf((*MockCacher)(nil).CreateKey), key, ttl)
}