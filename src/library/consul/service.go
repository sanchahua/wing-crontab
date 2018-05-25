package consul

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/consul/api"
	"sync"
	"time"
	"errors"
)

// 服务注册
const (
	Registered = 1 << iota
)
const (
	statusOnline    = "online"
	statusOffline   = "offline"
)
var membersEmpty   = errors.New("members is empty")
var leaderNotFound = errors.New("leader not found")

type ServiceMember struct {
	IsLeader bool
	ServiceID string
	Status string
	ServiceIp string
	Port int
}
type Service struct {
	ServiceName string //service name, like: service.add
	ServiceHost string //service host, like: 0.0.0.0, 127.0.0.1
	ServiceIp string // if ServiceHost is 0.0.0.0, ServiceIp must set,
	// like 127.0.0.1 or 192.168.9.12 or 114.55.56.168
	ServicePort int // service port, like: 9998
	Interval time.Duration // interval for update ttl
	Ttl int //check ttl
	ServiceID string //serviceID = fmt.Sprintf("%s-%s-%d", name, ip, port)
	client *api.Client ///consul client
	agent *api.Agent //consul agent
	status int // register status
	lock *sync.Mutex //sync lock
	session *Session
	Kv *api.KV
	health *api.Health
	leader bool
	onleader []OnLeaderFunc
	//lockKey string
	consulLock *Lock
}

type ServiceOption func(s *Service)
type OnLeaderFunc  func(isLeader bool)

// set ttl
func SetTtl(ttl int) ServiceOption {
	return func(s *Service){
		s.Ttl = ttl
	}
}

func SetOnLeader(f ...OnLeaderFunc) ServiceOption {
	return func(s *Service) {
		s.onleader = append(s.onleader, f...)
	}
}

//func SetLockKey(lock *Lock) ServiceOption  {
//	return func(s *Service) {
//		s.consulLock = lock//NewLock(s.session, s.Kv, lockKey)
//		//s.lockKey = lockKey
//	}
//}

// set interval
func SetInterval(interval time.Duration) ServiceOption {
	return func(s *Service){
		s.Interval = interval
	}
}

// new a service
// name: service name
// host: service host like 0.0.0.0 or 127.0.0.1
// port: service port, like 9998
// consulAddress: consul service address, like 127.0.0.1:8500
// opts: ServiceOption, like ServiceIp("127.0.0.1")
// return new service pointer
func NewService(
	address string, //127.0.0.1:8500
	lockKey string,
	name string,
	host string,
	port int,
	opts ...ServiceOption,
) *Service {

	consulConfig        := api.DefaultConfig()
	consulConfig.Address = address//ctx.Config.ConsulAddress
	c, err         := api.NewClient(consulConfig)

	if err != nil {
		log.Panicf("%v", err)
	}

	sev := &Service{
		ServiceName : name,
		ServiceHost : host,
		ServicePort : port,
		Interval    : time.Second * 3,
		Ttl         : 15,
		status      : 0,
		leader      : false,
		lock        : new(sync.Mutex),
		consulLock  : nil,//NewLock(),
	}
	sev.client    = c
	sev.Kv        = c.KV()
	sev.session   = NewSession(c.Session(), 10)
	for _, opt := range opts {
		opt(sev)
	}
	sev.consulLock = NewLock(sev.session, sev.Kv, lockKey)
	sev.ServiceID = fmt.Sprintf("%s-%s-%d", name, host, port)
	sev.agent     = sev.client.Agent()
	sev.health    = sev.client.Health()
	go sev.check()
	return sev
}

func (sev *Service) Deregister() error {
	err := sev.agent.ServiceDeregister(sev.ServiceID)
	if err != nil {
		log.Errorf("deregister service error: ", err.Error())
		return err
	}
	err = sev.agent.CheckDeregister(sev.ServiceID)
	if err != nil {
		log.Println("deregister check error: ", err.Error())
	}
	return err
}

func (sev *Service) updateTtl() {
	if sev.status & Registered <= 0 {
		return
	}
	//log.Debugf("current node %v:%v is leader=%v", sev.ServiceIp, sev.ServicePort, sev.leader)
	err := sev.agent.UpdateTTL(sev.ServiceID, fmt.Sprintf("isleader:%v", sev.leader), "passing")
	if err != nil {
		log.Errorf("update ttl of service error: ", err.Error())
	}
}

func (sev *Service) Register() error {
	sev.lock.Lock()
	if sev.status & Registered <= 0 {
		sev.status |= Registered
	}
	sev.lock.Unlock()
	// initial register service
	regis := &api.AgentServiceRegistration{
		ID:      sev.ServiceID,
		Name:    sev.ServiceName,
		Address: sev.ServiceHost,
		Port:    sev.ServicePort,
		Tags:    []string{fmt.Sprintf("isleader:%v", sev.leader)},
	}
	//log.Debugf("service register")
	err := sev.agent.ServiceRegister(regis)
	if err != nil {
		return fmt.Errorf("initial register service '%s' host to consul error: %s", sev.ServiceName, err.Error())
	}
	// initial register service check
	check := api.AgentServiceCheck{TTL: fmt.Sprintf("%ds", sev.Ttl), Status: "passing"}
	err = sev.agent.CheckRegister(&api.AgentCheckRegistration{
			ID: sev.ServiceID,
			Name: sev.ServiceName,
			ServiceID: sev.ServiceID,
			AgentServiceCheck: check,
		})
	if err != nil {
		return fmt.Errorf("initial register service check to consul error: %s", err.Error())
	}
	return nil
}

func (sev *Service) Close() {
	log.Infof("######################%v[%v] deregister", sev.ServiceName, sev.ServiceID)
	sev.Deregister()
	if sev.leader {
		sev.consulLock.Unlock()
		sev.consulLock.Delete()
		sev.leader = false
	}
}

func (sev *Service) GetServices(passingOnly bool) ([]*ServiceMember, error) {
	members, _, err := sev.health.Service(sev.ServiceName, "", passingOnly, nil)
	if err != nil {
		return nil, err
	}
	//return members, err
	data := make([]*ServiceMember, 0)
	for _, v := range members {
		//log.Debugf("GetServices： %+v", *v.Service)
		m := &ServiceMember{}
		if v.Checks.AggregatedStatus() == "passing" {
			m.Status = statusOnline
			m.IsLeader  = v.Service.Tags[0] == "isleader:true"
		} else {
			m.Status = statusOffline
			m.IsLeader  = false//v.Service.Tags[0] == "isleader:true"
		}
		m.ServiceID = v.Service.ID//Tags[1]
		m.ServiceIp = v.Service.Address
		m.Port      = v.Service.Port
		data        = append(data, m)
	}
	return data, nil
}

func (sev *Service) check() {
	time.Sleep(time.Second)
	success, err := sev.consulLock.Lock()
	if err == nil {
		sev.leader = success
		for _, f := range sev.onleader {
			go f(success)
		}
		sev.Register()
	}
	for {
		//log.Debugf("onleader num %v ", len(sev.onleader))
		success, err := sev.consulLock.Lock()
		if err == nil {
			if success != sev.leader {
				sev.leader = success
				for _, f := range sev.onleader {
					go f(success)
				}
				sev.Register()
			}
		}
		sev.session.Renew()
		sev.updateTtl()
		time.Sleep(time.Second * 3)
	}
}

func (sev *Service) GetLeader() (string, int, error) {
	members, _ := sev.GetServices(true)
	if members == nil {
		return "", 0, membersEmpty
	}
	for _, v := range members {
		//log.Debugf("getLeader: %+v", *v)
		if v.IsLeader {
			return v.ServiceIp, v.Port, nil
		}
	}
	return "", 0, leaderNotFound
}
