package consulet

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/shankj3/ocelot/util/ocelog"
	"strconv"
)

type Consulet struct {
	Client *api.Client
	Config *api.Config
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
	return consulet
}

//RegisterService registers a service at specified host, port, with name
func (consul Consulet) RegisterService(addr string, port int, name string) error {
	reg := &api.AgentServiceRegistration{
		ID:   name,
		Name: name,
		Port: port,
	}
	return consul.Client.Agent().ServiceRegister(reg)
}

//RemoveService removes a service by name
func (consul Consulet) RemoveService(name string) error {
	return consul.Client.Agent().ServiceDeregister(name)
}

//TODO: should key value operations be atomic??? Can switch to use CAS
func (consul Consulet) AddKeyValue(key string, value []byte) {
	kv := consul.Client.KV()
	kvPair := &api.KVPair{
		Key:   key,
		Value: value,
	}
	_, err := kv.Put(kvPair, nil)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
	}
}

//RemoveValue removes value at specified key
func (consul Consulet) RemoveValue(key string) {
	kv := consul.Client.KV()
	kv.Delete(key, nil)
}

//GetKeyValue gets key/value at specified key
func (consul Consulet) GetKeyValue(key string) *api.KVPair {
	kv := consul.Client.KV()
	val, _, err := kv.Get(key, nil)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
	}
	return val
}

//GetKeyValue gets key/value list at specified prefix
func (consul Consulet) GetKeyValues(prefix string) api.KVPairs {
	kv := consul.Client.KV()
	val, _, err := kv.List(prefix, nil)
	if err != nil {
		ocelog.IncludeErrField(err)
	}
	return val
}

func (consul Consulet) CreateNewSemaphore(path string, limit int) (*api.Semaphore, error) {
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
		return nil, err
	}
	return sema, nil
}
