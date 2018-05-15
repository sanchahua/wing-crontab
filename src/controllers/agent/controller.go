package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
	"runtime"
	"sync"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
	index int64
	dispatch chan *runItem
	lock *sync.Mutex
	ctx *app.Context
}

type runItem struct {
	id int64
	command string
}

type OnCommandFunc func(id int64, command string, dispatchServer string, runServer string)

func NewAgentController(
	ctx *app.Context,
	getLeader agent.GetLeaderFunc,
	onEvent agent.OnNodeEventFunc,
	onCommand OnCommandFunc,
) *AgentController {
	c      := &AgentController{
				index:0,
				dispatch:make(chan *runItem, 10000),
				lock:new(sync.Mutex),
				ctx:ctx,
			}
	server := agent.NewAgentServer(ctx.Context(),
				ctx.Config.BindAddress,
				agent.SetEventCallback(onEvent),
			)
	client := agent.NewAgentClient(
				ctx.Context(),
				agent.SetGetLeader(getLeader),
				agent.SetOnCommand(func(content []byte) {

					id := binary.LittleEndian.Uint64(content[:8])
					//log.Debugf("id == (%v) === (%v) ", id, content[:8])
					//log.Debugf("content == (%v) === (%v) ", string(content[8:]), content[:8])
					commandLen := binary.LittleEndian.Uint64(content[8:16])
					command    := content[16:16+commandLen]

					//dispatchServerLen := content[16+commandLen:24+commandLen]
					dispatchServer    := content[16+commandLen:]

					onCommand(int64(id), string(command), string(dispatchServer), ctx.Config.BindAddress)
				}),
			)
	c.server = server
	c.client = client
	cpu := runtime.NumCPU()
	for i:=0;i<cpu;i++ {
		go c.dispatchProcess()
	}
	return c
}

// send data to leader
func (c *AgentController) SendToLeader(data []byte) {
	c.client.Send(data)
}
func (c *AgentController) Dispatch(id int64, command string) {
	if len(c.dispatch) < cap(c.dispatch) {
		c.dispatch <- &runItem{id: id, command: command}
	} else {
		log.Errorf("dispatch cache full")
	}
}
func (c *AgentController) dispatchProcess() {
	//need to add wait for dispatch complete if exit
	// roundbin dispatch to all clients

	//dataDispatchServerLen := make([]byte, 8)
	//binary.LittleEndian.PutUint64(dataDispatchServerLen, uint64(len(c.ctx.Config.BindAddress)))


	for {
		select {
			case item, ok := <- c.dispatch:
				if !ok {
					return
				}
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
			//c.lock.Lock()
			//client := clients[c.index]
			//c.lock.Unlock()
			for key, client := range clients {
				if key != int(c.index) {
					continue
				}
				log.Infof("dispatch %v=>%v to client[%v]", item.id, item.command, c.index)
				//client := clients[c.index]
				atomic.AddInt64(&c.index, 1)
				data := make([]byte, 8)
				binary.LittleEndian.PutUint64(data, uint64(item.id))

				dataCommendLen := make([]byte, 8)
				binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))

				data = append(data, dataCommendLen...)
				data = append(data, []byte(item.command)...)

				//data = append(data, dataDispatchServerLen...)
				data = append(data, []byte(c.ctx.Config.BindAddress)...)

				client.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, data))
			}
			log.Debugf("dispatch use time %+v", time.Since(start))
		}
	}
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
