package consul

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"github.com/hashicorp/consul/api"
	"sync"
)

// 服务注册
const (
	Registered = 1 << iota
)
const (
	statusOnline    = "online"
	statusOffline   = "offline"
)
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
}

type ServiceOption func(s *Service)

// set ttl
func SetTtl(ttl int) ServiceOption {
	return func(s *Service){
		s.Ttl = ttl
	}
}

// set interval
func SetInterval(interval time.Duration) ServiceOption {
	return func(s *Service){
		s.Interval = interval
	}
}

// set service ip
//func ServiceIp(serviceIp string) ServiceOption {
//	return func(s *Service){
//		s.ServiceIp = serviceIp
//	}
//}

// new a service
// name: service name
// host: service host like 0.0.0.0 or 127.0.0.1
// port: service port, like 9998
// consulAddress: consul service address, like 127.0.0.1:8500
// opts: ServiceOption, like ServiceIp("127.0.0.1")
// return new service pointer
func NewService(
	c *api.Client,
	name string,
	host string,
	port int,
	isLeader bool,
	opts ...ServiceOption,
) *Service {
	sev := &Service{
		ServiceName : name,
		ServiceHost : host,
		ServicePort : port,
		Interval    : time.Second * 10,
		Ttl         : 15,
		status      : 0,
		leader      : isLeader,
		lock        : new(sync.Mutex),
	}
	for _, opt := range opts {
		opt(sev)
	}
	sev.client    = c
	sev.session   = NewSession(c.Session(), 0)
	sev.Kv        = c.KV()
	sev.ServiceID = fmt.Sprintf("%s-%s-%d", name, host, port)
	sev.agent     = sev.client.Agent()
	sev.health    = sev.client.Health()
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

func (sev *Service) UpdateTtl() {
	if sev.status & Registered <= 0 {
		return
	}
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
	log.Debugf("service register")
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
	sev.Deregister()
	if sev.leader {
		sev.leader = false
	}
}

func (sev *Service) GetServices() ([]*ServiceMember, error) {
	members, _, err := sev.health.Service(sev.ServiceName, "", false, nil)
	if err != nil {
		return nil, err
	}
	//return members, err
	data := make([]*ServiceMember, 0)
	for _, v := range members {
		log.Debugf("GetServices： %+v", *v.Service)
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

//func (sev *Service) getLeader() (string, int, error) {
//	members := sev.getMembers()
//	if members == nil {
//		return "", 0, membersEmpty
//	}
//	for _, v := range members {
//		log.Debugf("getLeader: %+v", *v)
//		if v.IsLeader {
//			return v.ServiceIp, v.Port, nil
//		}
//	}
//	return "", 0, leaderNotFound
//}

//func (sev *Service) ShowMembers() string {
//	data := sev.getMembers()
//	if data == nil {
//		return ""
//	}
//	hostname, err := os.Hostname()
//	if err != nil {
//		hostname = ""
//	}
//	res := fmt.Sprintf("current node: %s(%s:%d)\r\n", hostname, sev.ServiceIp, sev.ServicePort)
//	res += fmt.Sprintf("cluster size: %d node(s)\r\n", len(data))
//	res += fmt.Sprintf("======+=============================================+==========+===============\r\n")
//	res += fmt.Sprintf("%-6s| %-43s | %-8s | %s\r\n", "index", "node", "role", "status")
//	res += fmt.Sprintf("------+---------------------------------------------+----------+---------------\r\n")
//	for i, member := range data {
//		role := "follower"
//		if member.IsLeader {
//			role = "leader"
//		}
//		res += fmt.Sprintf("%-6d| %-43s | %-8s | %s\r\n", i, fmt.Sprintf("%s(%s:%d)", member.Hostname, member.ServiceIp, member.Port), role, member.Status)
//	}
//	res += fmt.Sprintf("------+---------------------------------------------+----------+---------------\r\n")
//	return res
//}
