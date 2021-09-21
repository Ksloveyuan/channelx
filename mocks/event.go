// Code generated by MockGen. DO NOT EDIT.
// Source: event.go

// Package channelx_mock is a generated GoMock package.
package channelx_mock

import (
	context "context"
	channelx "github.com/Ksloveyuan/channelx"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockEvent is a mock of Event interface
type MockEvent struct {
	ctrl     *gomock.Controller
	recorder *MockEventMockRecorder
}

// MockEventMockRecorder is the mock recorder for MockEvent
type MockEventMockRecorder struct {
	mock *MockEvent
}

// NewMockEvent creates a new mock instance
func NewMockEvent(ctrl *gomock.Controller) *MockEvent {
	mock := &MockEvent{ctrl: ctrl}
	mock.recorder = &MockEventMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEvent) EXPECT() *MockEventMockRecorder {
	return m.recorder
}

// ID mocks base method
func (m *MockEvent) ID() channelx.EventID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(channelx.EventID)
	return ret0
}

// ID indicates an expected call of ID
func (mr *MockEventMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockEvent)(nil).ID))
}

// MockEventHandler is a mock of EventHandler interface
type MockEventHandler struct {
	ctrl     *gomock.Controller
	recorder *MockEventHandlerMockRecorder
}

// MockEventHandlerMockRecorder is the mock recorder for MockEventHandler
type MockEventHandlerMockRecorder struct {
	mock *MockEventHandler
}

// NewMockEventHandler creates a new mock instance
func NewMockEventHandler(ctrl *gomock.Controller) *MockEventHandler {
	mock := &MockEventHandler{ctrl: ctrl}
	mock.recorder = &MockEventHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventHandler) EXPECT() *MockEventHandlerMockRecorder {
	return m.recorder
}

// OnEvent mocks base method
func (m *MockEventHandler) OnEvent(ctx context.Context, event channelx.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnEvent", ctx, event)
	ret0, _ := ret[0].(error)
	return ret0
}

// OnEvent indicates an expected call of OnEvent
func (mr *MockEventHandlerMockRecorder) OnEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnEvent", reflect.TypeOf((*MockEventHandler)(nil).OnEvent), ctx, event)
}

// Logger mocks base method
func (m *MockEventHandler) Logger() channelx.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logger")
	ret0, _ := ret[0].(channelx.Logger)
	return ret0
}

// Logger indicates an expected call of Logger
func (mr *MockEventHandlerMockRecorder) Logger() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logger", reflect.TypeOf((*MockEventHandler)(nil).Logger))
}

// CanAutoRetry mocks base method
func (m *MockEventHandler) CanAutoRetry(err error) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanAutoRetry", err)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CanAutoRetry indicates an expected call of CanAutoRetry
func (mr *MockEventHandlerMockRecorder) CanAutoRetry(err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanAutoRetry", reflect.TypeOf((*MockEventHandler)(nil).CanAutoRetry), err)
}