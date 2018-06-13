package agent

import (
	"library/agent"
	"app"
	"encoding/binary"
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"github.com/jilieryuyi/wing-go/tcp"
)

type Controller struct {
	client              *tcp.Client
	server              *tcp.Server
	dispatch            chan *runItem
	onPullChan          chan *tcp.ClientNode
	runningEndChan      chan int64
	sendQueueChan       chan *SendData
	delSendQueueChan    chan string
	statisticsStartChan chan []byte
	statisticsEndChan   chan []byte
	ctx                 *app.Context
	lock                *sync.Mutex
	onCronChange        OnCronChangeEventFunc
	onCommand           OnCommandFunc
	sendQueueLen        int64
	getLeader           GetLeaderFunc
	onDispatch          OnDispatchFunc
	OnCommandBack       OnCommandBackFunc
	codec ICodec
}

const (
	maxQueueLen             = 64
	dispatchChanLen         = 1000000
	onPullChanLen           = 128
	runningEndChanLen       = 1000
	sendQueueChanLen        = 1000
	delSendQueueChanLen     = 1000
	statisticsChanLen       = 1000
)

type OnDispatchFunc        func(cronId int64)
type OnCommandBackFunc     func(cronId int64, dispatchServer string)
type sendFunc              func(data []byte)  (int, error)
type OnCommandFunc         func(id int64, command string, dispatchServer string, runServer string, isMutex byte, after func())
type OnCronChangeEventFunc func(event int, data []byte)
type GetLeaderFunc         func()(string, int, error)

func NewController(
	ctx *app.Context,
	//这个参数提供了查询leader的服务ip和端口
	//最终用于 client连接leader时使用
	//来源于consulControl.GetLeader
	getLeader GetLeaderFunc,
	//http api增删改查定时任务会触发这个事件，最终这个事件影响到leader的定时任务
	//最终落入这个api crontabController.Add
	onCronChange OnCronChangeEventFunc,
	//这个事件由leader分发定时任务到节点，节点收到定时任务时触发
	//最终接收的api是crontabController.ReceiveCommand
	onCommand OnCommandFunc,
	//addlog AddLogFunc,
	onDispatch OnDispatchFunc,
	OnCommandBack    OnCommandBackFunc,
) *Controller {
	c := &Controller{
			dispatch:            make(chan *runItem, dispatchChanLen),
			onPullChan:          make(chan *tcp.ClientNode, onPullChanLen),
			runningEndChan:      make(chan int64, runningEndChanLen),
			sendQueueChan:       make(chan *SendData, sendQueueChanLen),
			delSendQueueChan:    make(chan string, delSendQueueChanLen),
			statisticsStartChan: make(chan []byte, statisticsChanLen),
			statisticsEndChan:   make(chan []byte, statisticsChanLen),
			ctx:                ctx,
			lock:               new(sync.Mutex),
			onCronChange:       onCronChange,
			onCommand:          onCommand,
			sendQueueLen:       0,
			getLeader:          getLeader,
			onDispatch:	        onDispatch,
			OnCommandBack:      OnCommandBack,
			codec:              &Codec{},
		}
	c.server = tcp.NewServer(ctx.Context(), ctx.Config.BindAddress, tcp.SetOnServerMessage(c.OnServerMessage))
	c.client = tcp.NewClient(ctx.Context())
	go c.sendService()
	go c.keep()
	return c
}

func (c *Controller) onClientEvent(tcp *tcp.Client, cmd int , content []byte) {
	switch cmd {
	case agent.CMD_RUN_COMMAND:
		var sendData SendData
		err := json.Unmarshal(content, &sendData)
		if err != nil {
			log.Errorf("json.Unmarshal with %v", err)
			return
		}
		id, isMutex, command, dispatchServer, err := unpack(sendData.Data)
		if err != nil {
			log.Errorf("%v", err)
			return
		}
		fmt.Fprintf(os.Stderr, "receive command, %v, %v, %v, %v, %v\r\n", id, isMutex, command, dispatchServer, err)
		sdata := make([]byte, 0)
		sid   := make([]byte, 8)
		binary.LittleEndian.PutUint64(sid, uint64(id))
		sdata = append(sdata, sid...)
		sdata = append(sdata, isMutex)
		sdata = append(sdata, []byte(sendData.Unique)...)
		c.onCommand(id, command, dispatchServer, c.ctx.Config.BindAddress, isMutex, func() {
			sd, _ := c.codec.Encode(agent.CMD_RUN_COMMAND, sdata)
			tcp.Send(sd)
		})
		fmt.Fprintf(os.Stderr, "receive command run end, %v, %v, %v, %v, %v\r\n", id, isMutex, command, dispatchServer, err)
	case agent.CMD_CRONTAB_CHANGE_OK:
		log.Infof("cron send to leader server ok (will delete from send queue): %+v", string(content))
		c.delSendQueueChan <-  string(content)
	case agent.CMD_CRONTAB_CHANGE:
		var sdata SendData
		err := json.Unmarshal(content, &sdata)
		if err != nil {
			log.Errorf("%+v", err)
		} else {
			event := binary.LittleEndian.Uint32(sdata.Data[:4])
			go c.onCronChange(int(event), sdata.Data[4:])
		}
	}
}

func (c *Controller) OnServerMessage(node *tcp.ClientNode, msgId int64, content []byte) {
	// content 二次解析后得到event
	// 这里的content全部使用json格式发送
	event, data, err := c.codec.Decode(content)
	if err != nil {
		return
	}
	//data := dataRaw.(Package)
	switch event {
	case agent.CMD_PULL_COMMAND:
		if len(c.onPullChan) < 32 {
			c.onPullChan <- node
		}
	case agent.CMD_CRONTAB_CHANGE:
		sdata := data.(SendData)
		//err := json.Unmarshal(content, &sdata)
		//if err != nil {
		//	log.Errorf("%+v", err)
		//} else {
		sd, err := c.codec.Encode(agent.CMD_CRONTAB_CHANGE_OK, sdata.Unique)
		if err == nil {
			node.AsyncSend(msgId, sd)
		}
		sd, err = c.codec.Encode(agent.CMD_CRONTAB_CHANGE, content)
		if err == nil {
			c.server.Broadcast(msgId, sd)
		}
		//}
	case agent.CMD_RUN_COMMAND:
		id      := int64(binary.LittleEndian.Uint64(content[:8]))
		isMutex := content[8]
		unique  := string(content[9:])
		fmt.Fprintf(os.Stderr, "receive run command end %v, %v, %v\r\n", id, isMutex, unique)

		if isMutex == 1 {
			sdata := make([]byte, 16)
			binary.LittleEndian.PutUint64(sdata[:8], uint64(id))
			binary.LittleEndian.PutUint64(sdata[8:], uint64(int64(time.Now().UnixNano() / 1000000)))
			c.statisticsEndChan <- sdata
			c.runningEndChan <- id
		}
		c.delSendQueueChan <- unique
		//定时任务运行完返回server端（leader）
		c.OnCommandBack(id, c.ctx.Config.BindAddress)
	}
}

// send data to leader
func (c *Controller) SyncToLeader(data []byte) {
	d := newSendData(agent.CMD_CRONTAB_CHANGE, data, func(data []byte) (int, error) {
		c.client.AsyncSend(data)
		return 0, nil
	}, 0, false)
	c.sendQueueChan <- d
}

// 这个api用来发送获取需要执行的定时任务
// 由crontab调用
// 一旦crontab执行完一定程度的定时任务，变得空闲就会主动获取新的定时任务
// 这个api就是发起主动获取请求
func (c *Controller) Pull() {
	sd, _ := c.codec.Encode(agent.CMD_PULL_COMMAND, "")
	c.client.AsyncSend(sd)
}

func (c *Controller) sendService() {
	var sendQueue = make(map[string]*SendData)
	var checkChan = make(chan struct{})
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
					if d.Status > 0 && (int64(time.Now().UnixNano()/1000000) - d.Time) <= 10000 {
						fmt.Fprintf(os.Stderr, "%v is still in sending status, wait for back\r\n", d.CronId)
						continue
					}
					d.Status = 1
					d.SendTimes++

					if d.SendTimes > 1 {
						log.Warnf("send times %v, %+v", d.SendTimes, *d)
					}
					d.Time    = int64(time.Now().UnixNano() / 1000000)
					sd       := d.encode()
					sendData := agent.Pack(d.Cmd, sd)
					if d.IsMutex {
						sdata := make([]byte, 16)
						binary.LittleEndian.PutUint64(sdata[:8], uint64(d.CronId))
						binary.LittleEndian.PutUint64(sdata[8:], uint64(int64(time.Now().UnixNano()/1000000)))
						c.statisticsStartChan <- sdata
					}
					d.send(sendData)
					delete(sendQueue, d.Unique)
					fmt.Fprintf(os.Stderr, "send use time %v\n", time.Since(start))
				}
				atomic.StoreInt64(&c.sendQueueLen, int64(len(sendQueue)))
			case unique, ok := <- c.delSendQueueChan:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "running complete -server %v\r\n", unique)
		}
	}
}

func (c *Controller) keep() {
	var queueMutex   = make(QMutex)
	var queueNomal   = make(QEs)
	var gindexMutex  = int64(0)
	var gindexNormal = int64(0)
	// subnum for wait queue len
	var setNum       = make(map[int64] func() int64)
	var mutexKeys    = make([]int64, 0)
	var normalKeys   = make([]int64, 0)

	for {
		select {
		case node, ok := <-c.onPullChan:
			if !ok {
				return
			}
			if atomic.LoadInt64(&c.sendQueueLen) < 32 {
				if len(mutexKeys) > 0 {
					start := time.Now()
					id := mutexKeys[int(gindexMutex)]
					queueMutex.dispatch(id, c.ctx.Config.BindAddress, func(data []byte) (int, error) {
						node.Send(1, data)
						return len(data), nil
					}, c.sendQueueChan, func(item *runItem) {
						set, ok := setNum[id]
						if ok {
							set()
						} else {
							log.Errorf("%v set num does not exists", id)
						}
						// add log 这里代表定时任务被发出去了
						c.onDispatch(item.id)
					})
					fmt.Fprintf(os.Stderr, "dispatch id= %v, OnPullCommand mutex use time %v\n", id, time.Since(start))
					gindexMutex++
					if gindexMutex >= int64(len(mutexKeys)) {
						gindexMutex = 0
					}
				}

				if len(normalKeys) > 0 {
					start := time.Now()
					id := normalKeys[int(gindexNormal)]
					queueNomal.dispatch(id, c.ctx.Config.BindAddress, func(data []byte) (int, error) {
						node.Send(1, data)
						return len(data), nil
					}, c.sendQueueChan, func(item *runItem) {
						set, ok := setNum[id]
						if ok {
							set()
						} else {
							log.Errorf("%v set num does not exists", id)
						}
						// add log 这里代表定时任务被发出去了
						c.onDispatch(item.id)
					})
					gindexNormal++
					if gindexNormal >= int64(len(normalKeys)) {
						gindexNormal = 0
					}
					fmt.Fprintf(os.Stderr, "OnPullCommand normal use time %v\n", time.Since(start))
				}
			}

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
			t  := int64(binary.LittleEndian.Uint64(sdata[8:])) //, uint64(int64(time.Now().UnixNano() / 1000000)))
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
			sta, ok := queueMutex[id]
			if ok {
				sta.sta.totalUseTime += t - sta.sta.startTime
			} else {
				log.Errorf("%v does not exists", id)
			}
		case item, ok := <-c.dispatch:
			if !ok {
				return
			}
			setNum[item.id] = item.subWaitNum
			if item.isMutex {
				if _, ok := queueMutex[item.id]; !ok {
					mutexKeys = append(mutexKeys, item.id)
				}
				if !queueMutex.append(item) {
					item.subWaitNum()
				}
			} else {
				if _, ok := queueNomal[item.id]; !ok {
					normalKeys = append(normalKeys, item.id)
				}
				if !queueNomal.append(item) {
					item.subWaitNum()
				}
			}
		}
	}
}

func (c *Controller) Dispatch(id int64, command string, isMutex bool, addWaitNum func(), subwaitNum func() int64) {
	if len(c.dispatch) >= cap(c.dispatch) {
		log.Errorf("dispatch cache full")
		return
	}
	addWaitNum()
	item := &runItem{
		id:         id,
		command:    command,
		isMutex:    isMutex,
		subWaitNum: subwaitNum,
	}
	c.dispatch <- item
}

// set on leader select callback
func (c *Controller) OnLeader(isLeader bool) {
	go func() {
		log.Debugf("==============agent client OnLeader %v===============", isLeader)
		var ip string
		var port int
		for {
			ip, port, _ = c.getLeader()
			if ip == "" || port <= 0 {
				log.Warnf("ip or port empty: %v, %v, wait for init", ip, port)
				time.Sleep(time.Second * 1)
				continue
			}
			break
		}
		log.Infof("leader %v:%v", ip, port)
		c.client.Connect(fmt.Sprintf("%v:%v", ip, port), time.Second * 3)
	}()
}

// start agent
func (c *Controller) Start() {
	c.server.Start()
}

// close agent
func (c *Controller) Close() {
	c.server.Close()
}
