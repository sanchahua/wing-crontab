package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
	index int64
	dispatchChannel chan *runItem
}

type runItem struct {
	id int64
	command string
}

type OnCommandFunc func(id int64, command string, runServer string)

func NewAgentController(
	ctx *app.Context,
	getLeader agent.GetLeaderFunc,
	onEvent agent.OnNodeEventFunc,
	onCommand OnCommandFunc,
) *AgentController {
	c      := &AgentController{index:0}
	server := agent.NewAgentServer(ctx.Context(),
				ctx.Config.BindAddress,
				agent.SetEventCallback(onEvent),
			)
	client := agent.NewAgentClient(
				ctx.Context(),
				agent.SetGetLeader(getLeader),
				agent.SetOnCommand(func(id int64, command string) {
					onCommand(id, command, ctx.Config.BindAddress)
				}),
			)
	c.server = server
	c.client = client
	return c
}

// send data to leader
func (c *AgentController) SendToLeader(data []byte) {
	c.client.Send(data)
}

// roundbin dispatch to all clients
func (c *AgentController) Dispatch(id int64, command string) {
	start := time.Now()
	clients := c.server.Clients()
	l := int64(len(clients))
	if l <= 0 {
		log.Debugf("clients empty")
		return
	}
	if c.index >= l {
		atomic.StoreInt64(&c.index, 0)
	}
	log.Infof("clients %+v", l)
	log.Infof("dispatch %v=>%v to client[%v]", id, command, c.index)

	client := clients[c.index]
	atomic.AddInt64(&c.index, 1)
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(id))
	data = append(data, []byte(command)...)
	client.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, data))

	log.Debugf("dispatch use time %+v", time.Since(start))
}

// set on leader select callback
func (c *AgentController) OnLeader(isLeader bool) {
	c.client.OnLeader(isLeader)
}

// start agent
func (c *AgentController) Start() {
	c.server.Start()
}

// close agent
func (c *AgentController) Close() {
	c.server.Close()
}
