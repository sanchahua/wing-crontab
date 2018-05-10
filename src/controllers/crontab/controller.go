package crontab

import (
	"models/cron"
	log "github.com/sirupsen/logrus"
	cronv2 "gopkg.in/robfig/cron.v2"
	"os/exec"
	"sync"
	"time"
)

type CrontabController struct {
	handler *cronv2.Cron
	crontabList map[int64] *CronEntity//cronv2.EntryID
	lock *sync.Mutex
	running bool
	onwillrun OnWillRunFunc
	onrun OnRunFunc
}
type OnRunFunc func(id int64, runServer string, output []byte, useTime time.Duration)
type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	CronId cronv2.EntryID    `json:"cron_id"`//runtime cron id
	onwillrun OnWillRunFunc `json:"-"`
}

func (row *CronEntity) Run() {
	if row.Stop {
		// 外部注入，停止执行定时任务支持
		log.Debugf("%+v was stop", row.Id)
		return
	}

	//roundbin to target server and run command
	log.Infof("will run: %+v", *row)
	row.onwillrun(row.Id, row.Command)
}

type OnWillRunFunc func(id int64, command string)
type CrontabControllerOption func(c *CrontabController)
func SetOnWillRun(f OnWillRunFunc) CrontabControllerOption {
	return func(c *CrontabController) {
		c.onwillrun = f
	}
}

func SetOnRun(f OnRunFunc) CrontabControllerOption {
	return func(c *CrontabController) {
		log.Debugf("set c.onrun")
		c.onrun = f
	}
}

func NewCrontabController(opts ...CrontabControllerOption) *CrontabController {
	c := &CrontabController{
		handler: cronv2.New(),
		crontabList:make(map[int64] *CronEntity),//cronv2.EntryID),
		lock:new(sync.Mutex),
		running:false,
	}
	for _, f := range opts {
		f(c)
	}
	return c
}
func (c *CrontabController) Start() {
	c.lock.Lock()
	c.running = true
	c.lock.Unlock()
	c.handler.Start()
}
func (c *CrontabController) Stop() {
	c.lock.Lock()
	c.running = false
	c.lock.Unlock()
	c.handler.Stop()
}

func (c *CrontabController) OnCrontabChange(event int, entity *cron.CronEntity) {
	var err error
	switch event {
	case cron.EVENT_ADD:
		log.Infof("add crontab: %+v", entity)

		// check if exists
		c.lock.Lock()
		e, ok := c.crontabList[entity.Id]
		if ok {
			c.lock.Unlock()
			return
		} else {
			c.lock.Unlock()
			e = &CronEntity{
				Id :entity.Id,//int64        `json:"id"`
				CronSet:entity.CronSet,// string  `json:"cron_set"`
				Command:entity.Command,// string  `json:"command"`
				Remark :entity.Remark,//string   `json:"remark"`
				Stop :entity.Stop,//bool       `json:"stop"`
				CronId :0,//int64    `json:"cron_id"`
				onwillrun:c.onwillrun,
			}
		}




		c.lock.Lock()
		if c.running {
			c.lock.Unlock()
			c.Stop()
		} else {
			c.lock.Unlock()
		}

		e.CronId, err = c.handler.AddJob(entity.CronSet, e)

		c.lock.Lock()
		if !c.running {
			c.lock.Unlock()
			c.Start()
		} else {
			c.lock.Unlock()
		}

		if err != nil {
			log.Errorf("%+v", err)
		} else {
			c.lock.Lock()
			c.crontabList[e.Id] = e//.CronId
			c.lock.Unlock()
		}
	case cron.EVENT_DELETE:
		log.Infof("delete crontab: %+v", entity)
		c.lock.Lock()
		e, ok := c.crontabList[entity.Id]
		if ok {
			delete(c.crontabList, entity.Id)
			c.handler.Remove(e.CronId)
		}
		c.lock.Unlock()
	case cron.EVENT_START:
		log.Infof("start crontab: %+v", entity)
		c.lock.Lock()
		e, ok := c.crontabList[entity.Id]
		if ok {
			e.Stop = false
		}
		c.lock.Unlock()

	case cron.EVENT_STOP:
		log.Infof("stop crontab: %+v", entity)
		c.lock.Lock()
		e, ok := c.crontabList[entity.Id]
		if ok {
			e.Stop = true
		}
		c.lock.Unlock()

	case cron.EVENT_UPDATE:
		log.Infof("update crontab: %+v", entity)
		c.lock.Lock()
		e, ok := c.crontabList[entity.Id]
		c.lock.Unlock()
		if ok {

			c.lock.Lock()
			if c.running {
				c.lock.Unlock()
				c.Stop()
			} else {
				c.lock.Unlock()
			}

			c.handler.Remove(e.CronId)

			e.CronSet     = entity.CronSet
			e.Command     = entity.Command
			e.Stop        = entity.Stop
			e.Remark      = entity.Remark
			e.CronId, err = c.handler.AddJob(entity.CronSet, e)
			if err != nil {
				log.Errorf("%+v", err)
			}

			c.lock.Lock()
			if !c.running {
				c.lock.Unlock()
				c.Start()
			} else {
				c.lock.Unlock()
			}
		}
	}
}

func (c *CrontabController) RunCommand(id int64, command string, runServer string) {
	var cmd *exec.Cmd
	var err error
	start := time.Now()
	cmd = exec.Command("bash", "-c", command)
	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("执行命令发%s生错误：%+v", command, err)
	}
	log.Debugf("%+v:%v was run", id, command)
	if c.onrun == nil {
		log.Errorf("c.onrun is nil")
		return
	}
	c.onrun(id, runServer, res, time.Since(start))
}
