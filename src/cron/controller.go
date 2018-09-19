package cron

import (
	"models/cron"
	modelLog "models/log"
	"gitlab.xunlei.cn/xllive/common/log"
	cronV2 "library/cron"
	"sync"
	"fmt"
	"errors"
	"sort"
	"models/statistics"
	"time"
	"os"
	"github.com/go-redis/redis"
)
const (
	IsRunning = 1
	StateStart = "start"
	StateSuccess = "success"
	 StateFail = "fail"
)
type Controller struct {
	serviceId int64
	cron     *cronV2.Cron
	cronList map[int64] *CronEntity
	lock     *sync.RWMutex
	status   int
	logModel *modelLog.DbLog
	cache    ListCronEntity//[]*CronEntity
	statisticsModel *statistics.Statistics
	Leader   bool
	redis *redis.Client
	RedisKeyPrex string
}

func NewController(redis *redis.Client, RedisKeyPrex string, logModel *modelLog.DbLog, statisticsModel *statistics.Statistics) *Controller {
	c := &Controller{
		serviceId: 0,
		cron:     cronV2.New(),
		cronList: make(map[int64] *CronEntity),
		lock:     new(sync.RWMutex),
		status:   0,
		logModel: logModel,
		cache:    nil,//make([]*CronEntity, 0),
		statisticsModel: statisticsModel,
		Leader:   false,
		redis: redis,
		RedisKeyPrex: RedisKeyPrex,
	}
	return c
}

func (c *Controller) SetLeader(isLeader bool) {
	c.lock.Lock()
	c.Leader = isLeader
	for _, v := range c.cronList {
		v.SetLeader(isLeader)
	}
	c.lock.Unlock()
}

func (c *Controller) SetServiceId(serviceId int64) {
	c.lock.Lock()
	c.serviceId = serviceId
	for _, v := range c.cronList {
		v.SetServiceId(serviceId)
	}
	c.lock.Unlock()
}

func (c *Controller) StartCron() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status & IsRunning > 0 {
		return
	}
	c.status |= IsRunning
	fmt.Fprintf(os.Stderr,"%v", "start run\r\n")
	c.cron.Start()
}

func (c *Controller) StopCron() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status & IsRunning <= 0 {
		return
	}
	c.status ^= IsRunning
	fmt.Fprintf(os.Stderr,"%v", "stop run\r\n")
	c.cron.Stop()
}

func (c *Controller) RestartCron() {
	c.StopCron()
	time.Sleep(1 * time.Second)
	c.StartCron()
}

func (c *Controller) Add(ce *cron.CronEntity) (*CronEntity, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	entity := newCronEntity(c.redis, c.RedisKeyPrex, ce, c.onRun)
	var err error
	entity.CronId, err = c.cron.AddJob(entity.CronSet, entity)
	if err != nil {
		log.Errorf("%+v", err)
		return entity, err
	}
	c.cronList[entity.Id] = entity
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
	e.Exit()
	c.cron.Remove(e.CronId)
	return e, nil
}

func (c *Controller) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime string, isMutex bool) (*CronEntity, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}

	delete(c.cronList, id)
	e.Exit()
	c.cron.Remove(e.CronId)

	//e.Update(cronSet, command, remark, stop, startTime, endTime, isMutex)
	entity := newCronEntity(c.redis, c.RedisKeyPrex, &cron.CronEntity{
		Id:        id,// int64        `json:"id"`
		CronSet:   cronSet,// string  `json:"cron_set"`
		Command:   command,// string  `json:"command"`
		Remark:    remark,// string   `json:"remark"`
		Stop:      stop,// bool       `json:"stop"`
		StartTime: startTime,// int64 `json:"start_time"`
		EndTime:   endTime,// int64   `json:"end_time"`
		IsMutex:   isMutex,// bool    `json:"is_mutex"`
	}, c.onRun)
	entity.CronId, _ = c.cron.AddJob(entity.CronSet, entity)
	c.cronList[entity.Id] = entity
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

// timeout 超时，单位秒
func (c *Controller) RunCommand(id int64, timeout int64) ([]byte, int, error)  {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return nil, 0, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	if timeout < 1 {
		timeout = 3
	}
	return e.runCommandWithTimeout(time.Duration(timeout) * time.Second)
}

func (c *Controller) Kill(id int64, processId int) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return// nil, 0, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e.Kill(processId)//(time.Duration(timeout) * time.Second)
}

func (c *Controller) ProcessIsRunning(id int64, processId int) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return false// nil, 0, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	return e.ProcessIsRunning(processId)//(time.Duration(timeout) * time.Second)
}

// 已处理线程安全问题
// 所有内容使用只读cache
func (c *Controller) GetList() ListCronEntity {
	c.lock.RLock()
	defer c.lock.RUnlock()
	l := len(c.cronList)
	if c.cache == nil || len(c.cache) != l {
		c.cache = make(ListCronEntity, l)
	}
	i := 0
	for _, v := range c.cronList {
		c.cache[i] = v.Clone()
		i++
	}
	sort.Sort(c.cache)
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

func (c *Controller) Mutex(id int64, mutex bool) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cronList[id]
	if !ok {
		return errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e.setMutex(mutex)
	return nil
}

func (c *Controller) onRun(dispatchServer, runServer int64, cronId int64, processId int, state, output string, useTime int64, remark, startTime string) {
	log.Tracef("%v %v write log", cronId, state)
	_, err := c.logModel.Add(dispatchServer, runServer, cronId, processId, state, output, useTime, remark, startTime)
	if err != nil {
		log.Errorf("onRun c.logModel.Add fail, cron_id=[%v], output=[%v], usetime=[%v], remark=[%v], startTime=[%v], error=[%v]", cronId, output, useTime, remark, startTime, err)
	}
	// onRun 在start状态时会被调用一遍
	// 运行结束的时候也会被运行一遍
	// 所以下面判断真正有写入 +1 > 0 时才写入数据库
	addSuccessNum := int64(0)
	addFailNum    := int64(0)
	if state == StateSuccess {
		addSuccessNum = 1
	} else if state == StateFail || state == StateStart+"-"+StateFail {
		addFailNum = 1
	}
	if addSuccessNum > 0 || addFailNum > 0 {
		c.statisticsModel.Add(cronId, time.Now().Format("2006-01-02"), addSuccessNum, addFailNum)
	}
}

func (c *Controller) SetAvgMaxData() {
	log.Tracef("start SetAvgMaxData ...")
	// 防止锁定时间过长，这里先获取id
	var ids = make([]int64, 0)
	c.lock.RLock()
	for id, _ := range c.cronList {
		ids = append(ids, id)
	}
	c.lock.RUnlock()

	fmt.Fprintf(os.Stderr, "%+v\r\n", ids)

	// 获取平均时长
	avg , _ := c.logModel.GetAvgRunTime()
	for _, id := range ids {
		var avgUseTime, maxUseTime int64 = 0, 0
		if avg != nil {
			avgUseTime = avg[id]
		}
		// 获取最大运行时长
		maxUseTime, _ = c.logModel.GetMaxRunTime(id)
		// 记录数据
		c.statisticsModel.SetAvgMAxUseTime(avgUseTime, maxUseTime, id)
		// 写平均运行时长
		// 写最大运行时长
		c.lock.RLock()
		r, ok := c.cronList[id]
		if ok {
			r.setAvgMAx(avgUseTime, maxUseTime)
		}
		c.lock.RUnlock()
	}
}