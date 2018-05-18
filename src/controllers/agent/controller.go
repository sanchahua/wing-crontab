package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"library/data"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
	index int64
	dispatch chan *runItem
	ctx *app.Context
	lock *sync.Mutex

	nums map[int64] int64
	numsLock *sync.Mutex
	queueNomal *data.EsQueue
	queueMutex *data.EsQueue
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
	//listLen := uint32(len(list))
	//if listLen < 1 {
	//	listLen = 1
	//}
	c      := &AgentController{index:0, dispatch:make(chan *runItem, 10000), ctx:ctx,
	lock:new(sync.Mutex), numsLock:new(sync.Mutex),
	queueNomal:data.NewQueue(maxQueueLen * listLen),
	queueMutex:data.NewQueue(maxQueueLen * listLen),
	nums:make(map[int64] int64),
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
	log.Debugf("######### on pull")
	itemI, ok, num := c.queueNomal.Get()
	if !ok || itemI == nil {
		log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
		return
	}
	item := itemI.(*runItem)
	log.Debugf("######## (onpull response) send %+v", *item)
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
	start := time.Now()
	node.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, sendData))
	log.Debugf("AsyncSend use time %+v", time.Since(start))
}

func (c *AgentController) Pull() {
	c.client.Write(agent.Pack(agent.CMD_PULL_COMMAND, []byte("")))
}

func (c *AgentController) Dispatch(id int64, command string, isMutex bool) {
	//start := time.Now().Unix()
	//for {
	//	if len(c.dispatch) < cap(c.dispatch) {
	//		break
	//	} else {
	//		log.Errorf("dispatch cache full")
	//	}
	//	// only wait 6 seconds, if timeout, just return
	//	if time.Now().Unix() - start >= 6 {
	//		log.Errorf("Dispatch wait timeout: %v, %v", id, command)
	//		return
	//	}
	//}

	// todo
	// 这里的派发
	// 优先派发c.nums[id]最少的，因为这个产生的周期比较长
	// 优先派发需要互斥运行的
	// 需要互斥运行的，每次会在收到上次的执行完成之后，才可以分发
	// 分发需要做可靠性处理

	//c.numsLock.Lock()
	//num, _ := c.nums[id]
	//c.numsLock.Unlock()
	//
	//if num >= maxQueueLen {
	//	log.Warnf("%v list is max then %v", id, maxQueueLen)
	//	return
	//}


	var ok = false
	if isMutex {
		if c.queueMutex.Quantity() < maxQueueLen {
			item := &runItem{id: id, command: command, isMutex: isMutex}
			ok, _ = c.queueMutex.Put(item)
			if !ok {
				log.Errorf("put queue failure")
			}
		} else {
			log.Warnf("%v list is max then %v", id, maxQueueLen)
		}
	} else {
		if c.queueMutex.Quantity() < maxQueueLen {
			item := &runItem{id: id, command: command, isMutex: isMutex}
			ok, _ = c.queueNomal.Put(item)
			if !ok {
				log.Errorf("put queue failure")
			}
		} else {
			log.Warnf("%v list is max then %v", id, maxQueueLen)
		}
	}



	//c.numsLock.Lock()
	//a, ok := c.nums[id]
	//if !ok {
	//	a = 0
	//}
	//a++
	//c.nums[id] = a
	//c.numsLock.Unlock()

	//start := time.Now()
	//c.dispatchProcess(id, command)
	//log.Debugf("dispatch use time %v", time.Since(start))
}

func (c *AgentController) dispatchProcess() {
	//need to add wait for dispatch complete if exit
	// roundbin dispatch to all clients

	//dataDispatchServerLen := make([]byte, 8)
	//binary.LittleEndian.PutUint64(dataDispatchServerLen, uint64(len(c.ctx.Config.BindAddress)))

	//var dis = func(item *runItem) {
	//	start := time.Now()
	//	start1 := time.Now()
	//	clients := c.server.Clients()
	//	log.Debugf("c.server.Clients use time: %+v", time.Since(start1))
	//
	//	start2 := time.Now()
	//	l := int64(len(clients))
	//	if l <= 0 {
	//		log.Debugf("clients empty")
	//		return
	//	}
	//	if c.index >= l {
	//		atomic.StoreInt64(&c.index, 0)
	//	}
	//	log.Infof("clients %+v", l)
	//	log.Debugf("c.server.Clients use time => 2 : %+v", time.Since(start2))
	//
	//	for key, client := range clients {
	//		if key != int(c.index) {
	//			continue
	//		}
	//		start3 := time.Now()
	//		log.Infof("dispatch %v=>%v to client[%v]", item.id, item.command, c.index)
	//		//client := clients[c.index]
	//		atomic.AddInt64(&c.index, 1)
	//		data := make([]byte, 8)
	//		binary.LittleEndian.PutUint64(data, uint64(item.id))
	//
	//		dataCommendLen := make([]byte, 8)
	//		binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))
	//
	//		data = append(data, dataCommendLen...)
	//		data = append(data, []byte(item.command)...)
	//
	//		//data = append(data, dataDispatchServerLen...)
	//		data = append(data, []byte(c.ctx.Config.BindAddress)...)
	//		log.Debugf("c.server.Clients use time => 3 : %+v", time.Since(start3))
	//
	//		start5 := time.Now()
	//		client.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, data))
	//		log.Debugf("c.server.Clients use time => 5 : %+v", time.Since(start5))
	//
	//	}
	//	log.Debugf("dispatch use time %+v", time.Since(start))
	//}
	//for {
	//	select {
	//		case item, ok := <- c.dispatch:
	//			if !ok {
	//				return
	//			}
	//			//c.lock.Lock()
	//			data := make([]byte, 8)
	//			binary.LittleEndian.PutUint64(data, uint64(item.id))
	//
	//			dataCommendLen := make([]byte, 8)
	//			binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))
	//
	//			currentTime := make([]byte, 8)
	//			binary.LittleEndian.PutUint64(currentTime, uint64(time.Now().Unix()))
	//			data = append(data, currentTime...)
	//
	//			data = append(data, dataCommendLen...)
	//			data = append(data, []byte(item.command)...)
	//
	//			data = append(data, []byte(c.ctx.Config.BindAddress)...)
	//			start := time.Now()
	//			c.server.RandSend(data)
	//			log.Debugf("dispatch use time %+v", time.Since(start))
	//
	//			//c.lock.Unlock()
	//	}
	//}
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
