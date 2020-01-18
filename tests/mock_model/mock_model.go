package mock_model

import (
        model "github.com/denismitr/auditbase/model"
        gomock "github.com/golang/mock/gomock"
        reflect "reflect"
)

// MockEventRepository is a mock of EventRepository interface
type MockEventRepository struct {
        ctrl     *gomock.Controller
        recorder *MockEventRepositoryMockRecorder
}

// MockEventRepositoryMockRecorder is the mock recorder for MockEventRepository
type MockEventRepositoryMockRecorder struct {
        mock *MockEventRepository
}

// NewMockEventRepository creates a new mock instance
func NewMockEventRepository(ctrl *gomock.Controller) *MockEventRepository {
        mock := &MockEventRepository{ctrl: ctrl}
        mock.recorder = &MockEventRepositoryMockRecorder{mock}
        return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventRepository) EXPECT() *MockEventRepositoryMockRecorder {
        return m.recorder
}

// Create mocks base method
func (m *MockEventRepository) Create(arg0 model.Event) error {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "Create", arg0)
        ret0, _ := ret[0].(error)
        return ret0
}

// Create indicates an expected call of Create
func (mr *MockEventRepositoryMockRecorder) Create(arg0 interface{}) *gomock.Call {
        mr.mock.ctrl.T.Helper()
        return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockEventRepository)(nil).Create), arg0)
}

// Delete mocks base method
func (m *MockEventRepository) Delete(arg0 string) error {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "Delete", arg0)
        ret0, _ := ret[0].(error)
        return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockEventRepositoryMockRecorder) Delete(arg0 interface{}) *gomock.Call {
        mr.mock.ctrl.T.Helper()
        return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockEventRepository)(nil).Delete), arg0)
}

// Count mocks base method
func (m *MockEventRepository) Count() (int, error) {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "Count")
        ret0, _ := ret[0].(int)
        ret1, _ := ret[1].(error)
        return ret0, ret1
}

// Count indicates an expected call of Count
func (mr *MockEventRepositoryMockRecorder) Count() *gomock.Call {
        mr.mock.ctrl.T.Helper()
        return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*MockEventRepository)(nil).Count))
}

// FindOneByID mocks base method
func (m *MockEventRepository) FindOneByID(arg0 string) (model.Event, error) {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "FindOneByID", arg0)
        ret0, _ := ret[0].(model.Event)
        ret1, _ := ret[1].(error)
        return ret0, ret1
}

// FindOneByID indicates an expected call of FindOneByID
func (mr *MockEventRepositoryMockRecorder) FindOneByID(arg0 interface{}) *gomock.Call {
        mr.mock.ctrl.T.Helper()
        return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneByID", reflect.TypeOf((*MockEventRepository)(nil).FindOneByID), arg0)
}

// SelectAll mocks base method
func (m *MockEventRepository) SelectAll() ([]model.Event, error) {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "SelectAll")
        ret0, _ := ret[0].([]model.Event)
        ret1, _ := ret[1].(error)
        return ret0, ret1
}

// SelectAll indicates an expected call of SelectAll
func (mr *MockEventRepositoryMockRecorder) SelectAll() *gomock.Call {
        mr.mock.ctrl.T.Helper()
        return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAll", reflect.TypeOf((*MockEventRepository)(nil).SelectAll))
}