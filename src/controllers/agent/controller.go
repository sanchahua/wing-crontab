package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
	index int64
}

type OnCommandFunc func(id int64, command string, runServer string)

func NewAgentController(
	ctx *app.Context,
	getleader agent.GetLeaferFunc,
	onevent agent.OnNodeEventFunc,
	oncommand OnCommandFunc,
) *AgentController {
	c := &AgentController{index:0}
	server := agent.NewAgentServer(ctx.Context(),
		ctx.Config.BindAddress,
		agent.SetEventCallback(onevent))
	client := agent.NewAgentClient(
		ctx.Context(),
		agent.SetGetLeafer(getleader),
		agent.SetOnCommand(func(id int64, command string) {
			oncommand(id, command, ctx.Config.BindAddress)
		}),
	)
	c.server = server
	c.client = client
	return c
}

func (c *AgentController) SendToLeader(data []byte) {
	c.client.Send(data)
}

func (c *AgentController) Dispatch(id int64, command string) {
	clients := c.server.Clients()
	l := int64(len(clients))
	if l <= 0 {
		log.Debugf("clients empty")
		return
	}
	if c.index >= l {
		atomic.StoreInt64(&c.index, 0)
	}
	client := clients[c.index]
	atomic.AddInt64(&c.index, 1)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(id))
	data = append(data, []byte(command)...)
	client.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, data))
}

func (c *AgentController) OnLeader(isLeader bool) {
	c.client.OnLeader(isLeader)
}

func (c *AgentController) Start() {
	c.server.Start()
}

func (c *AgentController) Close() {
	c.server.Close()
}
