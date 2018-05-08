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
}

func NewConsulController(ctx *app.Context) *ConsulController {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = ctx.Config.ConsulAddress
	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Panicf("%v", err)
	}

	st := strings.Split(ctx.Config.BindAddress, ":")
	host := st[0]
	port, err := strconv.ParseInt(st[1], 10, 64)
	if err != nil {
		log.Panicf("%v", err)
	}
	c := &ConsulController{}
	c.service = consul.NewService(
		client,
		ctx.Config.ServiceName,
		host,
		int(port),
		consul.SetLockKey(ctx.Config.LockKey),
		consul.SetOnLeader(c.OnLeader),
	)
	//select a leader
	c.service.SelectLeader()
	return c
}

// leader on select callback
func (c *ConsulController) OnLeader(isLeader bool) {

}
