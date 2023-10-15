// Code generated by MockGen. DO NOT EDIT.
// Source: server/store/store.go
//
// Generated by this command:
//
//	mockgen -source server/store/store.go -destination server/store/mock/store.go -package mock Store
//
// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	"github.com/hostinger/fireactions/server/models"
	prometheus "github.com/prometheus/client_golang/prometheus"
	gomock "go.uber.org/mock/gomock"
)

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStore) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStoreMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStore)(nil).Close))
}

// Collect mocks base method.
func (m *MockStore) Collect(arg0 chan<- prometheus.Metric) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Collect", arg0)
}

// Collect indicates an expected call of Collect.
func (mr *MockStoreMockRecorder) Collect(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Collect", reflect.TypeOf((*MockStore)(nil).Collect), arg0)
}

// DeleteFlavor mocks base method.
func (m *MockStore) DeleteFlavor(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFlavor", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFlavor indicates an expected call of DeleteFlavor.
func (mr *MockStoreMockRecorder) DeleteFlavor(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFlavor", reflect.TypeOf((*MockStore)(nil).DeleteFlavor), ctx, name)
}

// DeleteGroup mocks base method.
func (m *MockStore) DeleteGroup(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGroup", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGroup indicates an expected call of DeleteGroup.
func (mr *MockStoreMockRecorder) DeleteGroup(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGroup", reflect.TypeOf((*MockStore)(nil).DeleteGroup), ctx, name)
}

// DeleteImage mocks base method.
func (m *MockStore) DeleteImage(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteImage", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteImage indicates an expected call of DeleteImage.
func (mr *MockStoreMockRecorder) DeleteImage(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteImage", reflect.TypeOf((*MockStore)(nil).DeleteImage), ctx, id)
}

// DeleteJob mocks base method.
func (m *MockStore) DeleteJob(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteJob", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteJob indicates an expected call of DeleteJob.
func (mr *MockStoreMockRecorder) DeleteJob(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteJob", reflect.TypeOf((*MockStore)(nil).DeleteJob), ctx, id)
}

// DeleteNode mocks base method.
func (m *MockStore) DeleteNode(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNode", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteNode indicates an expected call of DeleteNode.
func (mr *MockStoreMockRecorder) DeleteNode(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNode", reflect.TypeOf((*MockStore)(nil).DeleteNode), ctx, id)
}

// DeleteRunner mocks base method.
func (m *MockStore) DeleteRunner(ctx context.Context, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRunner", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRunner indicates an expected call of DeleteRunner.
func (mr *MockStoreMockRecorder) DeleteRunner(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRunner", reflect.TypeOf((*MockStore)(nil).DeleteRunner), ctx, id)
}

// Describe mocks base method.
func (m *MockStore) Describe(arg0 chan<- *prometheus.Desc) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Describe", arg0)
}

// Describe indicates an expected call of Describe.
func (mr *MockStoreMockRecorder) Describe(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Describe", reflect.TypeOf((*MockStore)(nil).Describe), arg0)
}

// GetFlavor mocks base method.
func (m *MockStore) GetFlavor(ctx context.Context, name string) (*models.Flavor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFlavor", ctx, name)
	ret0, _ := ret[0].(*models.Flavor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFlavor indicates an expected call of GetFlavor.
func (mr *MockStoreMockRecorder) GetFlavor(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFlavor", reflect.TypeOf((*MockStore)(nil).GetFlavor), ctx, name)
}

// GetGroup mocks base method.
func (m *MockStore) GetGroup(ctx context.Context, name string) (*models.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroup", ctx, name)
	ret0, _ := ret[0].(*models.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroup indicates an expected call of GetGroup.
func (mr *MockStoreMockRecorder) GetGroup(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroup", reflect.TypeOf((*MockStore)(nil).GetGroup), ctx, name)
}

// GetImage mocks base method.
func (m *MockStore) GetImage(ctx context.Context, id string) (*models.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetImage", ctx, id)
	ret0, _ := ret[0].(*models.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetImage indicates an expected call of GetImage.
func (mr *MockStoreMockRecorder) GetImage(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetImage", reflect.TypeOf((*MockStore)(nil).GetImage), ctx, id)
}

// GetImageByID mocks base method.
func (m *MockStore) GetImageByID(ctx context.Context, id string) (*models.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetImageByID", ctx, id)
	ret0, _ := ret[0].(*models.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetImageByID indicates an expected call of GetImageByID.
func (mr *MockStoreMockRecorder) GetImageByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetImageByID", reflect.TypeOf((*MockStore)(nil).GetImageByID), ctx, id)
}

// GetImageByName mocks base method.
func (m *MockStore) GetImageByName(ctx context.Context, name string) (*models.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetImageByName", ctx, name)
	ret0, _ := ret[0].(*models.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetImageByName indicates an expected call of GetImageByName.
func (mr *MockStoreMockRecorder) GetImageByName(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetImageByName", reflect.TypeOf((*MockStore)(nil).GetImageByName), ctx, name)
}

// GetJob mocks base method.
func (m *MockStore) GetJob(ctx context.Context, id string) (*models.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJob", ctx, id)
	ret0, _ := ret[0].(*models.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJob indicates an expected call of GetJob.
func (mr *MockStoreMockRecorder) GetJob(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJob", reflect.TypeOf((*MockStore)(nil).GetJob), ctx, id)
}

// GetNode mocks base method.
func (m *MockStore) GetNode(ctx context.Context, id string) (*models.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNode", ctx, id)
	ret0, _ := ret[0].(*models.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNode indicates an expected call of GetNode.
func (mr *MockStoreMockRecorder) GetNode(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNode", reflect.TypeOf((*MockStore)(nil).GetNode), ctx, id)
}

// GetNodeByName mocks base method.
func (m *MockStore) GetNodeByName(ctx context.Context, name string) (*models.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeByName", ctx, name)
	ret0, _ := ret[0].(*models.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodeByName indicates an expected call of GetNodeByName.
func (mr *MockStoreMockRecorder) GetNodeByName(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeByName", reflect.TypeOf((*MockStore)(nil).GetNodeByName), ctx, name)
}

// GetRunner mocks base method.
func (m *MockStore) GetRunner(ctx context.Context, id string) (*models.Runner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRunner", ctx, id)
	ret0, _ := ret[0].(*models.Runner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRunner indicates an expected call of GetRunner.
func (mr *MockStoreMockRecorder) GetRunner(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRunner", reflect.TypeOf((*MockStore)(nil).GetRunner), ctx, id)
}

// ListFlavors mocks base method.
func (m *MockStore) ListFlavors(ctx context.Context) ([]*models.Flavor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFlavors", ctx)
	ret0, _ := ret[0].([]*models.Flavor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFlavors indicates an expected call of ListFlavors.
func (mr *MockStoreMockRecorder) ListFlavors(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFlavors", reflect.TypeOf((*MockStore)(nil).ListFlavors), ctx)
}

// ListGroups mocks base method.
func (m *MockStore) ListGroups(ctx context.Context) ([]*models.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListGroups", ctx)
	ret0, _ := ret[0].([]*models.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListGroups indicates an expected call of ListGroups.
func (mr *MockStoreMockRecorder) ListGroups(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListGroups", reflect.TypeOf((*MockStore)(nil).ListGroups), ctx)
}

// ListImages mocks base method.
func (m *MockStore) ListImages(ctx context.Context) ([]*models.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListImages", ctx)
	ret0, _ := ret[0].([]*models.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListImages indicates an expected call of ListImages.
func (mr *MockStoreMockRecorder) ListImages(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListImages", reflect.TypeOf((*MockStore)(nil).ListImages), ctx)
}

// ListJobs mocks base method.
func (m *MockStore) ListJobs(ctx context.Context) ([]*models.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListJobs", ctx)
	ret0, _ := ret[0].([]*models.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListJobs indicates an expected call of ListJobs.
func (mr *MockStoreMockRecorder) ListJobs(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListJobs", reflect.TypeOf((*MockStore)(nil).ListJobs), ctx)
}

// ListNodes mocks base method.
func (m *MockStore) ListNodes(ctx context.Context) ([]*models.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListNodes", ctx)
	ret0, _ := ret[0].([]*models.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListNodes indicates an expected call of ListNodes.
func (mr *MockStoreMockRecorder) ListNodes(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListNodes", reflect.TypeOf((*MockStore)(nil).ListNodes), ctx)
}

// ListRunners mocks base method.
func (m *MockStore) ListRunners(ctx context.Context) ([]*models.Runner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRunners", ctx)
	ret0, _ := ret[0].([]*models.Runner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRunners indicates an expected call of ListRunners.
func (mr *MockStoreMockRecorder) ListRunners(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRunners", reflect.TypeOf((*MockStore)(nil).ListRunners), ctx)
}

// ReleaseNodeResources mocks base method.
func (m *MockStore) ReleaseNodeResources(ctx context.Context, id string, cpu, mem int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReleaseNodeResources", ctx, id, cpu, mem)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReleaseNodeResources indicates an expected call of ReleaseNodeResources.
func (mr *MockStoreMockRecorder) ReleaseNodeResources(ctx, id, cpu, mem any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseNodeResources", reflect.TypeOf((*MockStore)(nil).ReleaseNodeResources), ctx, id, cpu, mem)
}

// ReserveNodeResources mocks base method.
func (m *MockStore) ReserveNodeResources(ctx context.Context, id string, cpu, mem int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReserveNodeResources", ctx, id, cpu, mem)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReserveNodeResources indicates an expected call of ReserveNodeResources.
func (mr *MockStoreMockRecorder) ReserveNodeResources(ctx, id, cpu, mem any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReserveNodeResources", reflect.TypeOf((*MockStore)(nil).ReserveNodeResources), ctx, id, cpu, mem)
}

// SaveFlavor mocks base method.
func (m *MockStore) SaveFlavor(ctx context.Context, flavor *models.Flavor) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveFlavor", ctx, flavor)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveFlavor indicates an expected call of SaveFlavor.
func (mr *MockStoreMockRecorder) SaveFlavor(ctx, flavor any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveFlavor", reflect.TypeOf((*MockStore)(nil).SaveFlavor), ctx, flavor)
}

// SaveGroup mocks base method.
func (m *MockStore) SaveGroup(ctx context.Context, group *models.Group) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGroup", ctx, group)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveGroup indicates an expected call of SaveGroup.
func (mr *MockStoreMockRecorder) SaveGroup(ctx, group any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGroup", reflect.TypeOf((*MockStore)(nil).SaveGroup), ctx, group)
}

// SaveImage mocks base method.
func (m *MockStore) SaveImage(ctx context.Context, image *models.Image) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveImage", ctx, image)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveImage indicates an expected call of SaveImage.
func (mr *MockStoreMockRecorder) SaveImage(ctx, image any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveImage", reflect.TypeOf((*MockStore)(nil).SaveImage), ctx, image)
}

// SaveJob mocks base method.
func (m *MockStore) SaveJob(ctx context.Context, job *models.Job) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveJob", ctx, job)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveJob indicates an expected call of SaveJob.
func (mr *MockStoreMockRecorder) SaveJob(ctx, job any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveJob", reflect.TypeOf((*MockStore)(nil).SaveJob), ctx, job)
}

// SaveNode mocks base method.
func (m *MockStore) SaveNode(ctx context.Context, node *models.Node) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveNode", ctx, node)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveNode indicates an expected call of SaveNode.
func (mr *MockStoreMockRecorder) SaveNode(ctx, node any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveNode", reflect.TypeOf((*MockStore)(nil).SaveNode), ctx, node)
}

// SaveRunner mocks base method.
func (m *MockStore) SaveRunner(ctx context.Context, runner *models.Runner) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveRunner", ctx, runner)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveRunner indicates an expected call of SaveRunner.
func (mr *MockStoreMockRecorder) SaveRunner(ctx, runner any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveRunner", reflect.TypeOf((*MockStore)(nil).SaveRunner), ctx, runner)
}
