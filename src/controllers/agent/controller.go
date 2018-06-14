package agent

import (
	"app"
	"encoding/binary"
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
	"models/cron"
	"fmt"
	"os"
	"sync/atomic"
	"github.com/jilieryuyi/wing-go/tcp"
)

type Controller struct {
	client              *tcp.Client
	server              *tcp.Server
	dispatch            chan *runItem
	onPullChan          chan message
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
const (
	CMD_ERROR = iota + 1    // 错误响应
	CMD_TICK                // 心跳包
	CMD_AGENT
	CMD_STOP
	CMD_RELOAD
	CMD_SHOW_MEMBERS
	CMD_CRONTAB_CHANGE
	CMD_RUN_COMMAND
	CMD_PULL_COMMAND
	CMD_DEL_CACHE
	CMD_CRONTAB_CHANGE_OK
)
type OnDispatchFunc        func(cronId int64)
type OnCommandBackFunc     func(cronId int64, dispatchServer string)
type sendFunc              func(msgId int64, data []byte)  (int, error)
type OnCommandFunc         func(id int64, command string, dispatchServer string, runServer string, isMutex byte, after func())
type OnCronChangeEventFunc func(event int, data *cron.CronEntity)
type GetLeaderFunc         func()(string, int, error)
type message struct {
	node *tcp.ClientNode
	msgId int64
}


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
			dispatch:            make(chan *runItem,  dispatchChanLen),
			onPullChan:          make(chan message,   onPullChanLen),
			runningEndChan:      make(chan int64,     runningEndChanLen),
			sendQueueChan:       make(chan *SendData, sendQueueChanLen),
			delSendQueueChan:    make(chan string,    delSendQueueChanLen),
			statisticsStartChan: make(chan []byte,    statisticsChanLen),
			statisticsEndChan:   make(chan []byte,    statisticsChanLen),
			ctx:                 ctx,
			lock:                new(sync.Mutex),
			onCronChange:        onCronChange,
			onCommand:           onCommand,
			sendQueueLen:        0,
			getLeader:           getLeader,
			onDispatch:	         onDispatch,
			OnCommandBack:       OnCommandBack,
			codec:               &Codec{},
		}
	c.server = tcp.NewServer(ctx.Context(), ctx.Config.BindAddress, tcp.SetOnServerMessage(c.OnServerMessage))
	c.client = tcp.NewClient(ctx.Context(), tcp.SetOnMessage(c.onClientEvent))
	go c.sendService()
	go c.keep()
	return c
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
					log.Infof("try to send crontab: %+v", d)
					jd := d.encode()//json.Marshal(d)
					sendData, _:= c.codec.Encode(d.Cmd, jd)
					if d.IsMutex {
						sdata := make([]byte, 16)
						binary.LittleEndian.PutUint64(sdata[:8], uint64(d.CronId))
						binary.LittleEndian.PutUint64(sdata[8:], uint64(int64(time.Now().UnixNano()/1000000)))
						c.statisticsStartChan <- sdata
					}
					d.send(d.MsgId, sendData)
					delete(sendQueue, d.Unique)
					fmt.Fprintf(os.Stderr, "send use time %v\n", time.Since(start))
				}
				atomic.StoreInt64(&c.sendQueueLen, int64(len(sendQueue)))
			//case unique, ok := <- c.delSendQueueChan:
			//	if !ok {
			//		return
			//	}
			//	fmt.Fprintf(os.Stderr, "running complete -server %v\r\n", unique)
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
		// client发送pull指令
		// server端收到pull指令之后进行定时任务分发操作
		case node, ok := <-c.onPullChan:
			if !ok {
				return
			}
			if atomic.LoadInt64(&c.sendQueueLen) < 32 {
				if len(mutexKeys) > 0 {
					// 分发互斥任务
					start := time.Now()
					id    := mutexKeys[int(gindexMutex)]
					queueMutex.dispatch(id, func(item *runItem) {
						set, ok := setNum[id]
						if ok {
							set()
						} else {
							log.Errorf("%v set num does not exists", id)
						}
						c.onDispatch(item.Id)
						itemData, _ := item.encode()
						c.sendQueueChan <- newSendData(node.msgId, CMD_RUN_COMMAND, itemData, node.node.Send, item.Id, item.IsMutex, c.ctx.Config.BindAddress)
					})
					fmt.Fprintf(os.Stderr, "dispatch id= %v, OnPullCommand mutex use time %v\n", id, time.Since(start))
					gindexMutex++
					if gindexMutex >= int64(len(mutexKeys)) {
						gindexMutex = 0
					}
				}

				if len(normalKeys) > 0 {
					// 分发普通任务
					start := time.Now()
					id    := normalKeys[int(gindexNormal)]
					queueNomal.dispatch(id, func(item *runItem) {
						set, ok := setNum[id]
						if ok {
							set()
						} else {
							log.Errorf("%v set num does not exists", id)
						}
						c.onDispatch(item.Id)
						itemData, _ := item.encode()
						c.sendQueueChan <- newSendData(node.msgId, CMD_RUN_COMMAND, itemData, node.node.Send, item.Id, item.IsMutex, c.ctx.Config.BindAddress)
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
			setNum[item.Id] = item.SubWaitNum
			if item.IsMutex {
				if _, ok := queueMutex[item.Id]; !ok {
					mutexKeys = append(mutexKeys, item.Id)
				}
				if !queueMutex.append(item) {
					item.SubWaitNum()
				}
			} else {
				if _, ok := queueNomal[item.Id]; !ok {
					normalKeys = append(normalKeys, item.Id)
				}
				if !queueNomal.append(item) {
					item.SubWaitNum()
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
		Id:         id,
		Command:    command,
		IsMutex:    isMutex,
		SubWaitNum: subwaitNum,
	}
	c.dispatch <- item
}

// set on leader select callback
func (c *Controller) OnLeader(isLeader bool) {
	go func() {
		log.Debugf("==============client OnLeader %v===============", isLeader)
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

// start
func (c *Controller) Start() {
	c.server.Start()
}

// close
func (c *Controller) Close() {
	c.server.Close()
}
