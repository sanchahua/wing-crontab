package cron

import (
	"models/cron"
	mlog "models/log"
	//log "github.com/cihub/seelog"
	log "gitlab.xunlei.cn/xllive/common/log"
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
	cron     *cronv2.Cron
	cronList map[int64] *CronEntity
	lock     *sync.RWMutex
	status   int
	logModel *mlog.DbLog
}

func NewCronController(db *sql.DB) *CronController {
	c := &CronController{
		cron:     cronv2.New(),
		cronList: make(map[int64] *CronEntity),
		lock:     new(sync.RWMutex),
		status:   0,
		logModel: mlog.NewLog(db),
	}
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
	var err error
	entity.CronId, err = c.cron.AddJob(entity.CronSet, entity)
	if err != nil {
		log.Errorf("%+v", err)
		return entity, err
	}
	c.cronList[entity.Id] = entity
	log.Tracef("Add success, entity=[%+v]", entity)
	return entity, nil
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
	e.CronSet   = cronSet
	e.Command   = command
	e.Stop      = stop
	e.Remark    = remark
	e.StartTime = startTime
	e.EndTime   = endTime
	e.IsMutex   = isMutex
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
	e.Stop = stop
	return e, nil
}

func (c *CronController) onRun(cron_id int64, output string, usetime int64, remark string) {
	_, err := c.logModel.Add(cron_id, output, usetime, remark)
	if err != nil {
		log.Errorf("onRun c.logModel.Add fail, cron_id=[%v], output=[%v], usetime=[%v], remark=[%v], error=[%v]", cron_id, output, usetime, remark, err)
	}
}