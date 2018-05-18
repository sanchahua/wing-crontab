package crontab

import (
	"models/cron"
	log "github.com/sirupsen/logrus"
	cronv2 "gopkg.in/robfig/cron.v2"
	"os/exec"
	"sync"
	"time"
	"runtime"
)

type CrontabController struct {
	handler *cronv2.Cron
	crontabList map[int64] *CronEntity//cronv2.EntryID
	lock *sync.Mutex
	running bool
	onwillrun OnWillRunFunc
	onrun OnRunFunc
	pullcommand PullCommandFunc
	fixTime int
	runList chan *runItem
	pullc chan struct{}
	times int64
}
const runListMaxLen = 10000
type runItem struct {
	id int64
	command string
	dispatchTime int64
	dispatchServer string
	runServer string
}
type PullCommandFunc func()
type OnRunFunc func(id int64, dispatchTime int64, dispatchServer string, runServer string, output []byte, useTime time.Duration)
type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	CronId cronv2.EntryID    `json:"cron_id"`//runtime cron id
	onwillrun OnWillRunFunc `json:"-"`
	StartTime int64 `json:"start_time"`
	EndTime int64 `json:"end_time"`
	IsMutex bool  `json:"is_mutex"`
}

type IFilter interface {
	Check() bool
}
type CronEntityMiddleWare func(entity *CronEntity) IFilter
type StopFilter struct {
	row *CronEntity
}
func StopMiddleware() CronEntityMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &StopFilter{entity}
	}
}

func (f *StopFilter) Check() bool {
	return f.row.Stop
}

type TimeFilter struct {
	row *CronEntity
	next IFilter
}
func TimeMiddleware(next IFilter) CronEntityMiddleWare {
	return func(entity *CronEntity) IFilter {
		return &TimeFilter{row:entity, next:next}
	}
}

func (f *TimeFilter) Check() bool {
	if f.next.Check() {
		return true
	}

	if f.row.EndTime <= 0 {
		return false
	}

	current := time.Now().Unix()
	if current >= f.row.StartTime && current < f.row.EndTime {
		return false
	}
	return true
}


func (row *CronEntity) Run() {
	//start := time.Now()

	m := StopMiddleware()(row)
	m  = TimeMiddleware(m)(row)
	if m.Check() {
		// 外部注入，停止执行定时任务支持
		log.Debugf("%+v was stop", row.Id)
		return
	}

	//roundbin to target server and run command
	row.onwillrun(row.Id, row.Command, row.IsMutex)
	//log.Infof("will run: %+v", *row)
}

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
		//log.Debugf("set c.onrun")
		c.onrun = f
	}
}


const (
	minFixTime = 0
	maxFixTime = 60
)

func NewCrontabController(opts ...CrontabControllerOption) *CrontabController {
	cpu := runtime.NumCPU()

	c := &CrontabController{
		handler: cronv2.New(),
		crontabList:make(map[int64] *CronEntity),//cronv2.EntryID),
		lock:new(sync.Mutex),
		running:false,
		fixTime:0,
		runList:make(chan *runItem, runListMaxLen),
		pullc:make(chan struct{}, cpu * 2),
		times:0,
	}
	for _, f := range opts {
		f(c)
	}

	for i := 0; i < cpu + 2; i++ {
		go c.run()
	}
	go c.pullCommand()
	go c.asyncPullCommand()
	return c
}

func (c *CrontabController) Start() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.running {
		return
	}
	c.running = true
	c.handler.Start()
}

func (c *CrontabController) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.running {
		return
	}
	c.running = false
	c.handler.Stop()
}

func (c *CrontabController) Add(event int, entity *cron.CronEntity) {
	c.Stop()
	defer c.Start()
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	switch event {
	case cron.EVENT_ADD:
		log.Infof("add crontab: %+v", entity)

		// check if exists

		e, ok := c.crontabList[entity.Id]
		if ok {
			return
		} else {
			e = &CronEntity{
				Id :entity.Id,//int64        `json:"id"`
				CronSet:entity.CronSet,// string  `json:"cron_set"`
				Command:entity.Command,// string  `json:"command"`
				Remark :entity.Remark,//string   `json:"remark"`
				Stop :entity.Stop,//bool       `json:"stop"`
				CronId :0,//int64    `json:"cron_id"`
				onwillrun:c.onwillrun,
				StartTime:entity.StartTime,
				EndTime:entity.EndTime,
				IsMutex:entity.IsMutex,
			}
		}

		e.CronId, err = c.handler.AddJob(entity.CronSet, e)

		if err != nil {
			log.Errorf("%+v", err)
		} else {
			c.crontabList[e.Id] = e//.CronId
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

			e.CronSet     = entity.CronSet
			e.Command     = entity.Command
			e.Stop        = entity.Stop
			e.Remark      = entity.Remark
			e.StartTime   = entity.StartTime
			e.EndTime     = entity.EndTime
			e.IsMutex = entity.IsMutex

			e.CronId, err = c.handler.AddJob(entity.CronSet, e)
			if err != nil {
				log.Errorf("%+v", err)
			}
			c.crontabList[entity.Id] = e
		}
	}
}

func (c *CrontabController) runCommand(data *runItem) {
	f := int(time.Now().Unix() - data.dispatchTime)
	//if f > minFixTime && f <= maxFixTime && f > c.fixTime {
	//	c.fixTime = f
	//}
	if f > minFixTime {
		//log.Warnf("diff time %v max then %v", f, minFixTime)
	}
	//log.Debugf("#######current fix time %v>%v", c.fixTime, f)
	//if c.fixTime > 0 {
	//	time.Sleep(time.Second * time.Duration(c.fixTime))
	//}
	var cmd *exec.Cmd
	var err error
	start := time.Now()
	cmd = exec.Command("bash", "-c", data.command)
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("执行命令(%v)发生错误：%+v", data.command, err)
	}
	log.Debugf("#################################%+v:%v was run", data.id, data.command)
	if c.onrun == nil {
		log.Errorf("c.onrun is nil")
		return
	}
	c.onrun(data.id, data.dispatchTime, data.dispatchServer, data.runServer, res, time.Since(start))
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
				c.runCommand(data)
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
					c.pullcommand()
				}
		}
	}
}

func (c *CrontabController) pullCommand() {
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
			log.Warnf("runlist len is min then %v < %v", len(c.runList), cpu)
			if len(c.pullc) < cap(c.pullc) {
				c.pullc <- struct{}{}
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (c *CrontabController) ReceiveCommand(id int64, command string, dispatchTime int64, dispatchServer string, runServer string) {
	if len(c.runList) >= runListMaxLen {
		log.Errorf("runlist len is max then %v", runListMaxLen)
		return
	}
	c.runList <- &runItem{
		id:id,
		command:command,
		dispatchTime:dispatchTime,
		dispatchServer:dispatchServer,
		runServer:runServer,
	}
	//c.times--
	log.Debugf("ReceiveCommand (%v) %v, %v, %v, %v, %v ", len(c.runList), id, command, dispatchTime, dispatchServer, runServer)


	//return
	//go func() {
	//	f := int(time.Now().Unix() - dispatchTime)
	//	//if f > minFixTime && f <= maxFixTime && f > c.fixTime {
	//	//	c.fixTime = f
	//	//}
	//	if f > minFixTime {
	//		log.Warnf("diff time %v max then %v", f, minFixTime)
	//	}
	//	//log.Debugf("#######current fix time %v>%v", c.fixTime, f)
	//	//if c.fixTime > 0 {
	//	//	time.Sleep(time.Second * time.Duration(c.fixTime))
	//	//}
	//	var cmd *exec.Cmd
	//	var err error
	//	start := time.Now()
	//	cmd = exec.Command("bash", "-c", command)
	//	res, err := cmd.CombinedOutput()
	//	if err != nil {
	//		log.Errorf("执行命令(%v)发生错误：%+v", command, err)
	//	}
	//	log.Debugf("%+v:%v was run", id, command)
	//	if c.onrun == nil {
	//		log.Errorf("c.onrun is nil")
	//		return
	//	}
	//	c.onrun(id, dispatchTime, dispatchServer, runServer, res, time.Since(start))
	//}()
}
