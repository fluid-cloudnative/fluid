// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/ddc/base/mock_engine.go

// Package base is a generated GoMock package.
package base

import (
	reflect "reflect"

	v1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	runtime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	gomock "github.com/golang/mock/gomock"
)

// MockEngine is a mock of Engine interface
type MockEngine struct {
	ctrl     *gomock.Controller
	recorder *MockEngineMockRecorder
}

// MockEngineMockRecorder is the mock recorder for MockEngine
type MockEngineMockRecorder struct {
	mock *MockEngine
}

// NewMockEngine creates a new mock instance
func NewMockEngine(ctrl *gomock.Controller) *MockEngine {
	mock := &MockEngine{ctrl: ctrl}
	mock.recorder = &MockEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEngine) EXPECT() *MockEngineMockRecorder {
	return m.recorder
}

// ID mocks base method
func (m *MockEngine) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID
func (mr *MockEngineMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockEngine)(nil).ID))
}

// Shutdown mocks base method
func (m *MockEngine) Shutdown() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown")
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown
func (mr *MockEngineMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockEngine)(nil).Shutdown))
}

// Setup mocks base method
func (m *MockEngine) Setup(ctx runtime.ReconcileRequestContext) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Setup", ctx)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Setup indicates an expected call of Setup
func (mr *MockEngineMockRecorder) Setup(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Setup", reflect.TypeOf((*MockEngine)(nil).Setup), ctx)
}

// CreateVolume mocks base method
func (m *MockEngine) CreateVolume() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateVolume")
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateVolume indicates an expected call of CreateVolume
func (mr *MockEngineMockRecorder) CreateVolume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVolume", reflect.TypeOf((*MockEngine)(nil).CreateVolume))
}

// DeleteVolume mocks base method
func (m *MockEngine) DeleteVolume() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteVolume")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteVolume indicates an expected call of DeleteVolume
func (mr *MockEngineMockRecorder) DeleteVolume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteVolume", reflect.TypeOf((*MockEngine)(nil).DeleteVolume))
}

// Sync mocks base method
func (m *MockEngine) Sync(ctx runtime.ReconcileRequestContext) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync
func (mr *MockEngineMockRecorder) Sync(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockEngine)(nil).Sync), ctx)
}

// MockImplement is a mock of Implement interface
type MockImplement struct {
	ctrl     *gomock.Controller
	recorder *MockImplementMockRecorder
}

func (m *MockImplement) UpdateOnUFSChange() (ready bool, err error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOnUFSChange")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockImplementMockRecorder) UpdateOnUFSChange() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOnUFSChange", reflect.TypeOf((*MockImplement)(nil).UpdateOnUFSChange))
}

func (m *MockImplement) CreateDataLoadJob(ctx runtime.ReconcileRequestContext, targetDataload v1alpha1.DataLoad) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDataLoadJob", ctx, targetDataload)
	ret0, _ := ret[0].(error)
	return ret0
}

func (m *MockImplement) CheckRuntimeReady() (ready bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckRuntimeReady")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (m *MockImplement) CheckExistenceOfPath(targetDataload v1alpha1.DataLoad) (ready bool, err error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckExistenceOfPath", targetDataload)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MockImplementMockRecorder is the mock recorder for MockImplement
type MockImplementMockRecorder struct {
	mock *MockImplement
}

// NewMockImplement creates a new mock instance
func NewMockImplement(ctrl *gomock.Controller) *MockImplement {
	mock := &MockImplement{ctrl: ctrl}
	mock.recorder = &MockImplementMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockImplement) EXPECT() *MockImplementMockRecorder {
	return m.recorder
}

// UsedStorageBytes mocks base method
func (m *MockImplement) UsedStorageBytes() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UsedStorageBytes")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UsedStorageBytes indicates an expected call of UsedStorageBytes
func (mr *MockImplementMockRecorder) UsedStorageBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UsedStorageBytes", reflect.TypeOf((*MockImplement)(nil).UsedStorageBytes))
}

// FreeStorageBytes mocks base method
func (m *MockImplement) FreeStorageBytes() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FreeStorageBytes")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FreeStorageBytes indicates an expected call of FreeStorageBytes
func (mr *MockImplementMockRecorder) FreeStorageBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FreeStorageBytes", reflect.TypeOf((*MockImplement)(nil).FreeStorageBytes))
}

// TotalStorageBytes mocks base method
func (m *MockImplement) TotalStorageBytes() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalStorageBytes")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TotalStorageBytes indicates an expected call of TotalStorageBytes
func (mr *MockImplementMockRecorder) TotalStorageBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalStorageBytes", reflect.TypeOf((*MockImplement)(nil).TotalStorageBytes))
}

// TotalFileNums mocks base method
func (m *MockImplement) TotalFileNums() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalFileNums")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TotalFileNums indicates an expected call of TotalFileNums
func (mr *MockImplementMockRecorder) TotalFileNums() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalFileNums", reflect.TypeOf((*MockImplement)(nil).TotalFileNums))
}

// CheckMasterReady mocks base method
func (m *MockImplement) CheckMasterReady() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckMasterReady")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckMasterReady indicates an expected call of CheckMasterReady
func (mr *MockImplementMockRecorder) CheckMasterReady() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckMasterReady", reflect.TypeOf((*MockImplement)(nil).CheckMasterReady))
}

// CheckWorkersReady mocks base method
func (m *MockImplement) CheckWorkersReady() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckWorkersReady")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckWorkersReady indicates an expected call of CheckWorkersReady
func (mr *MockImplementMockRecorder) CheckWorkersReady() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckWorkersReady", reflect.TypeOf((*MockImplement)(nil).CheckWorkersReady))
}

// IsSetupDone mocks base method
func (m *MockImplement) IsSetupDone() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsSetupDone")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsSetupDone indicates an expected call of IsSetupDone
func (mr *MockImplementMockRecorder) IsSetupDone() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSetupDone", reflect.TypeOf((*MockImplement)(nil).IsSetupDone))
}

// ShouldSetupMaster mocks base method
func (m *MockImplement) ShouldSetupMaster() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShouldSetupMaster")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShouldSetupMaster indicates an expected call of ShouldSetupMaster
func (mr *MockImplementMockRecorder) ShouldSetupMaster() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldSetupMaster", reflect.TypeOf((*MockImplement)(nil).ShouldSetupMaster))
}

// ShouldSetupWorkers mocks base method
func (m *MockImplement) ShouldSetupWorkers() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShouldSetupWorkers")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShouldSetupWorkers indicates an expected call of ShouldSetupWorkers
func (mr *MockImplementMockRecorder) ShouldSetupWorkers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldSetupWorkers", reflect.TypeOf((*MockImplement)(nil).ShouldSetupWorkers))
}

// ShouldCheckUFS mocks base method
func (m *MockImplement) ShouldCheckUFS() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShouldCheckUFS")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShouldCheckUFS indicates an expected call of ShouldCheckUFS
func (mr *MockImplementMockRecorder) ShouldCheckUFS() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldCheckUFS", reflect.TypeOf((*MockImplement)(nil).ShouldCheckUFS))
}

// SetupMaster mocks base method
func (m *MockImplement) SetupMaster() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupMaster")
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupMaster indicates an expected call of SetupMaster
func (mr *MockImplementMockRecorder) SetupMaster() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupMaster", reflect.TypeOf((*MockImplement)(nil).SetupMaster))
}

// SetupWorkers mocks base method
func (m *MockImplement) SetupWorkers() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupWorkers")
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupWorkers indicates an expected call of SetupWorkers
func (mr *MockImplementMockRecorder) SetupWorkers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupWorkers", reflect.TypeOf((*MockImplement)(nil).SetupWorkers))
}

// UpdateDatasetStatus mocks base method
func (m *MockImplement) UpdateDatasetStatus(phase v1alpha1.DatasetPhase) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDatasetStatus", phase)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateDatasetStatus indicates an expected call of UpdateDatasetStatus
func (mr *MockImplementMockRecorder) UpdateDatasetStatus(phase interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDatasetStatus", reflect.TypeOf((*MockImplement)(nil).UpdateDatasetStatus), phase)
}

// BindToDataset mocks base method
func (m *MockImplement) BindToDataset() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BindToDataset")
	ret0, _ := ret[0].(error)
	return ret0
}

// BindToDataset indicates an expected call of BindToDataset
func (mr *MockImplementMockRecorder) BindToDataset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BindToDataset", reflect.TypeOf((*MockImplement)(nil).BindToDataset))
}

// PrepareUFS mocks base method
func (m *MockImplement) PrepareUFS() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareUFS")
	ret0, _ := ret[0].(error)
	return ret0
}

// PrepareUFS indicates an expected call of PrepareUFS
func (mr *MockImplementMockRecorder) PrepareUFS() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareUFS", reflect.TypeOf((*MockImplement)(nil).PrepareUFS))
}

// Shutdown mocks base method
func (m *MockImplement) Shutdown() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown")
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown
func (mr *MockImplementMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockImplement)(nil).Shutdown))
}

// AssignNodesToCache mocks base method
func (m *MockImplement) AssignNodesToCache(desiredNum int32) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AssignNodesToCache", desiredNum)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AssignNodesToCache indicates an expected call of AssignNodesToCache
func (mr *MockImplementMockRecorder) AssignNodesToCache(desiredNum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AssignNodesToCache", reflect.TypeOf((*MockImplement)(nil).AssignNodesToCache), desiredNum)
}

// CheckRuntimeHealthy mocks base method
func (m *MockImplement) CheckRuntimeHealthy() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckRuntimeHealthy")
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckRuntimeHealthy indicates an expected call of CheckRuntimeHealthy
func (mr *MockImplementMockRecorder) CheckRuntimeHealthy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckRuntimeHealthy", reflect.TypeOf((*MockImplement)(nil).CheckRuntimeHealthy))
}

// UpdateCacheOfDataset mocks base method
func (m *MockImplement) UpdateCacheOfDataset() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCacheOfDataset")
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCacheOfDataset indicates an expected call of UpdateCacheOfDataset
func (mr *MockImplementMockRecorder) UpdateCacheOfDataset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCacheOfDataset", reflect.TypeOf((*MockImplement)(nil).UpdateCacheOfDataset))
}

// CheckAndUpdateRuntimeStatus mocks base method
func (m *MockImplement) CheckAndUpdateRuntimeStatus() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckAndUpdateRuntimeStatus")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckAndUpdateRuntimeStatus indicates an expected call of CheckAndUpdateRuntimeStatus
func (mr *MockImplementMockRecorder) CheckAndUpdateRuntimeStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckAndUpdateRuntimeStatus", reflect.TypeOf((*MockImplement)(nil).CheckAndUpdateRuntimeStatus))
}

// CreateVolume mocks base method
func (m *MockImplement) CreateVolume() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateVolume")
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateVolume indicates an expected call of CreateVolume
func (mr *MockImplementMockRecorder) CreateVolume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVolume", reflect.TypeOf((*MockImplement)(nil).CreateVolume))
}

// SyncReplicas mocks base method
func (m *MockImplement) SyncReplicas(ctx runtime.ReconcileRequestContext) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncReplicas", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncReplicas indicates an expected call of SyncReplicas
func (mr *MockImplementMockRecorder) SyncReplicas(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncReplicas", reflect.TypeOf((*MockImplement)(nil).SyncReplicas), ctx)
}

// SyncMetadata mocks base method
func (m *MockImplement) SyncMetadata() (err error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncMetadata")
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncMetadata indicates an expected call of SyncMetadata
func (mr *MockImplementMockRecorder) SyncMetadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncMetadata", reflect.TypeOf((*MockImplement)(nil).SyncMetadata))
}

// DeleteVolume mocks base method
func (m *MockImplement) DeleteVolume() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteVolume")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteVolume indicates an expected call of DeleteVolume
func (mr *MockImplementMockRecorder) DeleteVolume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteVolume", reflect.TypeOf((*MockImplement)(nil).DeleteVolume))
}

// MockUnderFileSystemService is a mock of UnderFileSystemService interface
type MockUnderFileSystemService struct {
	ctrl     *gomock.Controller
	recorder *MockUnderFileSystemServiceMockRecorder
}

// MockUnderFileSystemServiceMockRecorder is the mock recorder for MockUnderFileSystemService
type MockUnderFileSystemServiceMockRecorder struct {
	mock *MockUnderFileSystemService
}

// NewMockUnderFileSystemService creates a new mock instance
func NewMockUnderFileSystemService(ctrl *gomock.Controller) *MockUnderFileSystemService {
	mock := &MockUnderFileSystemService{ctrl: ctrl}
	mock.recorder = &MockUnderFileSystemServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUnderFileSystemService) EXPECT() *MockUnderFileSystemServiceMockRecorder {
	return m.recorder
}

// UsedStorageBytes mocks base method
func (m *MockUnderFileSystemService) UsedStorageBytes() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UsedStorageBytes")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UsedStorageBytes indicates an expected call of UsedStorageBytes
func (mr *MockUnderFileSystemServiceMockRecorder) UsedStorageBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UsedStorageBytes", reflect.TypeOf((*MockUnderFileSystemService)(nil).UsedStorageBytes))
}

// FreeStorageBytes mocks base method
func (m *MockUnderFileSystemService) FreeStorageBytes() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FreeStorageBytes")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FreeStorageBytes indicates an expected call of FreeStorageBytes
func (mr *MockUnderFileSystemServiceMockRecorder) FreeStorageBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FreeStorageBytes", reflect.TypeOf((*MockUnderFileSystemService)(nil).FreeStorageBytes))
}

// TotalStorageBytes mocks base method
func (m *MockUnderFileSystemService) TotalStorageBytes() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalStorageBytes")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TotalStorageBytes indicates an expected call of TotalStorageBytes
func (mr *MockUnderFileSystemServiceMockRecorder) TotalStorageBytes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalStorageBytes", reflect.TypeOf((*MockUnderFileSystemService)(nil).TotalStorageBytes))
}

// TotalFileNums mocks base method
func (m *MockUnderFileSystemService) TotalFileNums() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalFileNums")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TotalFileNums indicates an expected call of TotalFileNums
func (mr *MockUnderFileSystemServiceMockRecorder) TotalFileNums() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalFileNums", reflect.TypeOf((*MockUnderFileSystemService)(nil).TotalFileNums))
}
