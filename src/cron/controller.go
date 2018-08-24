package cron

import (
	"models/cron"
	modelLog "models/log"
	"gitlab.xunlei.cn/xllive/common/log"
	cronV2 "library/cron"
	"sync"
	"fmt"
	"errors"
)
const (
	IsRunning = 1
)
type Controller struct {
	cron     *cronV2.Cron
	cronList map[int64] *CronEntity
	lock     *sync.RWMutex
	status   int
	logModel *modelLog.DbLog
	cache    []*CronEntity
}

func NewController(logModel *modelLog.DbLog) *Controller {
	c := &Controller{
		cron:     cronV2.New(),
		cronList: make(map[int64] *CronEntity),
		lock:     new(sync.RWMutex),
		status:   0,
		logModel: logModel,
		cache:    nil,//make([]*CronEntity, 0),
	}
	return c
}

func (c *Controller) StartCron() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status & IsRunning > 0 {
		return
	}
	c.status |= IsRunning
	c.cron.Start()
}

func (c *Controller) StopCron() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status & IsRunning <= 0 {
		return
	}
	c.cron.Stop()
}

func (c *Controller) Add(ce *cron.CronEntity) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	entity := newCronEntity(ce, c.onRun)
	var err error
	entity.CronId, err = c.cron.AddJob(entity.CronSet, entity)
	if err != nil {
		log.Errorf("%+v", err)
		return entity, err
	}
	c.cronList[entity.Id] = entity
	debugStr, _:= entity.toJson()
	log.Tracef("Add success, entity=[%s]", debugStr)
	return entity, nil
}

func (c *Controller) Delete(id int64) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	delete(c.cronList, id)
	c.cron.Remove(e.CronId)
	return e, nil
}

func (c *Controller) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) (*CronEntity, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e.Update(cronSet, command, remark, stop, startTime, endTime, isMutex)
	return e, nil
}

func (c *Controller) Get(id int64) (*CronEntity, error)  {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	return e, nil
}

// 已处理线程安全问题
// 所有内容使用只读cache
func (c *Controller) GetList() []*CronEntity {
	c.lock.RLock()
	defer c.lock.RUnlock()
	l := len(c.cronList)
	if c.cache == nil || len(c.cache) != l {
		c.cache = make([]*CronEntity, l)
	}
	i := 0
	for _, v := range c.cronList {
		c.cache[i] = v.clone()
		i++
	}
	return c.cache
}

func (c *Controller) Stop(id int64, stop bool) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e.setStop(stop)
	return nil
}

func (c *Controller) onRun(cronId int64, output string, useTime int64, remark, startTime string) {
	_, err := c.logModel.Add(cronId, output, useTime, remark, startTime)
	if err != nil {
		log.Errorf("onRun c.logModel.Add fail, cron_id=[%v], output=[%v], usetime=[%v], remark=[%v], startTime=[%v], error=[%v]", cronId, output, useTime, remark, startTime, err)
	}
}