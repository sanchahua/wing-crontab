package cron

import (
	"models/cron"
	mlog "models/log"
	log "github.com/cihub/seelog"
	cronv2 "library/cron"
	"sync"
	"fmt"
	"database/sql"
	"errors"
)
const (
	IS_RUNNING = 1
)
type CronController struct {
	cron *cronv2.Cron
	cronList map[int64] *CronEntity
	lock *sync.RWMutex
	status int
	//onwillrun OnWillRunFunc
	//pullcommand PullCommandFunc
	//fixTime int
	runList chan *runItem
	//runListLen int64
	//pullc chan struct{}
	////waiting int64
	//runtimes uint64
	//usetime uint64

	//onBefore []OnRunFunc
	//onAfter  []OnRunFunc
	//db *sql.DB
	logModel *mlog.DbLog
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
	//uintMax = uint64(1) << 63
)
//type PullCommandFunc func()
//type OnRunFunc func(id int64, dispatchServer string, runServer string, output []byte, useTime time.Duration)
//type OnWillRunFunc func(id int64, command string, isMutex bool, addWaitNum func(), subWaitNum func() int64)
//type ControllerOption func(c *CronController)

//func SetOnWillRun(f OnWillRunFunc) ControllerOption {
//	return func(c *CronController) {
//		c.onwillrun = f
//	}
//}
//
//func SetPullCommand(f PullCommandFunc) ControllerOption {
//	return func(c *CronController) {
//		c.pullcommand = f
//	}
//}
//
//func SetOnBefore(f ...OnRunFunc) ControllerOption {
//	return func(c *CronController) {
//		c.onBefore = append(c.onBefore, f...)
//	}
//}
//
//func SetOnAfter(f ...OnRunFunc) ControllerOption {
//	return func(c *CronController) {
//		c.onAfter = append(c.onAfter, f...)
//	}
//}

func NewCronController(db *sql.DB) *CronController {
	//cpu := runtime.NumCPU()
	c := &CronController{
		cron:        cronv2.New(),
		cronList:    make(map[int64] *CronEntity),
		lock:        new(sync.RWMutex),
		status:      0,
		//fixTime:     0,
		runList:     make(chan *runItem, runListMaxLen),
		//pullc:       make(chan struct{}, cpu * 2 + 2),
		//onBefore:    make([]OnRunFunc, 0),
		//onAfter:     make([]OnRunFunc, 0),
		//db:          db,
		logModel: mlog.NewLog(db),
	}
	//for _, f := range opts {
	//	f(c)
	//}
	//for i := 0; i < cpu + 2; i++ {
	//	go c.run()
	//}
	//go c.asyncPullCommand()
	return c
}

func (c *CronController) StartCron() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status & IS_RUNNING > 0 {
		return
	}
	c.status |= IS_RUNNING
	c.cron.Start()
}

func (c *CronController) StopCron() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status & IS_RUNNING <= 0 {
		return
	}
	c.cron.Stop()
}

func (c *CronController) Add(ce *cron.CronEntity) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entity := newCronEntity(ce, c.onRun)
	//func() {
	//	var err error
	//	switch event {
	//	case cron.EVENT_ADD:
	// check if exists
	//e, ok := c.crontabList[entity.Id]
	//if ok {
	//	return
	//}
	//e = newCronEntity(entity, c.onwillrun)
	var err error
	entity.CronId, err = c.cron.AddJob(entity.CronSet, entity)

	if err != nil {
		log.Errorf("%+v", err)
		return entity, err
	}
	c.cronList[entity.Id] = entity //.CronId
	log.Infof("Add success, entity=[%+v]", entity)
	return entity, nil
	//		case cron.EVENT_DELETE:
	//			log.Infof("delete crontab: %+v", entity)
	//			e, ok := c.crontabList[entity.Id]
	//			if ok {
	//				delete(c.crontabList, entity.Id)
	//				c.handler.Remove(e.CronId)
	//			}
	//		case cron.EVENT_START:
	//			log.Infof("start crontab: %+v", entity)
	//			e, ok := c.crontabList[entity.Id]
	//			if ok {
	//				e.Stop = false
	//			}
	//
	//		case cron.EVENT_STOP:
	//			log.Infof("stop crontab: %+v", entity)
	//			e, ok := c.crontabList[entity.Id]
	//			if ok {
	//				e.Stop = true
	//			}
	//
	//		case cron.EVENT_UPDATE:
	//			log.Infof("update crontab: %+v", entity)
	//			e, ok := c.crontabList[entity.Id]
	//			if ok {
	//
	//				c.handler.Remove(e.CronId)
	//
	//				e.CronSet   = entity.CronSet
	//				e.Command   = entity.Command
	//				e.Stop      = entity.Stop
	//				e.Remark    = entity.Remark
	//				e.StartTime = entity.StartTime
	//				e.EndTime   = entity.EndTime
	//				e.IsMutex   = entity.IsMutex
	//
	//				e.CronId, err = c.handler.AddJob(entity.CronSet, e)
	//				if err != nil {
	//					log.Errorf("%+v", err)
	//				}
	//				c.crontabList[entity.Id] = e
	//			}
	//		}
	//	}()
	//	c.lock.Unlock()
	//	c.Start()
	//}
}

func (c *CronController) Delete(id int64) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	delete(c.cronList, id)
	c.cron.Remove(e.CronId)
	e.delete()
	return e, nil
}

func (c *CronController) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}

	//			if ok {
	//
	//				c.handler.Remove(e.CronId)
	//
					e.CronSet   = cronSet
					e.Command   = command
					e.Stop      = stop
					e.Remark    = remark
					e.StartTime = startTime
					e.EndTime   = endTime
					e.IsMutex   = isMutex
	//
	//				e.CronId, err = c.handler.AddJob(entity.CronSet, e)
	//				if err != nil {
	//					log.Errorf("%+v", err)
	//				}
	//				c.crontabList[entity.Id] = e
	//			}
	return e, nil
}

func (c *CronController) Get(id int64) (*CronEntity, error)  {
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	return e, nil
}

func (c *CronController) GetList() (map[int64]*CronEntity, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.cronList, nil
}

func (c *CronController) Stop(id int64, stop bool) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}

	//			if ok {
	//
	//				c.handler.Remove(e.CronId)
	//
	e.Stop      = stop
	//
	//				e.CronId, err = c.handler.AddJob(entity.CronSet, e)
	//				if err != nil {
	//					log.Errorf("%+v", err)
	//				}
	//				c.crontabList[entity.Id] = e
	//			}
	return e, nil
}

func (c *CronController) onRun(cron_id int64, output string, usetime int64, remark string) {
	_, err := c.logModel.Add(cron_id, output, usetime, remark)
	if err != nil {
		log.Errorf("onRun c.logModel.Add fail, cron_id=[%v], output=[%v], usetime=[%v], remark=[%v], error=[%v]", cron_id, output, usetime, remark, err)
	}
}

//func (c *CronController) run() {
//	cpu := runtime.NumCPU() * 2
//	var s = make(chan struct{})
//	go func(){
//		for {
//			s <- struct{}{}
//			time.Sleep(time.Second)
//		}
//	}()
//	for {
//		select {
//			case data, ok := <- c.runList:
//				if !ok {
//					return
//				}
//
//				ll := len(c.runList)
//				//run one command, pull one
//				if len(c.pullc) < cap(c.pullc) && ll < cpu {
//					c.pullc <- struct{}{}
//				}
//
//				atomic.StoreInt64(&c.runListLen, int64(ll))
//				fmt.Fprintf(os.Stderr, "\r\nrun list len %v\r\n", ll)
//
//				// 如果非互斥模式
//				// 尽快响应
//				if !data.isMutex {
//					data.after()
//				}
//
//				atomic.AddUint64(&c.runtimes, 1)
//				start := uint64(time.Now().UnixNano()/1000000)
//				c.runCommand(data.id, data.command , data.dispatchServer , data.runServer)
//				v := atomic.AddUint64(&c.usetime, uint64(time.Now().UnixNano()/1000000) - start)
//
//				if v >= uintMax {
//					avg      := uint64(0)
//					runtimes := uint64(0)
//					times    := atomic.LoadUint64(&c.runtimes)
//
//					if times > 0 {
//						runtimes = 1
//						avg = uint64(atomic.LoadUint64(&c.usetime)/times)
//					}
//
//					atomic.StoreUint64(&c.runtimes, runtimes)
//					atomic.StoreUint64(&c.usetime, avg)
//				}
//
//				// 严格互斥模式下，必须运行完才能响应
//				if data.isMutex {
//					data.after()
//				}
//			case <- s :
//				atomic.StoreInt64(&c.runListLen, int64(len(c.runList)))
//		}
//	}
//}

//func (c *CronController) asyncPullCommand() {
//	var ch = make(chan struct{})
//	cpu := int64(runtime.NumCPU() * 2)
//	go func() {
//		for {
//			ch <- struct{}{}
//			time.Sleep(time.Second)
//		}
//	}()
//	for {
//		select {
//			case _, ok := <- c.pullc:
//				if !ok {
//					return
//				}
//				if c.pullcommand != nil {
//					//atomic.AddInt64(&c.waiting, 1)
//					c.pullcommand()
//				}
//			case <- ch:
//				fmt.Fprintf(os.Stderr, "\r\nrun list len %v\r\n", len(c.runList))
//				if atomic.LoadInt64(&c.runListLen) < cpu {
//					c.pullc <- struct{}{}
//				}
//		}
//	}
//}

//func (c *CronController) ReceiveCommand(id int64, command string, dispatchServer string, runServer string, isMutex bool, after func()) {
//	if len(c.runList) >= cap(c.runList) {
//		log.Warnf("runlist len is max then %v", runListMaxLen)
//		return
//	}
//	c.runList <- &runItem{
//		id:             id,
//		command:        command,
//		dispatchServer: dispatchServer,
//		runServer:      runServer,
//		after:          after,
//		isMutex:        isMutex,
//	}
//}
