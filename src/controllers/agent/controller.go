package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	"sync"
	"time"
	"library/data"
	"sync/atomic"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"runtime"
	"errors"
)

type AgentController struct {
	client *agent.AgentClient
	server *agent.TcpService
	indexNormal int64
	indexMutex int64
	dispatch chan *runItem
	ctx *app.Context
	lock *sync.Mutex
	queueNomalLock *sync.Mutex
	queueNomal map[int64]*data.EsQueue
	queueMutexLock *sync.Mutex
	queueMutex map[int64]*Mutex
	keep map[string] []byte
	keepLock *sync.Mutex
	sendQueue map[string]*SendData
	sendQueueLock *sync.Mutex
	onCronChange OnCronChangeEventFunc
	onCommand OnCommandFunc
}

const (
	maxQueueLen = 64
	dispatchChanLen = 10000
)

type sendFunc func(data []byte)
type OnCommandFunc func(id int64, command string, dispatchTime int64, dispatchServer string, runServer string, isMutex byte, after func())
type OnCronChangeEventFunc func(event int, data []byte)

func NewAgentController(
	ctx *app.Context,
	//这个参数提供了查询leader的服务ip和端口
	//最终用于agent client连接leader时使用
	//来源于consulControl.GetLeader
	getLeader agent.GetLeaderFunc,
	//http api增删改查定时任务会触发这个事件，最终这个事件影响到leader的定时任务
	//最终落入这个api crontabController.Add
	onCronChange OnCronChangeEventFunc,
	//这个事件由leader分发定时任务到节点，节点收到定时任务时触发
	//最终接收的api是crontabController.ReceiveCommand
	onCommand OnCommandFunc,
) *AgentController {
	c      := &AgentController{
				indexNormal:    0,
				indexMutex:     0,
				dispatch:       make(chan *runItem, dispatchChanLen),
				ctx:            ctx,
				lock:           new(sync.Mutex),
				queueNomal:     make(map[int64]*data.EsQueue),
				queueMutex:     make(map[int64]*Mutex),
				queueNomalLock: new(sync.Mutex),
				queueMutexLock: new(sync.Mutex),
				keep:           make(map[string] []byte),
				keepLock:       new(sync.Mutex),
				sendQueue:      make(map[string]*SendData),
				sendQueueLock:  new(sync.Mutex),
				onCronChange:   onCronChange,
				onCommand:      onCommand,
			}
	c.server = agent.NewAgentServer(ctx.Context(), ctx.Config.BindAddress, agent.SetOnServerEvents(c.onServerEvent), )
	c.client = agent.NewAgentClient(ctx.Context(), agent.SetGetLeader(getLeader), agent.SetOnClientEvent(c.onClientEvent), )
	cpu := runtime.NumCPU()
	for i:= 0; i < cpu; i++ {
		go c.sendService()
	}
	return c
}

func (c *AgentController) onClientEvent(tcp *agent.AgentClient, cmd int , content []byte) {
	//log.Debugf("#############client receive: cmd=%d, content=%v", cmd, string(content))
	switch cmd {
	case agent.CMD_RUN_COMMAND:
		func() {
			var sendData SendData
			err := json.Unmarshal(content, &sendData)
			if err != nil {
				log.Errorf("%#############v", err)
				return
			}
			//log.Debugf("****************************command is begin to run 2 => %v, command start run time is %v", sendData.Unique, time.Now().UnixNano())

			//log.Debugf("#############receive command: %+v", sendData)
			//if len(sendData.Data) < 25 {
			//	return
			//}

			id, dispatchTime, isMutex, command, dispatchServer, err := c.unpack(sendData.Data)
			if err != nil {
				return
			}

			//id := binary.LittleEndian.Uint64(sendData.Data[:8])
			//dispatchTime := binary.LittleEndian.Uint64(sendData.Data[8:16])
			//// binary.LittleEndian.Uint64(sendData.Data[16:17]) == 1
			//
			//commandLen := binary.LittleEndian.Uint64(sendData.Data[16:24])
			//if len(sendData.Data) < int(24+commandLen) {
			//	return
			//}
			//command := sendData.Data[24:24+commandLen]
			//
			////log.Debugf("##############send: %v", sendData.Unique)
			sdata := make([]byte, 0)
			sid := make([]byte, 8)
			binary.LittleEndian.PutUint64(sid, uint64(id))
			sdata = append(sdata, sid...)
			sdata = append(sdata, isMutex)
			sdata = append(sdata, []byte(sendData.Unique)...)

			//
			//dispatchServer := sendData.Data[24+commandLen:]
			c.onCommand(id, command, dispatchTime, dispatchServer, c.ctx.Config.BindAddress, isMutex, func() {
				//log.Debugf("****************************command is run end 3 => %v, command run end time is %v", sendData.Unique, time.Now().UnixNano())
				tcp.Write(agent.Pack(agent.CMD_RUN_COMMAND, sdata))
			})
		}()
	case agent.CMD_CRONTAB_CHANGE:
		log.Infof("cron send to leader server ok (will delete from send queue): %+v", string(content))
		c.sendQueueLock.Lock()
		delete(c.sendQueue, string(content))
		//log.Debugf("send queue len: %v", len(c.sendQueue))
		c.sendQueueLock.Unlock()
	}
}

func (c *AgentController) onServerEvent(node *agent.TcpClientNode, event int, content []byte) {
	//log.Debugf("server receive:, %v, %v", event, content )
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
			go c.onCronChange(int(event), sdata.Data[4:])
			//log.Infof("receive event[%v] %+v", event, string(data.Data[4:]))
			node.AsyncSend(agent.Pack(agent.CMD_CRONTAB_CHANGE, []byte(sdata.Unique)))
		}
	case agent.CMD_RUN_COMMAND:

		//sdata := make([]byte, 0)
		//sid := make([]byte, 8)
		//binary.LittleEndian.PutUint64(sid, uint64(id))
		//sdata = append(sdata, sid...)
		//sdata = append(sdata, isMutex)
		//sdata = append(sdata, []byte(sendData.Unique)...)
		id := int64(binary.LittleEndian.Uint64(content[:8]))
		isMutex := content[8]
		unique := string(content[9:])

		if isMutex == 1 {
			//log.Debugf("set is running false")
			c.queueMutexLock.Lock()
			m ,ok := c.queueMutex[id]
			if ok {
				m.isRuning = false
			} else {
				log.Errorf("%v does not exists")
			}
			c.queueMutexLock.Unlock()
		}

		//log.Debugf("command is run (will delete from send queue): %v", string(content))
		c.sendQueueLock.Lock()
		delete(c.sendQueue, unique)
		//log.Debugf("send queue len: %v", len(c.sendQueue))
		c.sendQueueLock.Unlock()

		//log.Debugf("****************************command run back 4 => %v, command back time is %v", unique, time.Now().UnixNano())

	}
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
			time.Sleep(time.Millisecond * 100)
			continue
		}

		//log.Debugf("send queue len: %v", len(c.sendQueue))

		times3 := 0
		for _, d := range c.sendQueue {
			// status > 0 is sending
			// 发送中的数据，3秒之内不会在发送，超过3秒会进行2次重试
			// todo ？？这里的3秒设置的是否合理，这里最好的方式应该有一个实时发送时间反馈
			// 比如完成一次发送需要100ms，超时时间设置为 100ms + 3s 这样应该更合理
			// 即t+3模式
			// 默认60秒超时重试
			if d.Status > 0 && (time.Now().Unix() - d.Time) <= 60 {
				times3++
				continue
			}
			//log.Infof("try to send %+v", *d)
			d.Status = 1
			d.SendTimes++

			if d.SendTimes > 1 {
				log.Warnf("send times %v, *d", d.SendTimes, *d)
			}

			// 每次延迟3秒重试，最多20次，即1分钟之内才会重试
			if d.SendTimes >= 60 {
				delete(c.sendQueue, d.Unique)
				log.Warnf("send timeout(36s), delete %+v", *d)
				continue
			}
			d.Time    = time.Now().Unix()
			sd       := d.encode()
			sendData := agent.Pack(d.Cmd, sd)

			//log.Debugf("%+v", *d)
			//log.Debugf("****************************command is begin to run 1 => %v, send time is %v", d.Unique, time.Now().UnixNano())
			d.send(sendData)
		}
		c.sendQueueLock.Unlock()
		// 如果都是发送中，这里尝试等待10毫秒，让出cpu
		if times3 >= len(c.sendQueue) {
			time.Sleep(time.Millisecond * 10)
		}
	}
}

var commandLenError = errors.New("command len error")
func (c *AgentController) unpack(data []byte) (id int64, dispatchTime int64, isMutex byte, command string, dispatchServer string, err error) {
	if len(data) < 25 {
		err = commandLenError
		return
	}
	err = nil
	id = int64(binary.LittleEndian.Uint64(data[:8]))
	dispatchTime = int64(binary.LittleEndian.Uint64(data[8:16]))
	isMutex = data[16]

	commandLen := binary.LittleEndian.Uint64(data[17:25])
	if len(data) < int(25 + commandLen) {
		err = commandLenError//errors.New("command len error")
		return
	}
	command = string(data[25:25+commandLen])

	//log.Debugf("##############send: %v", sendData.Unique)
	//tcp.Write(agent.Pack(agent.CMD_RUN_COMMAND, []byte(sendData.Unique)))

	dispatchServer = string(data[25+commandLen:])
	return
}

func (c *AgentController) pack(item *runItem) []byte {
	json.Marshal(item)
	sendData := make([]byte, 8)
	binary.LittleEndian.PutUint64(sendData, uint64(item.id))

	dataCommendLen := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))

	currentTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(currentTime, uint64(time.Now().Unix()))
	sendData = append(sendData, currentTime...)

	if item.isMutex {
		sendData = append(sendData, byte(1))
	} else {
		sendData = append(sendData, byte(0))
	}

	sendData = append(sendData, dataCommendLen...)
	sendData = append(sendData, []byte(item.command)...)

	sendData = append(sendData, []byte(c.ctx.Config.BindAddress)...)
	return sendData
}

// 客户端主动发送pull请求到server端
// pull请求到达，说明客户端有能力执行当前的定时任务
// 这个时候可以继续分配定时任务给客户端
// 整个系统才去主动拉取的模式，只有客户端空闲达到一定程度，或者说足以负载当前的任务才会发起pull请求
func (c *AgentController) OnPullCommand(node *agent.TcpClientNode) {
	go func() {
		c.queueMutexLock.Lock()
		func() {
			indexMutex := int64(-1)
			if c.indexMutex >= int64(len(c.queueMutex)-1) {
				atomic.StoreInt64(&c.indexMutex, 0)
			}
			//log.Debugf("c.queueMutex len: %v", len(c.queueMutex))
			for _, queueMutex := range c.queueMutex {
				// 如果有未完成的任务，跳过
				// 这里的正在运行应该有一个超时时间
				// 一般情况下用不着，仅仅为了预防，提高可靠性
				// 最多锁定60秒
				if queueMutex.isRuning && (time.Now().Unix()-queueMutex.start) < 60 {
					//log.Debugf("================%v still running", id)
					continue
				}
				indexMutex++
				if indexMutex >= atomic.LoadInt64(&c.indexMutex) {

					atomic.AddInt64(&c.indexMutex, 1)
					itemI, ok, _ := queueMutex.queue.Get()
					if !ok || itemI == nil {
						//log.Warnf("queue get empty, %+v, %+v, %+v", ok, itemI)
						return
					}
					queueMutex.isRuning = true
					queueMutex.start = time.Now().Unix()
					item := itemI.(*runItem)
					//分发互斥定时任务
					sendData := c.pack(item)

					d := newSendData(agent.CMD_RUN_COMMAND, sendData, node.AsyncSend)
					//log.Debugf("###########dispatch mutex : %+v", *d)
					c.sendQueueLock.Lock()
					c.sendQueue[d.Unique] = d
					c.sendQueueLock.Unlock()
					break
				}

			}
		}()
		c.queueMutexLock.Unlock()
	}()
	go func() {
		c.queueNomalLock.Lock()
		func() {
			index := int64(-1)
			if c.indexNormal >= int64(len(c.queueNomal)-1) {
				atomic.StoreInt64(&c.indexNormal, 0)
			}

			for _, queueNormal := range c.queueNomal {
				index++
				if index != c.indexNormal {
					continue
				}
				atomic.AddInt64(&c.indexNormal, 1)
				itemI, ok, _ := queueNormal.Get()
				if !ok || itemI == nil {
					//log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
					return
				}
				item := itemI.(*runItem)

				////////////////////////

				//sendData := make([]byte, 8)
				//binary.LittleEndian.PutUint64(sendData, uint64(item.id))
				//
				//dataCommendLen := make([]byte, 8)
				//binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))
				//
				//currentTime := make([]byte, 8)
				//binary.LittleEndian.PutUint64(currentTime, uint64(time.Now().Unix()))
				//sendData = append(sendData, currentTime...)
				//
				//sendData = append(sendData, dataCommendLen...)
				//sendData = append(sendData, []byte(item.command)...)

				sendData := c.pack(item) //append(sendData, []byte(c.ctx.Config.BindAddress)...)
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
		}()
		c.queueNomalLock.Unlock()
	}()
}

func (c *AgentController) Pull() {
	//log.Debugf("##############################pull command(%v)", agent.CMD_PULL_COMMAND)
	c.client.Write(agent.Pack(agent.CMD_PULL_COMMAND, []byte("")))
}

func (c *AgentController) Dispatch(id int64, command string, isMutex bool) {
	//logrus.Debugf("Dispatch %v, %v, %v", id, command, isMutex)
	if isMutex {
		c.queueMutexLock.Lock()
		var queueMutex *Mutex = nil
		var ok bool = false
		queueMutex, ok = c.queueMutex[id]
		if !ok {
			queueMutex = &Mutex{
				isRuning:false,
				queue:data.NewQueue(maxQueueLen),
				start:0,
			}
			c.queueMutex[id] = queueMutex
		}
		c.queueMutexLock.Unlock()
		item := &runItem{
				id: id,
				command: command,
				isMutex: isMutex,
			}
			//log.Debugf("dispatch %+v", *item)
		queueMutex.queue.Put(item)
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
