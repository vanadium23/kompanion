// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces.go

// Package progress_test is a generated GoMock package.
package sync_test

import (
	context "context"
	reflect "reflect"

	entity "gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
	gomock "github.com/golang/mock/gomock"
)

// MockProgressRepo is a mock of ProgressRepo interface.
type MockProgressRepo struct {
	ctrl     *gomock.Controller
	recorder *MockProgressRepoMockRecorder
}

// MockProgressRepoMockRecorder is the mock recorder for MockProgressRepo.
type MockProgressRepoMockRecorder struct {
	mock *MockProgressRepo
}

// NewMockProgressRepo creates a new mock instance.
func NewMockProgressRepo(ctrl *gomock.Controller) *MockProgressRepo {
	mock := &MockProgressRepo{ctrl: ctrl}
	mock.recorder = &MockProgressRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProgressRepo) EXPECT() *MockProgressRepoMockRecorder {
	return m.recorder
}

// GetBookHistory mocks base method.
func (m *MockProgressRepo) GetBookHistory(ctx context.Context, bookID string, limit int) ([]entity.Progress, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBookHistory", ctx, bookID, limit)
	ret0, _ := ret[0].([]entity.Progress)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBookHistory indicates an expected call of GetBookHistory.
func (mr *MockProgressRepoMockRecorder) GetBookHistory(ctx, bookID, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBookHistory", reflect.TypeOf((*MockProgressRepo)(nil).GetBookHistory), ctx, bookID, limit)
}

// Store mocks base method.
func (m *MockProgressRepo) Store(ctx context.Context, t entity.Progress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", ctx, t)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockProgressRepoMockRecorder) Store(ctx, t interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockProgressRepo)(nil).Store), ctx, t)
}

// MockProgress is a mock of Progress interface.
type MockProgress struct {
	ctrl     *gomock.Controller
	recorder *MockProgressMockRecorder
}

// MockProgressMockRecorder is the mock recorder for MockProgress.
type MockProgressMockRecorder struct {
	mock *MockProgress
}

// NewMockProgress creates a new mock instance.
func NewMockProgress(ctrl *gomock.Controller) *MockProgress {
	mock := &MockProgress{ctrl: ctrl}
	mock.recorder = &MockProgressMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProgress) EXPECT() *MockProgressMockRecorder {
	return m.recorder
}

// Fetch mocks base method.
func (m *MockProgress) Fetch(ctx context.Context, bookID string) (entity.Progress, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fetch", ctx, bookID)
	ret0, _ := ret[0].(entity.Progress)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Fetch indicates an expected call of Fetch.
func (mr *MockProgressMockRecorder) Fetch(ctx, bookID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fetch", reflect.TypeOf((*MockProgress)(nil).Fetch), ctx, bookID)
}

// Sync mocks base method.
func (m *MockProgress) Sync(arg0 context.Context, arg1 entity.Progress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockProgressMockRecorder) Sync(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockProgress)(nil).Sync), arg0, arg1)
}
