package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	"sync"
	"time"
	"library/data"
	"sync/atomic"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
	index int64
	dispatch chan *runItem
	ctx *app.Context
	lock *sync.Mutex

	//nums map[int64] int64
	//numsLock *sync.Mutex

	queueNomalLock *sync.Mutex
	queueNomal map[int64]*data.EsQueue

	queueMutexLock *sync.Mutex
	queueMutex map[int64]*data.EsQueue
}

type runItem struct {
	id int64
	command string
	isMutex bool
}

type OnCommandFunc func(id int64, command string, dispatchTime int64, dispatchServer string, runServer string)
const maxQueueLen = 64
func NewAgentController(
	ctx *app.Context,
	listLen uint32,//[]*cron.CronEntity,
	getLeader agent.GetLeaderFunc,
	onEvent agent.OnNodeEventFunc,
	onCommand OnCommandFunc,
) *AgentController {
	c      := &AgentController{
				index:0, dispatch:make(chan *runItem, 10000), ctx:ctx,
				lock:new(sync.Mutex), //numsLock:new(sync.Mutex),
				queueNomal:make(map[int64]*data.EsQueue),
				queueMutex:make(map[int64]*data.EsQueue),
				queueNomalLock:new(sync.Mutex),
				queueMutexLock:new(sync.Mutex),
				//nums:make(map[int64] int64),
			}

	//for _, v := range list {
	//	c.nums[v.Id] = 0
	//}

	server := agent.NewAgentServer(ctx.Context(), ctx.Config.BindAddress, agent.SetEventCallback(onEvent), agent.SetServerOnPullCommand(c.OnPullCommand))
	client := agent.NewAgentClient(ctx.Context(), agent.SetGetLeader(getLeader),
				agent.SetOnCommand(func(content []byte) {
					id             := binary.LittleEndian.Uint64(content[:8])
					dispatchTime   := binary.LittleEndian.Uint64(content[8:16])
					commandLen     := binary.LittleEndian.Uint64(content[16:24])
					command        := content[24:24 + commandLen]
					dispatchServer := content[24 + commandLen:]
					onCommand(int64(id), string(command), int64(dispatchTime), string(dispatchServer), ctx.Config.BindAddress)
				}), )
	c.server = server
	c.client = client
	//cpu := runtime.NumCPU()
	//for i:= 0; i < cpu; i++ {
	//	go c.dispatchProcess()
	//}
	return c
}

// send data to leader
func (c *AgentController) SendToLeader(data []byte) {
	c.client.Send(agent.CMD_CRONTAB_CHANGE, data)
}

func (c *AgentController) OnPullCommand(node *agent.TcpClientNode) {
	//log.Debugf("######### on pull")

	// todo
	// 这里的派发
	// 优先派发queue num min 最少的，因为这个产生的周期比较长
	// 优先派发需要互斥运行的
	// 需要互斥运行的，每次会在收到上次的执行完成之后，才可以分发
	// 分发需要做可靠性处理
	//start := time.Now()
	//var queueNormal *data.EsQueue
	//num := uint32(0)
	//for _, q := range c.queueNomal {
	//	num = q.Quantity()
	//	if num > 0 {
	//		queueNormal = q
	//		break
	//	}
	//}
	//
	//if queueNormal == nil || num <= 0 {
	//	return
	//}
	//
	//for _, q := range c.queueNomal {
	//	qn := q.Quantity()
	//	if qn < num {
	//		queueNormal = q
	//		num = qn
	//
	//	}
	//}
	index := int64(-1)
	if c.index >= int64(len(c.queueNomal) - 1) {
		atomic.StoreInt64(&c.index, 0)
	}
	for _ , queueNormal := range c.queueNomal {
		index++
		if index != c.index {
			continue
		}
		atomic.AddInt64(&c.index, 1)
		itemI, ok, _ := queueNormal.Get()
		if !ok || itemI == nil {
			//log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
			return
		}
		item := itemI.(*runItem)
		//log.Debugf("######## (onpull response) send %+v", *item)
		sendData := make([]byte, 8)
		binary.LittleEndian.PutUint64(sendData, uint64(item.id))

		dataCommendLen := make([]byte, 8)
		binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))

		currentTime := make([]byte, 8)
		binary.LittleEndian.PutUint64(currentTime, uint64(time.Now().Unix()))
		sendData = append(sendData, currentTime...)

		sendData = append(sendData, dataCommendLen...)
		sendData = append(sendData, []byte(item.command)...)

		sendData = append(sendData, []byte(c.ctx.Config.BindAddress)...)
		//start2 := time.Now()
		node.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, sendData))
		//log.Debugf("AsyncSend use time %+v", time.Since(start2))
		//log.Debugf("OnPullCommand use time %+v", time.Since(start))
		break
	}
}

func (c *AgentController) Pull() {
	c.client.Write(agent.Pack(agent.CMD_PULL_COMMAND, []byte("")))
}

func (c *AgentController) Dispatch(id int64, command string, isMutex bool) {
	//logrus.Debug("Dispatch %v, %v, %v", id, command, isMutex)
	if isMutex {
		c.queueMutexLock.Lock()
		queueMutex, ok := c.queueMutex[id]
		if !ok {
			queueMutex = data.NewQueue(maxQueueLen)
			c.queueMutex[id] = queueMutex
		}
		c.queueMutexLock.Unlock()
		item := &runItem{id: id, command: command, isMutex: isMutex}
		queueMutex.Put(item)
		return
	}

	c.queueNomalLock.Lock()
	queueNormal, ok := c.queueNomal[id]
	if !ok {
		queueNormal = data.NewQueue(maxQueueLen)
		c.queueNomal[id] = queueNormal
	}
	c.queueNomalLock.Unlock()
	item := &runItem{id: id, command: command, isMutex: isMutex}
	queueNormal.Put(item)
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
