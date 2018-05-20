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
	mlog "models/log"
	"fmt"
	"os"
)

type Controller struct {
	client         *agent.AgentClient
	server         *agent.TcpService
	indexNormal    int64
	indexMutex     int64
	dispatch       chan *runItem
	ctx            *app.Context
	lock           *sync.Mutex
	queueNomalLock *sync.Mutex
	queueNomal     map[int64]*data.EsQueue
	queueMutexLock *sync.Mutex
	queueMutex     map[int64]*Mutex
	sendQueue      map[string]*SendData
	sendQueueLock  *sync.Mutex
	onCronChange   OnCronChangeEventFunc
	onCommand      OnCommandFunc
	addlog         AddLogFunc
	statistics     map[int64]*Statistics
	statisticsLock *sync.Mutex
}

const (
	maxQueueLen     = 64
	dispatchChanLen = 10000
)

type sendFunc              func(data []byte)
type OnCommandFunc         func(id int64, command string, dispatchTime int64, dispatchServer string, runServer string, isMutex byte, after func())
type OnCronChangeEventFunc func(event int, data []byte)
type AddLogFunc            func(cronId int64, output string, useTime int64, dispatchServer, runServer string, rtime int64, event string, remark string)

func NewController(
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
	addlog AddLogFunc,
) *Controller {
	c      := &Controller{
				indexNormal:    0,
				indexMutex:     0,
				dispatch:       make(chan *runItem, dispatchChanLen),
				ctx:            ctx,
				lock:           new(sync.Mutex),
				queueNomal:     make(map[int64]*data.EsQueue),
				queueMutex:     make(map[int64]*Mutex),
				queueNomalLock: new(sync.Mutex),
				queueMutexLock: new(sync.Mutex),
				sendQueue:      make(map[string]*SendData),
				sendQueueLock:  new(sync.Mutex),
				onCronChange:   onCronChange,
				onCommand:      onCommand,
				addlog:         addlog,
				statistics:     make(map[int64]*Statistics),
				statisticsLock: new(sync.Mutex),
			}
	c.server = agent.NewAgentServer(ctx.Context(), ctx.Config.BindAddress, agent.SetOnServerEvents(c.onServerEvent), )
	c.client = agent.NewAgentClient(ctx.Context(), agent.SetGetLeader(getLeader), agent.SetOnClientEvent(c.onClientEvent), )
	cpu := runtime.NumCPU()
	for i:= 0; i < cpu; i++ {
		go c.sendService()
	}
	return c
}

func (c *Controller) onClientEvent(tcp *agent.AgentClient, cmd int , content []byte) {
	//log.Debugf("#############client receive: cmd=%d, content=%v", cmd, string(content))
	switch cmd {
	case agent.CMD_RUN_COMMAND:
			var sendData SendData
			err := json.Unmarshal(content, &sendData)
			if err != nil {
				log.Errorf("json.Unmarshal with %v", err)
				return
			}
			id, dispatchTime, isMutex, command, dispatchServer, err := unpack(sendData.Data)
			if err != nil {
				return
			}

		c.addlog(id, "", 0, dispatchServer, c.ctx.Config.BindAddress, int64(time.Now().UnixNano()/1000000), mlog.EVENT_CRON_RUN, "定时任务开始运行 - 3")


		sdata := make([]byte, 0)
			sid   := make([]byte, 8)
			binary.LittleEndian.PutUint64(sid, uint64(id))
			sdata = append(sdata, sid...)
			sdata = append(sdata, isMutex)
			sdata = append(sdata, []byte(sendData.Unique)...)

			c.onCommand(id, command, dispatchTime, dispatchServer, c.ctx.Config.BindAddress, isMutex, func() {
				tcp.Write(agent.Pack(agent.CMD_RUN_COMMAND, sdata))
			})
	case agent.CMD_CRONTAB_CHANGE:
		log.Infof("cron send to leader server ok (will delete from send queue): %+v", string(content))
		c.sendQueueLock.Lock()
		delete(c.sendQueue, string(content))
		c.sendQueueLock.Unlock()
	}
}

func (c *Controller) onServerEvent(node *agent.TcpClientNode, event int, content []byte) {
	//log.Debugf("server receive:, %v, %v", event, content )
	switch event {
	case agent.CMD_PULL_COMMAND:
		//start := time.Now()
		c.OnPullCommand(node)
		//fmt.Fprintf(os.Stderr, "OnPullCommand use time %v\n", time.Since(start))
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
		id      := int64(binary.LittleEndian.Uint64(content[:8]))
		isMutex := content[8]
		unique  := string(content[9:])

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

		current := int64(time.Now().UnixNano() / 1000000)
		c.addlog(id, "", 0, c.ctx.Config.BindAddress, "", current, mlog.EVENT_CRON_END, "定时任务结束 - 5")
		//log.Debugf("****************************command run back 4 => %v, command back time is %v", unique, time.Now().UnixNano())
		c.statisticsLock.Lock()
		st, ok := c.statistics[id]
		if ok {
			st.totalUseTime += current - st.startTime
			fmt.Fprintf(os.Stderr, "%v avg use time = %vms\n", id, st.getAvg())
		}
		c.statisticsLock.Unlock()

	}
}

// send data to leader
func (c *Controller) SendToLeader(data []byte) {
	//c.client.Send(agent.CMD_CRONTAB_CHANGE, data)
	d := newSendData(agent.CMD_CRONTAB_CHANGE, data, c.client.Write, 0)
	c.sendQueueLock.Lock()
	c.sendQueue[d.Unique] = d
	c.sendQueueLock.Unlock()
}

func (c *Controller) sendService() {
	for {
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

			// 这里获取运行的平均时间，假设为t， 然后 t+60*1000 毫秒为超时时间
			c.statisticsLock.Lock()
			var timeout int64 = 0
			sta, ok := c.statistics[d.CronId]
			if ok {
				timeout = sta.getAvg() + 60 * 1000
			}
			c.statisticsLock.Unlock()

			if d.Status > 0 && (int64(time.Now().UnixNano()/1000000) - d.Time) <= timeout {
				times3++
				continue
			}
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

			//Start := int64(time.Now().UnixNano()/1000000)

			d.Time    = int64(time.Now().UnixNano()/1000000)
			sd       := d.encode()
			sendData := agent.Pack(d.Cmd, sd)

			//一个定时任务的运行周期从 mlog.EVENT_CRON_DISPATCH 开始到 mlog.EVENT_CRON_END 结束
			//todo 添加关键日志
			if d.CronId > 0 {
				c.addlog(d.CronId, "", 0, c.ctx.Config.BindAddress, "", int64(time.Now().UnixNano()/1000000), mlog.EVENT_CRON_DISPATCH, "定时任务分发 - 2")
			}


			c.statisticsLock.Lock()
			st, ok := c.statistics[d.CronId]
			if !ok {
				st = &Statistics{}
				c.statistics[d.CronId] = st
			}
			st.sendTimes++
			st.startTime = int64(time.Now().UnixNano()/1000000)
			c.statisticsLock.Unlock()

			d.send(sendData)


		}
		// 如果都是发送中，这里尝试等待10毫秒，让出cpu
		if times3 >= len(c.sendQueue) {
			time.Sleep(time.Millisecond * 10)
		}
		c.sendQueueLock.Unlock()
	}
}

// 客户端主动发送pull请求到server端
// pull请求到达，说明客户端有能力执行当前的定时任务
// 这个时候可以继续分配定时任务给客户端
// 整个系统才去主动拉取的模式，只有客户端空闲达到一定程度，或者说足以负载当前的任务才会发起pull请求
func (c *Controller) OnPullCommand(node *agent.TcpClientNode) {
	go func() {
		//start := time.Now()
		c.queueMutexLock.Lock()
		func() {
			indexMutex := int64(-1)
			if c.indexMutex >= int64(len(c.queueMutex)-1) {
				atomic.StoreInt64(&c.indexMutex, 0)
			}
			//log.Debugf("c.queueMutex len: %v", len(c.queueMutex))
			for id, queueMutex := range c.queueMutex {
				indexMutex++
				// 如果有未完成的任务，跳过
				// 这里的正在运行应该有一个超时时间
				// 一般情况下用不着，仅仅为了预防，提高可靠性
				// 最多锁定60秒

				// 获取平均原型周期 + 60s最为超时标准
				c.statisticsLock.Lock()
				var timeout int64 = 0
				sta, ok := c.statistics[id]
				if ok {
					timeout = sta.getAvg() + 60 * 1000
				}
				c.statisticsLock.Unlock()

				if queueMutex.isRuning && (int64(time.Now().UnixNano()/1000000) - queueMutex.start) < timeout {
					//log.Debugf("================%v still running", id)
					continue
				}
				if indexMutex >= atomic.LoadInt64(&c.indexMutex) {

					atomic.AddInt64(&c.indexMutex, 1)
					itemI, ok, _ := queueMutex.queue.Get()
					if !ok || itemI == nil {
						//log.Warnf("queue get empty, %+v, %+v, %+v", ok, itemI)
						continue
					}
					queueMutex.isRuning = true
					queueMutex.start = int64(time.Now().UnixNano()/1000000)//time.Now().Unix()
					item := itemI.(*runItem)
					//分发互斥定时任务
					sendData := pack(item, c.ctx.Config.BindAddress)

					d := newSendData(agent.CMD_RUN_COMMAND, sendData, node.AsyncSend, item.id)
					//log.Debugf("###########dispatch mutex : %+v", *d)
					c.sendQueueLock.Lock()
					c.sendQueue[d.Unique] = d
					c.sendQueueLock.Unlock()
					break
				}

			}
		}()
		c.queueMutexLock.Unlock()
		//fmt.Fprintf(os.Stderr, "OnPullCommand mutex use time %v\n", time.Since(start))
	}()
	go func() {
		//start := time.Now()
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
					continue
				}
				item := itemI.(*runItem)
				sendData := pack(item, c.ctx.Config.BindAddress)

				d := newSendData(agent.CMD_RUN_COMMAND, sendData, node.AsyncSend, item.id) //c.server.Broadcast)//
				c.sendQueueLock.Lock()
				c.sendQueue[d.Unique] = d
				c.sendQueueLock.Unlock()
				break
			}
		}()
		c.queueNomalLock.Unlock()
		//fmt.Fprintf(os.Stderr, "OnPullCommand normal use time %v\n", time.Since(start))
	}()
}

func (c *Controller) Pull() {
	//log.Debugf("##############################pull command(%v)", agent.CMD_PULL_COMMAND)
	c.client.Write(agent.Pack(agent.CMD_PULL_COMMAND, []byte("")))
}

func (c *Controller) Dispatch(id int64, command string, isMutex bool) {
	//logrus.Debugf("Dispatch %v, %v, %v", id, command, isMutex)
	if isMutex {
		c.queueMutexLock.Lock()
		var queueMutex *Mutex = nil
		var ok = false
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
func (c *Controller) OnLeader(isLeader bool) {
	c.client.OnLeader(isLeader)
}

// start agent
func (c *Controller) Start() {
	c.server.Start()
}

// close agent
func (c *Controller) Close() {
	c.server.Close()
}
