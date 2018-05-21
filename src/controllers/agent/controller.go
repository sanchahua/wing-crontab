package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	mlog "models/log"
	"fmt"
	"os"
)

type Controller struct {
	client           *agent.AgentClient
	server           *agent.TcpService
	indexNormal      int64
	indexMutex       int64

	dispatch         chan *runItem
	onPullChan       chan *agent.TcpClientNode
	runningEndChan   chan int64
	sendQueueChan    chan *SendData
	delSendQueueChan chan string
	statisticsStartChan   chan []byte
	statisticsEndChan   chan []byte

	ctx              *app.Context
	lock             *sync.Mutex
	onCronChange     OnCronChangeEventFunc
	onCommand        OnCommandFunc
	addlog           AddLogFunc
	statistics       map[int64]*Statistics
	statisticsLock   *sync.Mutex
}

const (
	maxQueueLen         = 64
	dispatchChanLen     = 10000
	onPullChanLen       = 128
	runningEndChanLen   = 1000
	sendQueueChanLen    = 1000
	delSendQueueChanLen = 1000
	statisticsChanLen   = 1000
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

				dispatch:         make(chan *runItem, dispatchChanLen),
				onPullChan:       make(chan *agent.TcpClientNode, onPullChanLen),
				runningEndChan:   make(chan int64, runningEndChanLen),
				sendQueueChan:    make(chan *SendData, sendQueueChanLen),
				delSendQueueChan: make(chan string, delSendQueueChanLen),
		statisticsStartChan:   make(chan []byte, statisticsChanLen),
		statisticsEndChan:   make(chan []byte, statisticsChanLen),

				ctx:            ctx,
				lock:           new(sync.Mutex),
				onCronChange:   onCronChange,
				onCommand:      onCommand,
				addlog:         addlog,
				statistics:     make(map[int64]*Statistics),
				statisticsLock: new(sync.Mutex),
			}
	c.server = agent.NewAgentServer(ctx.Context(), ctx.Config.BindAddress, agent.SetOnServerEvents(c.onServerEvent), )
	c.client = agent.NewAgentClient(ctx.Context(), agent.SetGetLeader(getLeader), agent.SetOnClientEvent(c.onClientEvent), )
	go c.sendService()
	go c.keep()
	return c
}

func (c *Controller) onClientEvent(tcp *agent.AgentClient, cmd int , content []byte) {
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
				log.Errorf("%v", err)
				return
			}
		fmt.Fprintf(os.Stderr, "receive command, %v, %v, %v, %v, %v,%v,%v\r\n", id, dispatchTime, isMutex, command, dispatchServer, err)

		c.addlog(id, "", 0, dispatchServer, c.ctx.Config.BindAddress, int64(time.Now().UnixNano()/1000000), mlog.EVENT_CRON_RUN, "定时任务开始运行 - 3")


		sdata := make([]byte, 0)
		sid   := make([]byte, 8)
		binary.LittleEndian.PutUint64(sid, uint64(id))
		sdata = append(sdata, sid...)
		sdata = append(sdata, isMutex)
		sdata = append(sdata, []byte(sendData.Unique)...)

		c.onCommand(id, command, dispatchTime, dispatchServer, c.ctx.Config.BindAddress, isMutex, func() {
			log.Debugf("command run send %v", sendData.Unique)
			tcp.Write(agent.Pack(agent.CMD_RUN_COMMAND, sdata))
		})
		fmt.Fprintf(os.Stderr, "receive command run end, %v, %v, %v, %v, %v,%v,%v\r\n", id, dispatchTime, isMutex, command, dispatchServer, err)

	case agent.CMD_CRONTAB_CHANGE:
		//
		log.Infof("cron send to leader server ok (will delete from send queue): %+v", string(content))
		c.delSendQueueChan <-  string(content)
	}
}

func (c *Controller) onServerEvent(node *agent.TcpClientNode, event int, content []byte) {
	//log.Debugf("###################server receive:%v, %v==CMD_PULL_COMMAND=%v", event, content,agent.CMD_PULL_COMMAND)
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
		fmt.Fprintf(os.Stderr, "receive run command end %v, %v, %v\r\n", id, isMutex, unique)

		if isMutex == 1 {
			//log.Debugf("set is running false")
			//c.queueMutexLock.Lock()
			//m ,ok := c.queueMutex[id]
			//if ok {
			//	m.isRuning = false
			//} else {
			//	log.Errorf("%v does not exists")
			//}
			//c.queueMutexLock.Unlock()

			sdata := make([]byte, 16)
			binary.LittleEndian.PutUint64(sdata[:8], uint64(id))
			binary.LittleEndian.PutUint64(sdata[8:], uint64(int64(time.Now().UnixNano() / 1000000)))

			c.statisticsEndChan <-	sdata

			c.runningEndChan <- id
		}

		//log.Debugf("command is run (will delete from send queue): %v", string(content))
		//c.sendQueueLock.Lock()
		//_, ex := c.sendQueue[unique]
		//if ex {
		//	delete(c.sendQueue, unique)
		//} else {
		//	log.Errorf("does not int send queue: %v", unique)
		//}
		////log.Debugf("send queue len: %v", len(c.sendQueue))
		//c.sendQueueLock.Unlock()

		c.delSendQueueChan <- unique

		//todo
		// 如果send queue里面存在这个消息才是正常的返回值
		// 后续返回握手也可能加入重发机制，所以这个判断很重要
		 {
			//current := int64(time.Now().UnixNano() / 1000000)
			c.addlog(id, "", 0, c.ctx.Config.BindAddress, "", int64(time.Now().UnixNano() / 1000000), mlog.EVENT_CRON_END, "定时任务结束 - 5")
			//log.Debugf("****************************command run back 4 => %v, command back time is %v", unique, time.Now().UnixNano())
			//c.statisticsLock.Lock()
			//st, ok := c.statistics[id]
			//if ok {
			//	st.totalUseTime += current - st.startTime
			//	fmt.Fprintf(os.Stderr, "%v avg use time = %vms\n", id, st.getAvg())
			//} else {
			//	log.Errorf("%v does not exists", id)
			//}
			//c.statisticsLock.Unlock()
			//c.setStatisticsTime(id)
		}

	}
}

// send data to leader
func (c *Controller) SendToLeader(data []byte) {
	d := newSendData(agent.CMD_CRONTAB_CHANGE, data, c.client.Write, 0, false)
	c.sendQueueChan <- d
}


// 客户端主动发送pull请求到server端
// pull请求到达，说明客户端有能力执行当前的定时任务
// 这个时候可以继续分配定时任务给客户端
// 整个系统才去主动拉取的模式，只有客户端空闲达到一定程度，或者说足以负载当前的任务才会发起pull请求
func (c *Controller) OnPullCommand(node *agent.TcpClientNode) {
	//log.Debugf("ou pull")
	c.onPullChan <- node
}

func (c *Controller) Pull() {
	//log.Debugf("##############################pull command(%v)", agent.CMD_PULL_COMMAND)
	c.client.Write(agent.Pack(agent.CMD_PULL_COMMAND, []byte("")))
}

func (c *Controller) setStatistics(id int64) {
	c.statisticsLock.Lock()
	st, ok := c.statistics[id]
	if !ok {
		st = &Statistics{}
		c.statistics[id] = st
	}
	st.sendTimes++
	st.startTime = int64(time.Now().UnixNano() / 1000000)
	c.statisticsLock.Unlock()
}

func (c *Controller) setStatisticsTime(id int64) {
	c.statisticsLock.Lock()
	st, ok := c.statistics[id]
	if ok {
		current := int64(time.Now().UnixNano() / 1000000)
		st.totalUseTime += current - st.startTime
		fmt.Fprintf(os.Stderr, "%v avg use time = %vms\n", id, st.getAvg())
	} else {
		log.Errorf("%v does not exists", id)
	}
	c.statisticsLock.Unlock()
}

func (c *Controller) sendService() {

	var sendQueue = make(map[string]*SendData)
	var checkChan = make(chan struct{}, 1000)

	// 信号生成，用于触发发送待发送的消息
	go func() {
		for {
			checkChan <- struct{}{}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	for {
		select {
		case d ,ok := <-c.sendQueueChan:
			if !ok {
				return
			}
			sendQueue[d.Unique] = d
			case _, ok:= <-checkChan:
				if !ok {
					return
				}

			for _, d := range sendQueue {
				start := time.Now()
				// status > 0 is sending
				// 发送中的数据，3秒之内不会在发送，超过3秒会进行2次重试
				// todo ？？这里的3秒设置的是否合理，这里最好的方式应该有一个实时发送时间反馈
				// 比如完成一次发送需要100ms，超时时间设置为 100ms + 3s 这样应该更合理
				// 即t+3模式
				// 默认60秒超时重试

				// 这里获取运行的平均时间，假设为t， 然后 t+60*1000 毫秒为超时时间
				//c.statisticsLock.Lock()
				//var timeout = c.getTimeout(d.CronId)//int64 = 60 * 1000
				//sta, ok := c.statistics[d.CronId]
				//if ok {
				//	avg := sta.getAvg()
				//	if avg > 0 {
				//		timeout = avg * 3
				//		if timeout > avg+60*1000 {
				//			timeout = avg + 60*1000
				//		} else if timeout < 300 {
				//			timeout = 1000
				//		}
				//	}
				//}
				//c.statisticsLock.Unlock()

				//if d.Status > 0 && (int64(time.Now().UnixNano()/1000000) - d.Time) <= timeout {
				//
				//	//fmt.Fprintf(os.Stderr, "%v is still sending, wait for back\r\n", d.CronId)
				//	continue
				//}
				//d.Status = 1
				//d.SendTimes++
				//
				//if d.SendTimes > 1 {
				//	log.Warnf("send times %v, %+v", d.SendTimes, *d)
				//}
				//
				//// 每次延迟3秒重试，最多20次，即1分钟之内才会重试
				//if d.SendTimes >= 60 {
				//	delete(sendQueue, d.Unique)
				//	log.Warnf("send times max then 60, delete %+v", *d)
				//	continue
				//}

				//Start := int64(time.Now().UnixNano()/1000000)

				d.Time    = int64(time.Now().UnixNano() / 1000000)
				sd       := d.encode()
				sendData := agent.Pack(d.Cmd, sd)

				//一个定时任务的运行周期从 mlog.EVENT_CRON_DISPATCH 开始到 mlog.EVENT_CRON_END 结束
				//todo 添加关键日志
				if d.CronId > 0 {
					c.addlog(d.CronId, "", 0, c.ctx.Config.BindAddress, "", int64(time.Now().UnixNano()/1000000), mlog.EVENT_CRON_DISPATCH, "定时任务分发 - 2")
				}

				//c.setStatistics(d.CronId)
				//st.sendTimes++
				//st.startTime = int64(time.Now().UnixNano() / 1000000)

				if d.IsMutex {
					sdata := make([]byte, 16)
					binary.LittleEndian.PutUint64(sdata[:8], uint64(d.CronId))
					binary.LittleEndian.PutUint64(sdata[8:], uint64(int64(time.Now().UnixNano()/1000000)))
					c.statisticsStartChan <- sdata
				}

				log.Debugf("#################################send %+v", *d)
				d.send(sendData)
				delete(sendQueue, d.Unique)
				fmt.Fprintf(os.Stderr, "send use time %v\n", time.Since(start))

			}
			case unique, ok := <- c.delSendQueueChan:
				if !ok {
					return
				}
				log.Debugf("running complete %v", unique)
				//log.Debugf("=========================delete from send queue %v", unique)
				//_, exists := sendQueue[unique]
				//if exists {
				//	delete(sendQueue, unique)
				//} else {
				//	log.Errorf("does not in send queue %v", unique)
				//}
		}
	}
}



func (c *Controller) getTimeout(id int64) int64 {
	c.statisticsLock.Lock()
	var timeout int64 = 60 * 1000
	sta, ok := c.statistics[id]
	if ok {
		avg := sta.getAvg()
		if avg > 0 {
			timeout = avg * 3
			if timeout > avg + 60 * 1000 {
				timeout = avg + 60 * 1000
			} else if timeout < 300 {
				timeout = 1000
			}
		}
	}
	c.statisticsLock.Unlock()
	log.Debugf("%v avg timeout is %v", id, timeout)
	return timeout
}

func (c *Controller) keep() {
	var queueMutex   = make(QMutex)
	var queueNomal   = make(QEs)
	var gindexMutex  = int64(0)
	var gindexNormal = int64(0)

	for {
		select {
		case item, ok := <-c.dispatch:
			if !ok {
				return
			}
			if item.isMutex {
				queueMutex.append(item)
			} else {
				queueNomal.append(item)
			}
		case node, ok := <-c.onPullChan:
			if !ok {
				return
			}
			queueMutex.dispatch(&gindexMutex,  c.ctx.Config.BindAddress, node.AsyncSend, c.sendQueueChan)
			queueNomal.dispatch(&gindexNormal, c.ctx.Config.BindAddress, node.AsyncSend, c.sendQueueChan)
		case endId, ok := <-c.runningEndChan:
			if !ok {
				return
			}
			queueMutex.setRunning(endId, false)
		case sdata, ok := <-c.statisticsStartChan:
			if !ok {
				return
			}

			id := int64(binary.LittleEndian.Uint64(sdata[:8]))
			t := int64(binary.LittleEndian.Uint64(sdata[8:])) //, uint64(int64(time.Now().UnixNano() / 1000000)))

			log.Debugf(" %v start at %v", id, t)
			sta, ok := queueMutex[id]
			if ok {
				sta.sta.sendTimes++
				sta.sta.startTime = t
			} else {
				log.Errorf("%v does not exists", id)
			}

		case sdata, ok := <- c.statisticsEndChan:
			if !ok {
				return
			}

			id := int64(binary.LittleEndian.Uint64(sdata[:8]))
			t  := int64(binary.LittleEndian.Uint64(sdata[8:])) //, uint64(int64(time.Now().UnixNano() / 1000000)))
			log.Debugf(" %v end at %v", id, t)

			sta, ok := queueMutex[id]
			if ok {
				sta.sta.totalUseTime += t - sta.sta.startTime
				log.Debugf("#############avg=%v", sta.sta.getAvg())
			} else {
				log.Errorf("%v does not exists", id)
			}
		}
	}
}

func (c *Controller) Dispatch(id int64, command string, isMutex bool, logId int64) {
	item := &runItem{
		id:      id,
		command: command,
		isMutex: isMutex,
		logId:   logId,
	}
	c.dispatch <- item
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
