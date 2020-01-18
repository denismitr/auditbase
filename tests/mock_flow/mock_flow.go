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
func (m *MockEventFlow) Send(e model.Event) error {
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
func (m *MockEventFlow) Receive(consumer string) <-chan flow.ReceivedEvent {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "Receive", consumer)
        ret0, _ := ret[0].(<-chan flow.ReceivedEvent)
        return ret0
}

// Receive indicates an expected call of Receive
func (mr *MockEventFlowMockRecorder) Receive(consumer interface{}) *gomock.Call {
        mr.mock.ctrl.T.Helper()
        return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Receive", reflect.TypeOf((*MockEventFlow)(nil).Receive), consumer)
}

// Inspect mocks base method
func (m *MockEventFlow) Inspect() (int, int, error) {
        m.ctrl.T.Helper()
        ret := m.ctrl.Call(m, "Inspect")
        ret0, _ := ret[0].(int)
        ret1, _ := ret[1].(int)
        ret2, _ := ret[2].(error)
        return ret0, ret1, ret2
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