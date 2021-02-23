// Code generated by MockGen. DO NOT EDIT.
// Source: flow/flow.go

// Package mock_flow is a generated GoMock package.
package mock_flow

import (
	flow "github.com/denismitr/auditbase/flow"
	model "github.com/denismitr/auditbase/model"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockEventFlow is a mock of EventFlow interface
type MockEventFlow struct {
	ctrl     *gomock.Controller
	recorder *MockEventFlowMockRecorder
}

// MockEventFlowMockRecorder is the mock recorder for MockEventFlow
type MockEventFlowMockRecorder struct {
	mock *MockEventFlow
}

// NewMockEventFlow creates a new mock instance
func NewMockEventFlow(ctrl *gomock.Controller) *MockEventFlow {
	mock := &MockEventFlow{ctrl: ctrl}
	mock.recorder = &MockEventFlowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventFlow) EXPECT() *MockEventFlowMockRecorder {
	return m.recorder
}

// Send mocks base method
func (m *MockEventFlow) Send(e *model.Action) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", e)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockEventFlowMockRecorder) Send(e interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockEventFlow)(nil).Send), e)
}

// Receive mocks base method
func (m *MockEventFlow) Receive(queue, consumer string) <-chan flow.ReceivedAction {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Receive", queue, consumer)
	ret0, _ := ret[0].(<-chan flow.ReceivedAction)
	return ret0
}

// Receive indicates an expected call of Receive
func (mr *MockEventFlowMockRecorder) Receive(queue, consumer interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Receive", reflect.TypeOf((*MockEventFlow)(nil).Receive), queue, consumer)
}

// Requeue mocks base method
func (m *MockEventFlow) Requeue(arg0 flow.ReceivedAction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Requeue", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Requeue indicates an expected call of Requeue
func (mr *MockEventFlowMockRecorder) Requeue(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Requeue", reflect.TypeOf((*MockEventFlow)(nil).Requeue), arg0)
}

// Ack mocks base method
func (m *MockEventFlow) Ack(arg0 flow.ReceivedAction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ack", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ack indicates an expected call of Ack
func (mr *MockEventFlowMockRecorder) Ack(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ack", reflect.TypeOf((*MockEventFlow)(nil).Ack), arg0)
}

// Reject mocks base method
func (m *MockEventFlow) Reject(arg0 flow.ReceivedAction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reject", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reject indicates an expected call of Reject
func (mr *MockEventFlowMockRecorder) Reject(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reject", reflect.TypeOf((*MockEventFlow)(nil).Reject), arg0)
}

// Inspect mocks base method
func (m *MockEventFlow) Inspect() (flow.Status, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Inspect")
	ret0, _ := ret[0].(flow.Status)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Inspect indicates an expected call of Inspect
func (mr *MockEventFlowMockRecorder) Inspect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Inspect", reflect.TypeOf((*MockEventFlow)(nil).Inspect))
}

// Scaffold mocks base method
func (m *MockEventFlow) Scaffold() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Scaffold")
	ret0, _ := ret[0].(error)
	return ret0
}

// Scaffold indicates an expected call of Scaffold
func (mr *MockEventFlowMockRecorder) Scaffold() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scaffold", reflect.TypeOf((*MockEventFlow)(nil).Scaffold))
}

// NotifyOnStateChange mocks base method
func (m *MockEventFlow) NotifyOnStateChange(arg0 chan flow.State) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "NotifyOnStateChange", arg0)
}

// NotifyOnStateChange indicates an expected call of NotifyOnStateChange
func (mr *MockEventFlowMockRecorder) NotifyOnStateChange(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyOnStateChange", reflect.TypeOf((*MockEventFlow)(nil).NotifyOnStateChange), arg0)
}

// Start mocks base method
func (m *MockEventFlow) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockEventFlowMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockEventFlow)(nil).Start))
}

// Stop mocks base method
func (m *MockEventFlow) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop
func (mr *MockEventFlowMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockEventFlow)(nil).Stop))
}
