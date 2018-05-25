package consul

import (
	"library/consul"
	"app"
	log "github.com/sirupsen/logrus"
	"strings"
	"strconv"
)

type ConsulController struct {
	service *consul.Service
}

type ConsulOption func(c *ConsulController)
func SetOnleader(f ...consul.OnLeaderFunc) ConsulOption {
	return func(c *ConsulController) {
		consul.SetOnLeader(f...)(c.service)
	}
}

func NewConsulController(ctx *app.Context) *ConsulController {
	st        := strings.Split(ctx.Config.BindAddress, ":")
	host      := st[0]
	port, err := strconv.ParseInt(st[1], 10, 64)

	if err != nil {
		log.Panicf("%v", err)
	}

	c := &ConsulController{}
	c.service = consul.NewService(
		ctx.Config.ConsulAddress,
		ctx.Config.LockKey,
		ctx.Config.ServiceName,
		host,
		int(port),
	)
	return c
}

func (c *ConsulController) Start() {
}

func (c *ConsulController) Close() {
	c.service.Close()
}

func (c *ConsulController) GetLeader() (string, int, error) {
	return c.service.GetLeader()
}