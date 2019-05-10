// Code generated by MockGen. DO NOT EDIT.
// Source: remoteconfigcred_interface.go

// Package credentials is a generated GoMock package.
package config

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pb "github.com/level11consulting/orbitalci/models/pb"
	storage "github.com/level11consulting/orbitalci/storage"
	consul "github.com/shankj3/go-til/consul"
	vault "github.com/shankj3/go-til/vault"
)

// MockRemoteConfigCred is a mock of RemoteConfigCred interface
type MockRemoteConfigCred struct {
	ctrl     *gomock.Controller
	recorder *MockRemoteConfigCredMockRecorder
}

// MockRemoteConfigCredMockRecorder is the mock recorder for MockRemoteConfigCred
type MockRemoteConfigCredMockRecorder struct {
	mock *MockRemoteConfigCred
}

// NewMockRemoteConfigCred creates a new mock instance
func NewMockRemoteConfigCred(ctrl *gomock.Controller) *MockRemoteConfigCred {
	mock := &MockRemoteConfigCred{ctrl: ctrl}
	mock.recorder = &MockRemoteConfigCredMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRemoteConfigCred) EXPECT() *MockRemoteConfigCredMockRecorder {
	return m.recorder
}

// GetClientSecret mocks base method
func (m *MockRemoteConfigCred) GetClientSecret() string {
	ret := m.ctrl.Call(m, "GetClientSecret")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetClientSecret indicates an expected call of GetClientSecret
func (mr *MockRemoteConfigCredMockRecorder) GetClientSecret() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClientSecret", reflect.TypeOf((*MockRemoteConfigCred)(nil).GetClientSecret))
}

// SetAcctNameAndType mocks base method
func (m *MockRemoteConfigCred) SetAcctNameAndType(name, typ string) {
	m.ctrl.Call(m, "SetAcctNameAndType", name, typ)
}

// SetAcctNameAndType indicates an expected call of SetAcctNameAndType
func (mr *MockRemoteConfigCredMockRecorder) SetAcctNameAndType(name, typ interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAcctNameAndType", reflect.TypeOf((*MockRemoteConfigCred)(nil).SetAcctNameAndType), name, typ)
}

// GetAcctName mocks base method
func (m *MockRemoteConfigCred) GetAcctName() string {
	ret := m.ctrl.Call(m, "GetAcctName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetAcctName indicates an expected call of GetAcctName
func (mr *MockRemoteConfigCredMockRecorder) GetAcctName() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAcctName", reflect.TypeOf((*MockRemoteConfigCred)(nil).GetAcctName))
}

// GetType mocks base method
func (m *MockRemoteConfigCred) GetType() string {
	ret := m.ctrl.Call(m, "GetType")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetType indicates an expected call of GetType
func (mr *MockRemoteConfigCredMockRecorder) GetType() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetType", reflect.TypeOf((*MockRemoteConfigCred)(nil).GetType))
}

// SetSecret mocks base method
func (m *MockRemoteConfigCred) SetSecret(arg0 string) {
	m.ctrl.Call(m, "SetSecret", arg0)
}

// SetSecret indicates an expected call of SetSecret
func (mr *MockRemoteConfigCredMockRecorder) SetSecret(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSecret", reflect.TypeOf((*MockRemoteConfigCred)(nil).SetSecret), arg0)
}

// SetAdditionalFields mocks base method
func (m *MockRemoteConfigCred) SetAdditionalFields(key, val string) {
	m.ctrl.Call(m, "SetAdditionalFields", key, val)
}

// SetAdditionalFields indicates an expected call of SetAdditionalFields
func (mr *MockRemoteConfigCredMockRecorder) SetAdditionalFields(key, val interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAdditionalFields", reflect.TypeOf((*MockRemoteConfigCred)(nil).SetAdditionalFields), key, val)
}

// AddAdditionalFields mocks base method
func (m *MockRemoteConfigCred) AddAdditionalFields(consule *consul.Consulet, path string) error {
	ret := m.ctrl.Call(m, "AddAdditionalFields", consule, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddAdditionalFields indicates an expected call of AddAdditionalFields
func (mr *MockRemoteConfigCredMockRecorder) AddAdditionalFields(consule, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAdditionalFields", reflect.TypeOf((*MockRemoteConfigCred)(nil).AddAdditionalFields), consule, path)
}

// BuildCredPath mocks base method
func (m *MockRemoteConfigCred) BuildCredPath(credType, acctName string) string {
	ret := m.ctrl.Call(m, "BuildCredPath", credType, acctName)
	ret0, _ := ret[0].(string)
	return ret0
}

// BuildCredPath indicates an expected call of BuildCredPath
func (mr *MockRemoteConfigCredMockRecorder) BuildCredPath(credType, acctName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildCredPath", reflect.TypeOf((*MockRemoteConfigCred)(nil).BuildCredPath), credType, acctName)
}

// Spawn mocks base method
func (m *MockRemoteConfigCred) Spawn() RemoteConfigCred {
	ret := m.ctrl.Call(m, "Spawn")
	ret0, _ := ret[0].(RemoteConfigCred)
	return ret0
}

// Spawn indicates an expected call of Spawn
func (mr *MockRemoteConfigCredMockRecorder) Spawn() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Spawn", reflect.TypeOf((*MockRemoteConfigCred)(nil).Spawn))
}

// MockStorageCred is a mock of StorageCred interface
type MockStorageCred struct {
	ctrl     *gomock.Controller
	recorder *MockStorageCredMockRecorder
}

// MockStorageCredMockRecorder is the mock recorder for MockStorageCred
type MockStorageCredMockRecorder struct {
	mock *MockStorageCred
}

// NewMockStorageCred creates a new mock instance
func NewMockStorageCred(ctrl *gomock.Controller) *MockStorageCred {
	mock := &MockStorageCred{ctrl: ctrl}
	mock.recorder = &MockStorageCredMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStorageCred) EXPECT() *MockStorageCredMockRecorder {
	return m.recorder
}

// GetStorageCreds mocks base method
func (m *MockStorageCred) GetStorageCreds(typ *storage.Dest) (*StorageCreds, error) {
	ret := m.ctrl.Call(m, "GetStorageCreds", typ)
	ret0, _ := ret[0].(*StorageCreds)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStorageCreds indicates an expected call of GetStorageCreds
func (mr *MockStorageCredMockRecorder) GetStorageCreds(typ interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStorageCreds", reflect.TypeOf((*MockStorageCred)(nil).GetStorageCreds), typ)
}

// GetStorageType mocks base method
func (m *MockStorageCred) GetStorageType() (storage.Dest, error) {
	ret := m.ctrl.Call(m, "GetStorageType")
	ret0, _ := ret[0].(storage.Dest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStorageType indicates an expected call of GetStorageType
func (mr *MockStorageCredMockRecorder) GetStorageType() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStorageType", reflect.TypeOf((*MockStorageCred)(nil).GetStorageType))
}

// GetOcelotStorage mocks base method
func (m *MockStorageCred) GetOcelotStorage() (storage.OcelotStorage, error) {
	ret := m.ctrl.Call(m, "GetOcelotStorage")
	ret0, _ := ret[0].(storage.OcelotStorage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOcelotStorage indicates an expected call of GetOcelotStorage
func (mr *MockStorageCredMockRecorder) GetOcelotStorage() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOcelotStorage", reflect.TypeOf((*MockStorageCred)(nil).GetOcelotStorage))
}

// MockHealthyMaintainer is a mock of HealthyMaintainer interface
type MockHealthyMaintainer struct {
	ctrl     *gomock.Controller
	recorder *MockHealthyMaintainerMockRecorder
}

// MockHealthyMaintainerMockRecorder is the mock recorder for MockHealthyMaintainer
type MockHealthyMaintainerMockRecorder struct {
	mock *MockHealthyMaintainer
}

// NewMockHealthyMaintainer creates a new mock instance
func NewMockHealthyMaintainer(ctrl *gomock.Controller) *MockHealthyMaintainer {
	mock := &MockHealthyMaintainer{ctrl: ctrl}
	mock.recorder = &MockHealthyMaintainerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHealthyMaintainer) EXPECT() *MockHealthyMaintainerMockRecorder {
	return m.recorder
}

// Reconnect mocks base method
func (m *MockHealthyMaintainer) Reconnect() error {
	ret := m.ctrl.Call(m, "Reconnect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconnect indicates an expected call of Reconnect
func (mr *MockHealthyMaintainerMockRecorder) Reconnect() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconnect", reflect.TypeOf((*MockHealthyMaintainer)(nil).Reconnect))
}

// Healthy mocks base method
func (m *MockHealthyMaintainer) Healthy() bool {
	ret := m.ctrl.Call(m, "Healthy")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Healthy indicates an expected call of Healthy
func (mr *MockHealthyMaintainerMockRecorder) Healthy() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Healthy", reflect.TypeOf((*MockHealthyMaintainer)(nil).Healthy))
}

// MockCVRemoteConfig is a mock of CVRemoteConfig interface
type MockCVRemoteConfig struct {
	ctrl     *gomock.Controller
	recorder *MockCVRemoteConfigMockRecorder
}

// MockCVRemoteConfigMockRecorder is the mock recorder for MockCVRemoteConfig
type MockCVRemoteConfigMockRecorder struct {
	mock *MockCVRemoteConfig
}

// NewMockCVRemoteConfig creates a new mock instance
func NewMockCVRemoteConfig(ctrl *gomock.Controller) *MockCVRemoteConfig {
	mock := &MockCVRemoteConfig{ctrl: ctrl}
	mock.recorder = &MockCVRemoteConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCVRemoteConfig) EXPECT() *MockCVRemoteConfigMockRecorder {
	return m.recorder
}

// GetConsul mocks base method
func (m *MockCVRemoteConfig) GetConsul() consul.Consuletty {
	ret := m.ctrl.Call(m, "GetConsul")
	ret0, _ := ret[0].(consul.Consuletty)
	return ret0
}

// GetConsul indicates an expected call of GetConsul
func (mr *MockCVRemoteConfigMockRecorder) GetConsul() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConsul", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetConsul))
}

// SetConsul mocks base method
func (m *MockCVRemoteConfig) SetConsul(consul consul.Consuletty) {
	m.ctrl.Call(m, "SetConsul", consul)
}

// SetConsul indicates an expected call of SetConsul
func (mr *MockCVRemoteConfigMockRecorder) SetConsul(consul interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetConsul", reflect.TypeOf((*MockCVRemoteConfig)(nil).SetConsul), consul)
}

// GetVault mocks base method
func (m *MockCVRemoteConfig) GetVault() vault.Vaulty {
	ret := m.ctrl.Call(m, "GetVault")
	ret0, _ := ret[0].(vault.Vaulty)
	return ret0
}

// GetVault indicates an expected call of GetVault
func (mr *MockCVRemoteConfigMockRecorder) GetVault() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVault", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetVault))
}

// SetVault mocks base method
func (m *MockCVRemoteConfig) SetVault(vault vault.Vaulty) {
	m.ctrl.Call(m, "SetVault", vault)
}

// SetVault indicates an expected call of SetVault
func (mr *MockCVRemoteConfigMockRecorder) SetVault(vault interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetVault", reflect.TypeOf((*MockCVRemoteConfig)(nil).SetVault), vault)
}

// AddSSHKey mocks base method
func (m *MockCVRemoteConfig) AddSSHKey(path string, sshKeyFile []byte) error {
	ret := m.ctrl.Call(m, "AddSSHKey", path, sshKeyFile)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddSSHKey indicates an expected call of AddSSHKey
func (mr *MockCVRemoteConfigMockRecorder) AddSSHKey(path, sshKeyFile interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSSHKey", reflect.TypeOf((*MockCVRemoteConfig)(nil).AddSSHKey), path, sshKeyFile)
}

// CheckSSHKeyExists mocks base method
func (m *MockCVRemoteConfig) CheckSSHKeyExists(path string) error {
	ret := m.ctrl.Call(m, "CheckSSHKeyExists", path)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckSSHKeyExists indicates an expected call of CheckSSHKeyExists
func (mr *MockCVRemoteConfigMockRecorder) CheckSSHKeyExists(path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckSSHKeyExists", reflect.TypeOf((*MockCVRemoteConfig)(nil).CheckSSHKeyExists), path)
}

// GetPassword mocks base method
func (m *MockCVRemoteConfig) GetPassword(scType pb.SubCredType, acctName string, ocyCredType pb.CredType, identifier string) (string, error) {
	ret := m.ctrl.Call(m, "GetPassword", scType, acctName, ocyCredType, identifier)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPassword indicates an expected call of GetPassword
func (mr *MockCVRemoteConfigMockRecorder) GetPassword(scType, acctName, ocyCredType, identifier interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPassword", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetPassword), scType, acctName, ocyCredType, identifier)
}

// DeleteCred mocks base method
func (m *MockCVRemoteConfig) DeleteCred(store storage.CredTable, anyCred pb.OcyCredder) error {
	ret := m.ctrl.Call(m, "DeleteCred", store, anyCred)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCred indicates an expected call of DeleteCred
func (mr *MockCVRemoteConfigMockRecorder) DeleteCred(store, anyCred interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCred", reflect.TypeOf((*MockCVRemoteConfig)(nil).DeleteCred), store, anyCred)
}

// GetCredsByType mocks base method
func (m *MockCVRemoteConfig) GetCredsByType(store storage.CredTable, ctype pb.CredType, hideSecret bool) ([]pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetCredsByType", store, ctype, hideSecret)
	ret0, _ := ret[0].([]pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCredsByType indicates an expected call of GetCredsByType
func (mr *MockCVRemoteConfigMockRecorder) GetCredsByType(store, ctype, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCredsByType", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetCredsByType), store, ctype, hideSecret)
}

// GetAllCreds mocks base method
func (m *MockCVRemoteConfig) GetAllCreds(store storage.CredTable, hideSecret bool) ([]pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetAllCreds", store, hideSecret)
	ret0, _ := ret[0].([]pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllCreds indicates an expected call of GetAllCreds
func (mr *MockCVRemoteConfigMockRecorder) GetAllCreds(store, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllCreds", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetAllCreds), store, hideSecret)
}

// GetCred mocks base method
func (m *MockCVRemoteConfig) GetCred(store storage.CredTable, subCredType pb.SubCredType, identifier, accountName string, hideSecret bool) (pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetCred", store, subCredType, identifier, accountName, hideSecret)
	ret0, _ := ret[0].(pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCred indicates an expected call of GetCred
func (mr *MockCVRemoteConfigMockRecorder) GetCred(store, subCredType, identifier, accountName, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCred", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetCred), store, subCredType, identifier, accountName, hideSecret)
}

// GetCredsBySubTypeAndAcct mocks base method
func (m *MockCVRemoteConfig) GetCredsBySubTypeAndAcct(store storage.CredTable, stype pb.SubCredType, accountName string, hideSecret bool) ([]pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetCredsBySubTypeAndAcct", store, stype, accountName, hideSecret)
	ret0, _ := ret[0].([]pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCredsBySubTypeAndAcct indicates an expected call of GetCredsBySubTypeAndAcct
func (mr *MockCVRemoteConfigMockRecorder) GetCredsBySubTypeAndAcct(store, stype, accountName, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCredsBySubTypeAndAcct", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetCredsBySubTypeAndAcct), store, stype, accountName, hideSecret)
}

// AddCreds mocks base method
func (m *MockCVRemoteConfig) AddCreds(store storage.CredTable, anyCred pb.OcyCredder, overwriteOk bool) error {
	ret := m.ctrl.Call(m, "AddCreds", store, anyCred, overwriteOk)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCreds indicates an expected call of AddCreds
func (mr *MockCVRemoteConfigMockRecorder) AddCreds(store, anyCred, overwriteOk interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCreds", reflect.TypeOf((*MockCVRemoteConfig)(nil).AddCreds), store, anyCred, overwriteOk)
}

// UpdateCreds mocks base method
func (m *MockCVRemoteConfig) UpdateCreds(store storage.CredTable, anyCred pb.OcyCredder) error {
	ret := m.ctrl.Call(m, "UpdateCreds", store, anyCred)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCreds indicates an expected call of UpdateCreds
func (mr *MockCVRemoteConfigMockRecorder) UpdateCreds(store, anyCred interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCreds", reflect.TypeOf((*MockCVRemoteConfig)(nil).UpdateCreds), store, anyCred)
}

// Reconnect mocks base method
func (m *MockCVRemoteConfig) Reconnect() error {
	ret := m.ctrl.Call(m, "Reconnect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconnect indicates an expected call of Reconnect
func (mr *MockCVRemoteConfigMockRecorder) Reconnect() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconnect", reflect.TypeOf((*MockCVRemoteConfig)(nil).Reconnect))
}

// Healthy mocks base method
func (m *MockCVRemoteConfig) Healthy() bool {
	ret := m.ctrl.Call(m, "Healthy")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Healthy indicates an expected call of Healthy
func (mr *MockCVRemoteConfigMockRecorder) Healthy() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Healthy", reflect.TypeOf((*MockCVRemoteConfig)(nil).Healthy))
}

// GetStorageCreds mocks base method
func (m *MockCVRemoteConfig) GetStorageCreds(typ *storage.Dest) (*StorageCreds, error) {
	ret := m.ctrl.Call(m, "GetStorageCreds", typ)
	ret0, _ := ret[0].(*StorageCreds)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStorageCreds indicates an expected call of GetStorageCreds
func (mr *MockCVRemoteConfigMockRecorder) GetStorageCreds(typ interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStorageCreds", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetStorageCreds), typ)
}

// GetStorageType mocks base method
func (m *MockCVRemoteConfig) GetStorageType() (storage.Dest, error) {
	ret := m.ctrl.Call(m, "GetStorageType")
	ret0, _ := ret[0].(storage.Dest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStorageType indicates an expected call of GetStorageType
func (mr *MockCVRemoteConfigMockRecorder) GetStorageType() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStorageType", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetStorageType))
}

// GetOcelotStorage mocks base method
func (m *MockCVRemoteConfig) GetOcelotStorage() (storage.OcelotStorage, error) {
	ret := m.ctrl.Call(m, "GetOcelotStorage")
	ret0, _ := ret[0].(storage.OcelotStorage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOcelotStorage indicates an expected call of GetOcelotStorage
func (mr *MockCVRemoteConfigMockRecorder) GetOcelotStorage() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOcelotStorage", reflect.TypeOf((*MockCVRemoteConfig)(nil).GetOcelotStorage))
}

// MockInsecureCredStorage is a mock of InsecureCredStorage interface
type MockInsecureCredStorage struct {
	ctrl     *gomock.Controller
	recorder *MockInsecureCredStorageMockRecorder
}

// MockInsecureCredStorageMockRecorder is the mock recorder for MockInsecureCredStorage
type MockInsecureCredStorageMockRecorder struct {
	mock *MockInsecureCredStorage
}

// NewMockInsecureCredStorage creates a new mock instance
func NewMockInsecureCredStorage(ctrl *gomock.Controller) *MockInsecureCredStorage {
	mock := &MockInsecureCredStorage{ctrl: ctrl}
	mock.recorder = &MockInsecureCredStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInsecureCredStorage) EXPECT() *MockInsecureCredStorageMockRecorder {
	return m.recorder
}

// GetCredsByType mocks base method
func (m *MockInsecureCredStorage) GetCredsByType(store storage.CredTable, ctype pb.CredType, hideSecret bool) ([]pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetCredsByType", store, ctype, hideSecret)
	ret0, _ := ret[0].([]pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCredsByType indicates an expected call of GetCredsByType
func (mr *MockInsecureCredStorageMockRecorder) GetCredsByType(store, ctype, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCredsByType", reflect.TypeOf((*MockInsecureCredStorage)(nil).GetCredsByType), store, ctype, hideSecret)
}

// GetAllCreds mocks base method
func (m *MockInsecureCredStorage) GetAllCreds(store storage.CredTable, hideSecret bool) ([]pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetAllCreds", store, hideSecret)
	ret0, _ := ret[0].([]pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllCreds indicates an expected call of GetAllCreds
func (mr *MockInsecureCredStorageMockRecorder) GetAllCreds(store, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllCreds", reflect.TypeOf((*MockInsecureCredStorage)(nil).GetAllCreds), store, hideSecret)
}

// GetCred mocks base method
func (m *MockInsecureCredStorage) GetCred(store storage.CredTable, subCredType pb.SubCredType, identifier, accountName string, hideSecret bool) (pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetCred", store, subCredType, identifier, accountName, hideSecret)
	ret0, _ := ret[0].(pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCred indicates an expected call of GetCred
func (mr *MockInsecureCredStorageMockRecorder) GetCred(store, subCredType, identifier, accountName, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCred", reflect.TypeOf((*MockInsecureCredStorage)(nil).GetCred), store, subCredType, identifier, accountName, hideSecret)
}

// GetCredsBySubTypeAndAcct mocks base method
func (m *MockInsecureCredStorage) GetCredsBySubTypeAndAcct(store storage.CredTable, stype pb.SubCredType, accountName string, hideSecret bool) ([]pb.OcyCredder, error) {
	ret := m.ctrl.Call(m, "GetCredsBySubTypeAndAcct", store, stype, accountName, hideSecret)
	ret0, _ := ret[0].([]pb.OcyCredder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCredsBySubTypeAndAcct indicates an expected call of GetCredsBySubTypeAndAcct
func (mr *MockInsecureCredStorageMockRecorder) GetCredsBySubTypeAndAcct(store, stype, accountName, hideSecret interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCredsBySubTypeAndAcct", reflect.TypeOf((*MockInsecureCredStorage)(nil).GetCredsBySubTypeAndAcct), store, stype, accountName, hideSecret)
}

// AddCreds mocks base method
func (m *MockInsecureCredStorage) AddCreds(store storage.CredTable, anyCred pb.OcyCredder, overwriteOk bool) error {
	ret := m.ctrl.Call(m, "AddCreds", store, anyCred, overwriteOk)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCreds indicates an expected call of AddCreds
func (mr *MockInsecureCredStorageMockRecorder) AddCreds(store, anyCred, overwriteOk interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCreds", reflect.TypeOf((*MockInsecureCredStorage)(nil).AddCreds), store, anyCred, overwriteOk)
}

// UpdateCreds mocks base method
func (m *MockInsecureCredStorage) UpdateCreds(store storage.CredTable, anyCred pb.OcyCredder) error {
	ret := m.ctrl.Call(m, "UpdateCreds", store, anyCred)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCreds indicates an expected call of UpdateCreds
func (mr *MockInsecureCredStorageMockRecorder) UpdateCreds(store, anyCred interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCreds", reflect.TypeOf((*MockInsecureCredStorage)(nil).UpdateCreds), store, anyCred)
}
