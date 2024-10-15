package server

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// mockPoolManager is a mock of PoolManager interface.
type mockPoolManager struct {
	ctrl     *gomock.Controller
	recorder *mockPoolManagerMockRecorder
}

// mockPoolManagerMockRecorder is the mock recorder for mockPoolManager.
type mockPoolManagerMockRecorder struct {
	mock *mockPoolManager
}

// newMockPoolManager creates a new mock instance.
func newMockPoolManager(ctrl *gomock.Controller) *mockPoolManager {
	mock := &mockPoolManager{ctrl: ctrl}
	mock.recorder = &mockPoolManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *mockPoolManager) EXPECT() *mockPoolManagerMockRecorder {
	return m.recorder
}

// GetPool mocks base method.
func (m *mockPoolManager) GetPool(ctx context.Context, id string) (*Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPool", ctx, id)
	ret0, _ := ret[0].(*Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPool indicates an expected call of GetPool.
func (mr *mockPoolManagerMockRecorder) GetPool(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPool", reflect.TypeOf((*mockPoolManager)(nil).GetPool), ctx, id)
}

// ListPools mocks base method.
func (m *mockPoolManager) ListPools(ctx context.Context) ([]*Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPools", ctx)
	ret0, _ := ret[0].([]*Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPools indicates an expected call of ListPools.
func (mr *mockPoolManagerMockRecorder) ListPools(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPools", reflect.TypeOf((*mockPoolManager)(nil).ListPools), ctx)
}

// PausePool mocks base method.
func (m *mockPoolManager) PausePool(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PausePool", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// PausePool indicates an expected call of PausePool.
func (mr *mockPoolManagerMockRecorder) PausePool(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PausePool", reflect.TypeOf((*mockPoolManager)(nil).PausePool), ctx, id)
}

// Reload mocks base method.
func (m *mockPoolManager) Reload(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reload", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reload indicates an expected call of Reload.
func (mr *mockPoolManagerMockRecorder) Reload(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reload", reflect.TypeOf((*mockPoolManager)(nil).Reload), ctx)
}

// ResumePool mocks base method.
func (m *mockPoolManager) ResumePool(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResumePool", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResumePool indicates an expected call of ResumePool.
func (mr *mockPoolManagerMockRecorder) ResumePool(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResumePool", reflect.TypeOf((*mockPoolManager)(nil).ResumePool), ctx, id)
}

// ScalePool mocks base method.
func (m *mockPoolManager) ScalePool(ctx context.Context, id string, delta int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScalePool", ctx, id, delta)
	ret0, _ := ret[0].(error)
	return ret0
}

// ScalePool indicates an expected call of ScalePool.
func (mr *mockPoolManagerMockRecorder) ScalePool(ctx, id, delta any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScalePool", reflect.TypeOf((*mockPoolManager)(nil).ScalePool), ctx, id, delta)
}
