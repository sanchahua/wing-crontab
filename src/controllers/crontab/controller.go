package crontab

import (
	"models/cron"
	log "github.com/sirupsen/logrus"
	cronv2 "library/cron"
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
	pullcommand PullCommandFunc
	fixTime int
	runList chan *runItem
	runListLen int64
	pullc chan struct{}
	//waiting int64
	runtimes uint64
	usetime uint64

	onBefore []OnRunFunc
	onAfter  []OnRunFunc
}
type runItem struct {
	id int64
	command string
	//dispatchTime int64
	dispatchServer string
	runServer string
	after func()
	isMutex bool
}

const (
	runListMaxLen = 10000
	uintMax = uint64(1) << 63
)
type PullCommandFunc func()
type OnRunFunc func(id int64, dispatchServer string, runServer string, output []byte, useTime time.Duration)
type OnWillRunFunc func(id int64, command string, isMutex bool, addWaitNum func(), subWaitNum func() int64)
type ControllerOption func(c *CrontabController)

func SetOnWillRun(f OnWillRunFunc) ControllerOption {
	return func(c *CrontabController) {
		c.onwillrun = f
	}
}

func SetPullCommand(f PullCommandFunc) ControllerOption {
	return func(c *CrontabController) {
		c.pullcommand = f
	}
}

func SetOnBefore(f ...OnRunFunc) ControllerOption {
	return func(c *CrontabController) {
		c.onBefore = append(c.onBefore, f...)
	}
}

func SetOnAfter(f ...OnRunFunc) ControllerOption {
	return func(c *CrontabController) {
		c.onAfter = append(c.onAfter, f...)
	}
}

func NewCrontabController(opts ...ControllerOption) *CrontabController {
	cpu := runtime.NumCPU()
	c := &CrontabController{
		handler:     cronv2.New(),
		crontabList: make(map[int64] *CronEntity),
		lock:        new(sync.Mutex),
		running:     0,
		fixTime:     0,
		runList:     make(chan *runItem, runListMaxLen),
		pullc:       make(chan struct{}, cpu * 2 + 2),
		onBefore:    make([]OnRunFunc, 0),
		onAfter:     make([]OnRunFunc, 0),
	}
	for _, f := range opts {
		f(c)
	}
	for i := 0; i < cpu + 2; i++ {
		go c.run()
	}
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

func (c *CrontabController) runCommand(id int64, command string, dispatchServer string, runServer string) {
	for _, f := range c.onBefore {
		f(id, dispatchServer, runServer, []byte(""), 0)
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
	fmt.Fprintf(os.Stderr, "##########################%+v was run: %v##########################\r\n", id, command)
	for _, f := range c.onAfter {
		f(id, dispatchServer, runServer, res, time.Since(start))
	}
}

func (c *CrontabController) run() {
	cpu := runtime.NumCPU() * 2
	var s = make(chan struct{})
	go func(){
		for {
			s <- struct{}{}
			time.Sleep(time.Second)
		}
	}()
	for {
		select {
			case data, ok := <- c.runList:
				if !ok {
					return
				}

				ll := len(c.runList)
				//run one command, pull one
				if len(c.pullc) < cap(c.pullc) && ll < cpu {
					c.pullc <- struct{}{}
				}

				atomic.StoreInt64(&c.runListLen, int64(ll))
				fmt.Fprintf(os.Stderr, "\r\nrun list len %v\r\n", ll)

				// 如果非互斥模式
				// 尽快响应
				if !data.isMutex {
					data.after()
				}

				atomic.AddUint64(&c.runtimes, 1)
				start := uint64(time.Now().UnixNano()/1000000)
				c.runCommand(data.id, data.command , data.dispatchServer , data.runServer)
				v := atomic.AddUint64(&c.usetime, uint64(time.Now().UnixNano()/1000000) - start)

				if v >= uintMax {
					avg      := uint64(0)
					runtimes := uint64(0)
					times    := atomic.LoadUint64(&c.runtimes)

					if times > 0 {
						runtimes = 1
						avg = uint64(atomic.LoadUint64(&c.usetime)/times)
					}

					atomic.StoreUint64(&c.runtimes, runtimes)
					atomic.StoreUint64(&c.usetime, avg)
				}

				// 严格互斥模式下，必须运行完才能响应
				if data.isMutex {
					data.after()
				}
			case <- s :
				atomic.StoreInt64(&c.runListLen, int64(len(c.runList)))
		}
	}
}

func (c *CrontabController) asyncPullCommand() {
	var ch = make(chan struct{})
	cpu := int64(runtime.NumCPU() * 2)
	go func() {
		for {
			ch <- struct{}{}
			time.Sleep(time.Second)
		}
	}()
	for {
		select {
			case _, ok := <- c.pullc:
				if !ok {
					return
				}
				if c.pullcommand != nil {
					//atomic.AddInt64(&c.waiting, 1)
					c.pullcommand()
				}
			case <- ch:
				fmt.Fprintf(os.Stderr, "\r\nrun list len %v\r\n", len(c.runList))
				if atomic.LoadInt64(&c.runListLen) < cpu {
					c.pullc <- struct{}{}
				}
		}
	}
}

func (c *CrontabController) ReceiveCommand(id int64, command string, dispatchServer string, runServer string, isMutex bool, after func()) {
	if len(c.runList) >= cap(c.runList) {
		log.Warnf("runlist len is max then %v", runListMaxLen)
		return
	}
	c.runList <- &runItem{
		id:             id,
		command:        command,
		dispatchServer: dispatchServer,
		runServer:      runServer,
		after:          after,
		isMutex:        isMutex,
	}
}
