package consul

import (
	"library/consul"
	"app"
	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
	"strings"
	"strconv"
)

type ConsulController struct {
	service *consul.Service
	onleader []consul.OnLeaderFunc
}

func NewConsulController(ctx *app.Context) *ConsulController {
	consulConfig        := api.DefaultConfig()
	consulConfig.Address = ctx.Config.ConsulAddress
	client, err         := api.NewClient(consulConfig)

	if err != nil {
		log.Panicf("%v", err)
	}

	st        := strings.Split(ctx.Config.BindAddress, ":")
	host      := st[0]
	port, err := strconv.ParseInt(st[1], 10, 64)

	if err != nil {
		log.Panicf("%v", err)
	}

	c       := &ConsulController{}
	session := consul.NewSession(client.Session(), 0)
	kv      := client.KV()
	lock    := consul.NewLock(session, kv, ctx.Config.LockKey)

	c.service = consul.NewService(
		client,
		session,
		kv,
		ctx.Config.ServiceName,
		host,
		int(port),
		consul.SetLockKey(lock),
		consul.SetOnLeader(c.OnLeader),
	)

	//select a leader
	watch := consul.NewWatch(client, ctx.Config.ServiceName,host, int(port), consul.SetServiceChange(func() {
		log.Infof("#################### reselect a new leader #######################")
		lock.Unlock()
		lock.Delete()
		c.service.SelectLeader()
	}))
	go watch.Start()
	return c
}

type ConsulControllerOption func(c* ConsulController)
func SetOnleader(f consul.OnLeaderFunc) ConsulControllerOption {
	return func(c *ConsulController) {
		c.onleader = append(c.onleader, f)
	}
}

// leader on select callback
func (c *ConsulController) OnLeader(isLeader bool) {
	log.Errorf("on leader( just for debug == %v )", isLeader)
	for _, f := range c.onleader {
		f(isLeader)
	}
}

func (c *ConsulController) Start() {
	c.service.SelectLeader()
}

func (c *ConsulController) Close() {
	c.service.Close()
}

func (c *ConsulController) GetLeader() (string, int, error) {
	return c.service.GetLeader()
}
