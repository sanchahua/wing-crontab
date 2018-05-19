package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	"sync"
	"time"
	"library/data"
	"sync/atomic"
	wstring "library/string"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"runtime"
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
	keep map[string] []byte
	keepLock *sync.Mutex

	sendQueue map[string]*SendData
	sendQueueLock *sync.Mutex
}

type SendData struct {
	Unique string `json:"unique"`
	Data []byte `json:"data"`
	Status int `json:"status"`
	Time int64 `json:"time"`
	SendTimes int `json:"send_times"`
	Cmd int `json:"cmd"`
	send sendFunc `json:"-"`
}

type sendFunc func(data []byte)

func newSendData(cmd int, data []byte, send sendFunc) *SendData {
	return &SendData{
		Unique:    wstring.RandString(128),
		Data:      data,
		Status:    0,
		Time:      0,
		SendTimes: 0,
		Cmd:       cmd,
		send:      send,
	}

}

func (d *SendData) encode() []byte {
	b, e := json.Marshal(d)
	if e != nil {
		return nil
	}
	return b
}

type runItem struct {
	id int64
	command string
	isMutex bool
}

type OnCommandFunc func(id int64, command string, dispatchTime int64, dispatchServer string, runServer string)
const maxQueueLen = 64
const dispatchChanLen = 10000
func NewAgentController(
	ctx *app.Context,
	getLeader agent.GetLeaderFunc,
	onEvent agent.OnNodeEventFunc,
	onCommand OnCommandFunc,
) *AgentController {
	c      := &AgentController{
				index:0,
				dispatch:make(chan *runItem, dispatchChanLen),
				ctx:ctx,
				lock:new(sync.Mutex),
				queueNomal:make(map[int64]*data.EsQueue),
				queueMutex:make(map[int64]*data.EsQueue),
				queueNomalLock:new(sync.Mutex),
				queueMutexLock:new(sync.Mutex),
				keep: make(map[string] []byte),
				keepLock:new(sync.Mutex),
				sendQueue: make(map[string]*SendData),
				sendQueueLock: new(sync.Mutex),
				//nums:make(map[int64] int64),
			}

	//for _, v := range list {
	//	c.nums[v.Id] = 0
	//}

	server := agent.NewAgentServer(
			ctx.Context(),
			ctx.Config.BindAddress,
			agent.SetOnServerEvents(func(node *agent.TcpClientNode, event int, content []byte) {
				log.Debugf("server receive:, %v, %v", event, content )
				switch event {
				case agent.CMD_PULL_COMMAND:
					c.OnPullCommand(node)
				case agent.CMD_CRONTAB_CHANGE:
					var sdata SendData
					err := json.Unmarshal(content, &sdata)
					if err != nil {
						log.Errorf("%+v", err)
					} else {
						event := binary.LittleEndian.Uint32(sdata.Data[:4])
						go onEvent(int(event), sdata.Data[4:])
						//log.Infof("receive event[%v] %+v", event, string(data.Data[4:]))
						node.AsyncSend(agent.Pack(agent.CMD_CRONTAB_CHANGE, []byte(sdata.Unique)))
					}
				case agent.CMD_RUN_COMMAND:
					log.Debugf("command is run (will delete from send queue): %v", string(content))
					c.sendQueueLock.Lock()
					delete(c.sendQueue, string(content))
					log.Debugf("send queue len: %v", len(c.sendQueue))
					c.sendQueueLock.Unlock()
				}
			}),
			//agent.SetEventCallback(onEvent),
		)
	client := agent.NewAgentClient(ctx.Context(),
				agent.SetGetLeader(getLeader),
				agent.SetOnClientEvent(func(tcp *agent.AgentClient, cmd int , content []byte) {
					log.Debugf("#############client receive: cmd=%d, content=%v", cmd, string(content))
					switch cmd {
					case agent.CMD_RUN_COMMAND:
						func() {
							var sendData SendData
							err := json.Unmarshal(content, &sendData)
							if err != nil {
								log.Errorf("%#############v", err)
								return
							}

							log.Debugf("#############receive command: %+v", sendData)
							if len(sendData.Data) < 24 {
								return
							}
							id := binary.LittleEndian.Uint64(sendData.Data[:8])
							dispatchTime := binary.LittleEndian.Uint64(sendData.Data[8:16])
							commandLen := binary.LittleEndian.Uint64(sendData.Data[16:24])
							if len(sendData.Data) < int(24+commandLen) {
								return
							}
							command := sendData.Data[24:24+commandLen]

							log.Debugf("##############send: %v", sendData.Unique)
							tcp.Write(agent.Pack(agent.CMD_RUN_COMMAND, []byte(sendData.Unique)))

							dispatchServer := sendData.Data[24+commandLen:]
							onCommand(int64(id), string(command), int64(dispatchTime), string(dispatchServer), ctx.Config.BindAddress)
						}()
					case agent.CMD_CRONTAB_CHANGE:
						log.Infof("cron send to leader server ok (will delete from send queue): %+v", string(content))
						c.sendQueueLock.Lock()
						delete(c.sendQueue, string(content))
						log.Debugf("send queue len: %v", len(c.sendQueue))
						c.sendQueueLock.Unlock()
					}
				}), )
	c.server = server
	c.client = client
	cpu := runtime.NumCPU()
	for i:= 0; i < cpu; i++ {
		go c.sendService()
	}
	return c
}

// send data to leader
func (c *AgentController) SendToLeader(data []byte) {
	//
	//c.client.Send(agent.CMD_CRONTAB_CHANGE, data)

	d := newSendData(agent.CMD_CRONTAB_CHANGE, data, c.client.Write)
	c.sendQueueLock.Lock()
	c.sendQueue[d.Unique] = d
	c.sendQueueLock.Unlock()
}

func (c *AgentController) sendService() {
	for {
		//select {
		//case <-tcp.ctx.Done():
		//	log.Debugf("keepalive exit 1")
		//	return
		//default:
		//}


		c.sendQueueLock.Lock()
		if len(c.sendQueue) <= 0 {
			c.sendQueueLock.Unlock()
			time.Sleep(time.Microsecond*10)
			continue
		}

		log.Debugf("send queue len: %v", len(c.sendQueue))

		for _, d := range c.sendQueue {
			// status > 0 is sending
			// 发送中的数据，3秒之内不会在发送，超过3秒会进行2次重试
			// todo ？？这里的3秒设置的是否合理，这里最好的方式应该有一个实时发送时间反馈
			// 比如完成一次发送需要100ms，超时时间设置为 100ms + 3s 这样应该更合理
			// 即t+3模式
			if d.Status > 0 && (time.Now().Unix() - d.Time) <= 3 {
				continue
			}
			//log.Infof("try to send %+v", *d)
			d.Status = 1
			d.SendTimes++

			if d.SendTimes > 1 {
				log.Warnf("send times %v, *d", d.SendTimes, *d)
			}

			// 每次延迟3秒重试，最多20次，即1分钟之内才会重试
			if d.SendTimes >= 20 {
				delete(c.sendQueue, d.Unique)
				log.Warnf("send timeout(36s), delete %+v", *d)
				continue
			}
			d.Time    = time.Now().Unix()
			sd       := d.encode()
			sendData := agent.Pack(d.Cmd, sd)

			d.send(sendData)
		}
		c.sendQueueLock.Unlock()
		//time.Sleep(time.Second * 10)
	}
}

// 客户端主动发送pull请求到server端
// pull请求到达，说明客户端有能力执行当前的定时任务
// 这个时候可以继续分配定时任务给客户端
// 整个系统才去主动拉取的模式，只有客户端空闲达到一定程度，或者说足以负载当前的任务才会发起pull请求
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
	c.queueNomalLock.Lock()

	for _ , queueNormal := range c.queueNomal {
		index++
		if index != c.index {
			continue
		}
		atomic.AddInt64(&c.index, 1)
		itemI, ok, _ := queueNormal.Get()
		if !ok || itemI == nil {
			c.queueNomalLock.Unlock()
			//log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
			return
		}
		item := itemI.(*runItem)

		////////////////////////

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
		//node.AsyncSend(agent.Pack(agent.CMD_RUN_COMMAND, sendData))

		d := newSendData(agent.CMD_RUN_COMMAND, sendData, node.AsyncSend)
		c.sendQueueLock.Lock()
		c.sendQueue[d.Unique] = d
		c.sendQueueLock.Unlock()

		//log.Debugf("######## (onpull response) send %+v", *d)


		//log.Debugf("AsyncSend use time %+v", time.Since(start2))
		//log.Debugf("OnPullCommand use time %+v", time.Since(start))
		break
	}
	c.queueNomalLock.Unlock()
	//
}

func (c *AgentController) Pull() {
	//log.Debugf("##############################pull command(%v)", agent.CMD_PULL_COMMAND)
	c.client.Write(agent.Pack(agent.CMD_PULL_COMMAND, []byte("")))
}

func (c *AgentController) Dispatch(id int64, command string, isMutex bool) {
	//logrus.Debugf("Dispatch %v, %v, %v", id, command, isMutex)
	if isMutex {
		c.queueMutexLock.Lock()
		queueMutex, ok := c.queueMutex[id]
		if !ok {
			queueMutex = data.NewQueue(maxQueueLen)
			c.queueMutex[id] = queueMutex
		}
		c.queueMutexLock.Unlock()
		item := &runItem{
				id: id,
				command: command,
				isMutex: isMutex,
			}
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
	item := &runItem{id: id, command: command, isMutex: isMutex,}
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
