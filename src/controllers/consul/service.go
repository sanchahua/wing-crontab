package consul

import (
	"app"
	log "github.com/sirupsen/logrus"
	"strings"
	"strconv"
	sconsul "github.com/jilieryuyi/wing-go/consul"
)

type OnLeaderFunc func(bool)
type ConsulController struct {
	sService sconsul.ILeader
	onLeader []OnLeaderFunc
}

type ConsulOption func(c *ConsulController)
func SetOnleader(f ...OnLeaderFunc) ConsulOption {
	return func(c *ConsulController) {
		c.onLeader = append(c.onLeader, f...)//c.service)
	}
}

func NewConsulController(ctx *app.Context) *ConsulController {
	st        := strings.Split(ctx.Config.BindAddress, ":")
	host      := st[0]
	port, err := strconv.ParseInt(st[1], 10, 64)
	if err != nil {
		log.Panicf("%v", err)
	}
	c := &ConsulController{
		onLeader: make([]OnLeaderFunc, 0),
	}
	c.sService = sconsul.NewLeader(
		ctx.Config.ConsulAddress, ctx.Config.LockKey,
		ctx.Config.ServiceName,
		host,
		int(port))
	return c
}

func (c *ConsulController) Start() {
	c.sService.Select(func(member *sconsul.ServiceMember) {
		for _, f := range c.onLeader {
			f(member.IsLeader)
		}
	})
}

func (c *ConsulController) Close() {
	c.sService.Free()
}

func (c *ConsulController) GetLeader() (string, int, error) {
	l, err := c.sService.Get()
	if l == nil {
		return "", 0, err
	}
	return l.ServiceIp, l.Port, err
}