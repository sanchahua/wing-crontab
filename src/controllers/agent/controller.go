package agent

import (
	"library/agent"
	"app"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
}

func NewAgentController(
	ctx *app.Context,
	getleader agent.GetLeaferFunc,
	//onevent agent.OnEventFunc,
) *AgentController {
	c := &AgentController{}
	server := agent.NewAgentServer(ctx.Context(), ctx.Config.BindAddress)
	client := agent.NewAgentClient(
		ctx.Context(),
		agent.SetGetLeafer(getleader),
		//agent.SetOnEvent(onevent),
	)
	c.server = server
	c.client = client
	return c
}

func (c *AgentController) SendToLeader(data []byte) {
	c.client.Send(data)
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
