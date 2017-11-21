package consulet

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/shankj3/ocelot/util/ocelog"
	"strconv"
	"strings"
)

//Consulet is a wrapper for interfacing with consul
type Consulet struct {
	Client *api.Client
	Config *api.Config
	Connected bool
}

//Default will assume consul is running at localhost:8500
func Default() *Consulet {
	consulet := &Consulet{}
	consulet.Config = api.DefaultConfig()
	c, err := api.NewClient(consulet.Config)

	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return nil
	}
	consulet.Client = c
	consulet.checkIfConnected()
	return consulet
}

//New allows for configuration of consul host + port
func New(consulHost string, consulPort int) *Consulet {
	consulet := &Consulet{}

	consulet.Config = &api.Config{
		Address: consulHost + ":" + strconv.Itoa(consulPort),
	}
	c, err := api.NewClient(consulet.Config)

	if err != nil {
		ocelog.IncludeErrField(err).Error()
	}

	consulet.Client = c
	consulet.checkIfConnected()
	return consulet
}

//RegisterService registers a service at specified host, port, with name
func (consul *Consulet) RegisterService(addr string, port int, name string) (err error) {
	reg := &api.AgentServiceRegistration{
		ID:   name,
		Name: name,
		Port: port,
	}
	err = consul.Client.Agent().ServiceRegister(reg)
	consul.updateConnection(err)
	return
}

//RemoveService removes a service by name
func (consul *Consulet) RemoveService(name string) error {
	err := consul.Client.Agent().ServiceDeregister(name)
	consul.updateConnection(err)
	return err
}

//TODO: should key value operations be atomic??? Can switch to use CAS
func (consul *Consulet) AddKeyValue(key string, value []byte) {
	kv := consul.Client.KV()
	kvPair := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := kv.Put(kvPair, nil)
	consul.updateConnection(err)
}

//RemoveValue removes value at specified key
func (consul *Consulet) RemoveValue(key string) {
	kv := consul.Client.KV()
	_, err := kv.Delete(key, nil)
	consul.updateConnection(err)
}

//GetKeyValue gets key/value at specified key
func (consul *Consulet) GetKeyValue(key string) *api.KVPair {
	kv := consul.Client.KV()
	val, _, err := kv.Get(key, nil)
	consul.updateConnection(err)
	return val
}

//GetKeyValue gets key/value list at specified prefix
func (consul *Consulet) GetKeyValues(prefix string) api.KVPairs {
	kv := consul.Client.KV()
	val, _, err := kv.List(prefix, nil)
	consul.updateConnection(err)
	return val
}

func (consul *Consulet) CreateNewSemaphore(path string, limit int) (*api.Semaphore, error) {
	sessionName := fmt.Sprintf("semaphore_%s", path)
	// create new session. the health check is just gossip failure detector, session will
	// be held as long as the default serf health check hasn't declared node unhealthy.
	// if that node is unhealthy, it probably won't be able to finish running the build so someone
	// else can pick it up... sidenote.. we need to handle if a worker goes down.
	sessionId, meta, err := consul.Client.Session().Create(&api.SessionEntry{
		Name: sessionName,
	}, nil)
	ocelog.Log().Info("meta: ", meta)
	if err != nil {
		consul.updateConnection(err)
		return nil, err
	}
	semaphoreOpts := &api.SemaphoreOptions{
		Prefix:      path,
		Limit:       limit,
		Session:     sessionId,
		SessionName: sessionName,
	}
	sema, err := consul.Client.SemaphoreOpts(semaphoreOpts)
	if err != nil {
		consul.updateConnection(err)
		return nil, err
	}
	return sema, nil
}

//checkIfConnected is called inside of initilization functions to properly update
//the Connected bool flag. It just tries to read at a key that shouldn't exist
func (consul *Consulet) checkIfConnected() {
	consul.GetKeyValue("OCELOT-TEST")
}

//updateConnection takes in an error message and will update the Connection bool to be false if we get connection refused error.
//Also logs error if exists. TODO? Can probably add a new field to Consulet struct that shows most recent failure's error message
func (consul *Consulet) updateConnection(err error) {
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		if strings.Contains(err.Error(), "getsockopt: connection refused") {
			consul.Connected = false
		}
	} else {
		consul.Connected = true
	}
}