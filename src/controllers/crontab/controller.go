package crontab

import (
	"models/cron"
	log "github.com/sirupsen/logrus"
	cronv2 "gopkg.in/robfig/cron.v2"
	"os/exec"
	"sync"
	"time"
	"runtime"
	"sync/atomic"
	"fmt"
	"os"
)

type CrontabController struct {
	handler *cronv2.Cron
	crontabList map[int64] *CronEntity
	lock *sync.Mutex
	running int64
	onwillrun OnWillRunFunc
	onrun OnRunFunc
	pullcommand PullCommandFunc
	fixTime int
	runList chan *runItem
	pullc chan struct{}
}
type runItem struct {
	id int64
	command string
	dispatchTime int64
	dispatchServer string
	runServer string
	after func()
	isMutex bool
	logId int64
}

const (
	minFixTime = 0
	maxFixTime = 60
	runListMaxLen = 10000
)
type PullCommandFunc func()
type OnRunFunc func(id int64, dispatchTime int64, dispatchServer string, runServer string, output []byte, useTime time.Duration, logId int64)
type OnWillRunFunc func(id int64, command string, isMutex bool)
type CrontabControllerOption func(c *CrontabController)

func SetOnWillRun(f OnWillRunFunc) CrontabControllerOption {
	return func(c *CrontabController) {
		c.onwillrun = f
	}
}

func SetPullCommand(f PullCommandFunc) CrontabControllerOption {
	return func(c *CrontabController) {
		c.pullcommand = f
	}
}

func SetOnRun(f OnRunFunc) CrontabControllerOption {
	return func(c *CrontabController) {
		c.onrun = f
	}
}

func NewCrontabController(opts ...CrontabControllerOption) *CrontabController {
	cpu := runtime.NumCPU()

	c := &CrontabController{
		handler: cronv2.New(),
		crontabList:make(map[int64] *CronEntity),
		lock:new(sync.Mutex),
		running:0,
		fixTime:0,
		runList:make(chan *runItem, runListMaxLen),
		pullc:make(chan struct{}, cpu * 2),
	}
	for _, f := range opts {
		f(c)
	}

	log.Debugf("cpu num %v", cpu)
	for i := 0; i < cpu + 2; i++ {
		go c.run()
	}
	go c.checkCommandLen()
	go c.asyncPullCommand()
	return c
}

func (c *CrontabController) Start() {
	c.lock.Lock()
	if atomic.LoadInt64(&c.running) == 1 {
		c.lock.Unlock()
		return
	}
	atomic.StoreInt64(&c.running, 1)
	c.handler.Start()
	c.lock.Unlock()
}

func (c *CrontabController) Stop() {
	c.lock.Lock()
	if atomic.LoadInt64(&c.running) == 0 {
		c.lock.Unlock()
		return
	}
	atomic.StoreInt64(&c.running, 0)
	c.handler.Stop()
	c.lock.Unlock()
}

func (c *CrontabController) Add(event int, entity *cron.CronEntity) {
	c.Stop()
	c.lock.Lock()
	func() {
		var err error
		switch event {
		case cron.EVENT_ADD:
			log.Infof("add crontab: %+v", entity)
			// check if exists
			e, ok := c.crontabList[entity.Id]
			if ok {
				return
			}
			e = newCronEntity(entity, c.onwillrun)
			//&CronEntity{
			//	Id:        entity.Id,      //int64        `json:"id"`
			//	CronSet:   entity.CronSet, // string  `json:"cron_set"`
			//	Command:   entity.Command, // string  `json:"command"`
			//	Remark:    entity.Remark,  //string   `json:"remark"`
			//	Stop:      entity.Stop,    //bool       `json:"stop"`
			//	CronId:    0,              //int64    `json:"cron_id"`
			//	onwillrun: c.onwillrun,
			//	StartTime: entity.StartTime,
			//	EndTime:   entity.EndTime,
			//	IsMutex:   entity.IsMutex,
			//}

			e.CronId, err = c.handler.AddJob(entity.CronSet, e)

			if err != nil {
				log.Errorf("%+v", err)
			} else {
				c.crontabList[e.Id] = e //.CronId
			}
		case cron.EVENT_DELETE:
			log.Infof("delete crontab: %+v", entity)
			e, ok := c.crontabList[entity.Id]
			if ok {
				delete(c.crontabList, entity.Id)
				c.handler.Remove(e.CronId)
			}
		case cron.EVENT_START:
			log.Infof("start crontab: %+v", entity)
			e, ok := c.crontabList[entity.Id]
			if ok {
				e.Stop = false
			}

		case cron.EVENT_STOP:
			log.Infof("stop crontab: %+v", entity)
			e, ok := c.crontabList[entity.Id]
			if ok {
				e.Stop = true
			}

		case cron.EVENT_UPDATE:
			log.Infof("update crontab: %+v", entity)
			e, ok := c.crontabList[entity.Id]
			if ok {

				c.handler.Remove(e.CronId)

				e.CronSet   = entity.CronSet
				e.Command   = entity.Command
				e.Stop      = entity.Stop
				e.Remark    = entity.Remark
				e.StartTime = entity.StartTime
				e.EndTime   = entity.EndTime
				e.IsMutex   = entity.IsMutex

				e.CronId, err = c.handler.AddJob(entity.CronSet, e)
				if err != nil {
					log.Errorf("%+v", err)
				}
				c.crontabList[entity.Id] = e
			}
		}
	}()
	c.lock.Unlock()
	c.Start()
}

func (c *CrontabController) runCommand(id int64, command string, dispatchTime int64, dispatchServer string, runServer string, logId int64) {
	f := int(time.Now().Unix() - dispatchTime)
	if f > minFixTime {
		//log.Warnf("diff time %v max then %v", f, minFixTime)
	}
	var cmd *exec.Cmd
	var err error
	start := time.Now()
	cmd = exec.Command("bash", "-c", command)
	res, err := cmd.CombinedOutput()
	if err != nil {
		res = append(res, []byte("  error: " + err.Error())...)
		log.Errorf("执行命令(%v)发生错误：%+v", command, err)
	}
	log.Infof("%+v was run: %v", id, command)
	if c.onrun == nil {
		log.Warnf("c.onrun is nil")
		return
	}
	c.onrun(id, dispatchTime, dispatchServer, runServer, res, time.Since(start), logId)
}

func (c *CrontabController) run() {
	cpu := runtime.NumCPU() * 2
	for {
		select {
			case data, ok := <- c.runList:
				if !ok {
					return
				}
				//run one command, pull one
				if len(c.pullc) < cap(c.pullc) && len(c.runList) < cpu {
					c.pullc <- struct{}{}
				}
				// 如果非互斥模式
				// 尽快响应
				if !data.isMutex {
					data.after()
				}
				c.runCommand(data.id, data.command , data.dispatchTime , data.dispatchServer , data.runServer, data.logId)
				// 严格互斥模式下，必须运行完才能响应
				if data.isMutex {
					data.after()
				}
		}
	}
}

func (c *CrontabController) asyncPullCommand() {
	for {
		select {
			case _, ok := <- c.pullc:
				if !ok {
					return
				}
				if c.pullcommand != nil {
					fmt.Fprintf(os.Stderr, "send pull\r\n")
					c.pullcommand()
				}
		}
	}
}

func (c *CrontabController) checkCommandLen() {
	for {
		if c.pullcommand == nil {
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	cpu := runtime.NumCPU()
	for {
		if len(c.pullc) < cap(c.pullc) {
			c.pullc <- struct{}{}
		}
		if len(c.runList) >= cpu * 2 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	// just for check error
	for {
		if len(c.runList) < cpu {
			fmt.Fprintf(os.Stderr, "warning: runlist len is min then %v < %v\r\n", len(c.runList), cpu)
			//log.Warnf("runlist len is min then %v < %v", len(c.runList), cpu)
			if len(c.pullc) < cap(c.pullc) {
				c.pullc <- struct{}{}
			}
		} else {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		time.Sleep(time.Millisecond * 1)
	}
}

func (c *CrontabController) ReceiveCommand(id int64, command string, dispatchTime int64, dispatchServer string, runServer string, isMutex byte, logId int64, after func()) {
	if len(c.runList) >= runListMaxLen {
		log.Errorf("runlist len is max then %v", runListMaxLen)
		return
	}
	// 如果指定异步执行
	//if isMutex != 1 {
		//after()
		c.runList <- &runItem{
			id:             id,
			command:        command,
			dispatchTime:   dispatchTime,
			dispatchServer: dispatchServer,
			runServer:      runServer,
			after:          after,
			isMutex:        isMutex == 1,
			logId:          logId,
		}
	//} else {
	//	//同步执行
	//	c.runCommand(id, command , dispatchTime , dispatchServer , runServer)
	//	after()
	//}
}
