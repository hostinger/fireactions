// Code generated by MockGen. DO NOT EDIT.
// Source: client/rootfs/os.go
//
// Generated by this command:
//
//	mockgen -source client/rootfs/os.go -destination client/rootfs/mocks/os.go -package mocks OSInterface
//
// Package mocks is a generated GoMock package.
package mocks

import (
	fs "io/fs"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockOSInterface is a mock of OSInterface interface.
type MockOSInterface struct {
	ctrl     *gomock.Controller
	recorder *MockOSInterfaceMockRecorder
}

// MockOSInterfaceMockRecorder is the mock recorder for MockOSInterface.
type MockOSInterfaceMockRecorder struct {
	mock *MockOSInterface
}

// NewMockOSInterface creates a new mock instance.
func NewMockOSInterface(ctrl *gomock.Controller) *MockOSInterface {
	mock := &MockOSInterface{ctrl: ctrl}
	mock.recorder = &MockOSInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOSInterface) EXPECT() *MockOSInterfaceMockRecorder {
	return m.recorder
}

// IsNotExist mocks base method.
func (m *MockOSInterface) IsNotExist(err error) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsNotExist", err)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsNotExist indicates an expected call of IsNotExist.
func (mr *MockOSInterfaceMockRecorder) IsNotExist(err any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsNotExist", reflect.TypeOf((*MockOSInterface)(nil).IsNotExist), err)
}

// MkdirTemp mocks base method.
func (m *MockOSInterface) MkdirTemp(dir, pattern string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MkdirTemp", dir, pattern)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MkdirTemp indicates an expected call of MkdirTemp.
func (mr *MockOSInterfaceMockRecorder) MkdirTemp(dir, pattern any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MkdirTemp", reflect.TypeOf((*MockOSInterface)(nil).MkdirTemp), dir, pattern)
}

// Remove mocks base method.
func (m *MockOSInterface) Remove(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockOSInterfaceMockRecorder) Remove(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockOSInterface)(nil).Remove), name)
}

// RemoveAll mocks base method.
func (m *MockOSInterface) RemoveAll(path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveAll", path)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveAll indicates an expected call of RemoveAll.
func (mr *MockOSInterfaceMockRecorder) RemoveAll(path any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAll", reflect.TypeOf((*MockOSInterface)(nil).RemoveAll), path)
}

// Symlink mocks base method.
func (m *MockOSInterface) Symlink(oldname, newname string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Symlink", oldname, newname)
	ret0, _ := ret[0].(error)
	return ret0
}

// Symlink indicates an expected call of Symlink.
func (mr *MockOSInterfaceMockRecorder) Symlink(oldname, newname any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Symlink", reflect.TypeOf((*MockOSInterface)(nil).Symlink), oldname, newname)
}

// WriteFile mocks base method.
func (m *MockOSInterface) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteFile", name, data, perm)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteFile indicates an expected call of WriteFile.
func (mr *MockOSInterfaceMockRecorder) WriteFile(name, data, perm any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteFile", reflect.TypeOf((*MockOSInterface)(nil).WriteFile), name, data, perm)
}
