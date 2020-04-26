// Code generated by MockGen. DO NOT EDIT.
// Source: queue/message.go

// Package mock_queue is a generated GoMock package.
package mock_queue

import (
	queue "github.com/denismitr/auditbase/queue"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockMessage is a mock of Message interface
type MockMessage struct {
	ctrl     *gomock.Controller
	recorder *MockMessageMockRecorder
}

// MockMessageMockRecorder is the mock recorder for MockMessage
type MockMessageMockRecorder struct {
	mock *MockMessage
}

// NewMockMessage creates a new mock instance
func NewMockMessage(ctrl *gomock.Controller) *MockMessage {
	mock := &MockMessage{ctrl: ctrl}
	mock.recorder = &MockMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMessage) EXPECT() *MockMessageMockRecorder {
	return m.recorder
}

// Body mocks base method
func (m *MockMessage) Body() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Body")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Body indicates an expected call of Body
func (mr *MockMessageMockRecorder) Body() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Body", reflect.TypeOf((*MockMessage)(nil).Body))
}

// ContentType mocks base method
func (m *MockMessage) ContentType() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContentType")
	ret0, _ := ret[0].(string)
	return ret0
}

// ContentType indicates an expected call of ContentType
func (mr *MockMessageMockRecorder) ContentType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContentType", reflect.TypeOf((*MockMessage)(nil).ContentType))
}

// Attempt mocks base method
func (m *MockMessage) Attempt() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Attempt")
	ret0, _ := ret[0].(int)
	return ret0
}

// Attempt indicates an expected call of Attempt
func (mr *MockMessageMockRecorder) Attempt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Attempt", reflect.TypeOf((*MockMessage)(nil).Attempt))
}

// MockReceivedMessage is a mock of ReceivedMessage interface
type MockReceivedMessage struct {
	ctrl     *gomock.Controller
	recorder *MockReceivedMessageMockRecorder
}

// MockReceivedMessageMockRecorder is the mock recorder for MockReceivedMessage
type MockReceivedMessageMockRecorder struct {
	mock *MockReceivedMessage
}

// NewMockReceivedMessage creates a new mock instance
func NewMockReceivedMessage(ctrl *gomock.Controller) *MockReceivedMessage {
	mock := &MockReceivedMessage{ctrl: ctrl}
	mock.recorder = &MockReceivedMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReceivedMessage) EXPECT() *MockReceivedMessageMockRecorder {
	return m.recorder
}

// Body mocks base method
func (m *MockReceivedMessage) Body() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Body")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Body indicates an expected call of Body
func (mr *MockReceivedMessageMockRecorder) Body() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Body", reflect.TypeOf((*MockReceivedMessage)(nil).Body))
}

// Queue mocks base method
func (m *MockReceivedMessage) Queue() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Queue")
	ret0, _ := ret[0].(string)
	return ret0
}

// Queue indicates an expected call of Queue
func (mr *MockReceivedMessageMockRecorder) Queue() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Queue", reflect.TypeOf((*MockReceivedMessage)(nil).Queue))
}

// Attempt mocks base method
func (m *MockReceivedMessage) Attempt() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Attempt")
	ret0, _ := ret[0].(int)
	return ret0
}

// Attempt indicates an expected call of Attempt
func (mr *MockReceivedMessageMockRecorder) Attempt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Attempt", reflect.TypeOf((*MockReceivedMessage)(nil).Attempt))
}

// CloneToReque mocks base method
func (m *MockReceivedMessage) CloneToReque() queue.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloneToReque")
	ret0, _ := ret[0].(queue.Message)
	return ret0
}

// CloneToReque indicates an expected call of CloneToReque
func (mr *MockReceivedMessageMockRecorder) CloneToReque() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloneToReque", reflect.TypeOf((*MockReceivedMessage)(nil).CloneToReque))
}

// Tag mocks base method
func (m *MockReceivedMessage) Tag() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tag")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// Tag indicates an expected call of Tag
func (mr *MockReceivedMessageMockRecorder) Tag() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tag", reflect.TypeOf((*MockReceivedMessage)(nil).Tag))
}
