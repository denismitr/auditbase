// Code generated by MockGen. DO NOT EDIT.
// Source: queue/message.go

// Package mock_queue is a generated GoMock package.
package mock_queue

import (
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

// Ack mocks base method
func (m *MockReceivedMessage) Ack() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ack")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ack indicates an expected call of Ack
func (mr *MockReceivedMessageMockRecorder) Ack() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ack", reflect.TypeOf((*MockReceivedMessage)(nil).Ack))
}

// Reject mocks base method
func (m *MockReceivedMessage) Reject(requeue bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reject", requeue)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reject indicates an expected call of Reject
func (mr *MockReceivedMessageMockRecorder) Reject(requeue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reject", reflect.TypeOf((*MockReceivedMessage)(nil).Reject), requeue)
}
