package hookhandler
//has necessary functions for running hookhandler in dev mode

import (
	"bitbucket.org/level11consulting/go-til/consul"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"
	"io/ioutil"
	"os"
)

type MockHookHandlerContext struct {
	HookHandlerContext //embedding hookhandler context cause didn't want to stub out getters/setters
}

func (hhc *MockHookHandlerContext) GetBitbucketClient(cfg *models.VCSCreds) (handler.VCSHandler, string, error) {
	mockVcsHandler := &MockVCSHandler{}
	return mockVcsHandler , "", nil
}

func (hhc *MockHookHandlerContext) GetRemoteConfig() cred.CVRemoteConfig {
	return &MockRemoteConfig{}
}

////mock remote config/////

type MockRemoteConfig struct {}

func (mrc *MockRemoteConfig) GetConsul()	*consul.Consulet {
	return nil
}
func (mrc *MockRemoteConfig) SetConsul(consul *consul.Consulet) {}
func (mrc *MockRemoteConfig) GetVault() ocevault.Vaulty {
	return &MockVaulty{}
}

func (mrc *MockRemoteConfig) SetVault(vault ocevault.Vaulty) {}
func (mrc *MockRemoteConfig) GetCredAt(path string, hideSecret bool, rcc cred.RemoteConfigCred) (map[string]cred.RemoteConfigCred, error) {
	mockMap := make(map[string]cred.RemoteConfigCred)
	mockVcsCreds := &models.VCSCreds{}

	mockMap["bitbucket/mariannefeng"] = mockVcsCreds
	return mockMap, nil
}
func (mrc *MockRemoteConfig) GetPassword(path string) (string, error) {
	return "", nil
}
func (mrc *MockRemoteConfig) AddCreds(path string, anyCred cred.RemoteConfigCred) (err error) {
	return nil
}

func (mrc *MockRemoteConfig) GetStorageCreds(typ storage.Dest) (*cred.StorageCreds, error) {
	return &cred.StorageCreds{
		DbName: "postgres",
		Location: "localhost",
		User: "postgres",
		Port: 5432,
		Password: "mysecretpassword",
	}, nil
}

func (mrc *MockRemoteConfig) AddSSHKey(path string, sshKeyFile []byte) error {
	return nil
}

func (mrc *MockRemoteConfig) CheckSSHKeyExists(path string) (error) {
	return nil
}

func (mrc *MockRemoteConfig) GetStorageType() (storage.Dest, error) {
	return storage.Postgres, nil
	//return storage.FileSystem, nil
}

func (mrc *MockRemoteConfig) GetOcelotStorage() (storage.OcelotStorage, error) {
	typ, err := mrc.GetStorageType()
	if err != nil {
		return nil, err
	}
	if typ == storage.Postgres {
		fmt.Println("postgres storage")
	}
	creds, err := mrc.GetStorageCreds(typ)
	if err != nil {
		return nil, err
	}
	switch typ {
	case storage.FileSystem:
		return storage.NewFileBuildStorage(creds.Location), nil
	case storage.Postgres:
		return storage.NewPostgresStorage(creds.User, creds.Password, creds.Location, creds.Port, creds.DbName), nil
	default:
		return nil, errors.New("unknown type")
	}
	return nil, errors.New("could not grab ocelot storage")
}

////mock vault////

type MockVaulty struct {}

func (mv *MockVaulty) AddUserAuthData(user string, data map[string]interface{}) (*api.Secret, error) {
	return nil, nil
}

func (mv *MockVaulty) GetAddress() string{
	return ""
}

func (mv *MockVaulty) DeleteSecret(path string) error {
	return nil
}

func (mv *MockVaulty) GetUserAuthData(user string) (map[string]interface{}, error) {
	mockMap := make(map[string]interface{})
	return mockMap, nil
}

func (mv *MockVaulty) AddVaultData(path string, data map[string]interface{}) (*api.Secret, error) {
	return nil, nil
}

func (mv *MockVaulty) GetVaultData(user string) (map[string]interface{}, error) {
	mockMap := make(map[string]interface{})
	return mockMap, nil
}

func (mv *MockVaulty) CreateToken(request *api.TokenCreateRequest) (token string, err error) {
	return "", nil
}
func (mv *MockVaulty) CreateThrowawayToken() (token string, err error) {
	return "", nil
}
func (mv *MockVaulty) CreateVaultPolicy() error {
	return nil
}

////mock vcs handler////

type MockVCSHandler struct {}

func (mvh *MockVCSHandler) Walk() error {
	return nil
}

func (mvh *MockVCSHandler) GetHashDetail(acctRepo, hash string) (pb.PaginatedRepository_RepositoryValues, error) {
	return pb.PaginatedRepository_RepositoryValues{}, nil
}

func (mvh *MockVCSHandler) GetRepoDetail(acctRepo string) (pb.PaginatedRepository_RepositoryValues, error) {
	return pb.PaginatedRepository_RepositoryValues{}, nil
}

//****WARNING**** this assumes you're running this inside of the hookhandler folder
func (mvh *MockVCSHandler) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	pwd, _ := os.Getwd()
	return ioutil.ReadFile(pwd + "/test-fixtures/dev-ocelot.yml")
}

func (mvh *MockVCSHandler) CreateWebhook(webhookURL string) error {
	return nil
}

func (mvh *MockVCSHandler) GetCallbackURL() string {
	return ""
}

func (mvh *MockVCSHandler) SetCallbackURL(callbackURL string) {}

func (mvh *MockVCSHandler) SetBaseURL(baseURL string) {}

func (mvh *MockVCSHandler) GetBaseURL() string {
	return ""
}

func (mvh *MockVCSHandler) FindWebhooks(getWebhookURL string) bool {
	return true
}