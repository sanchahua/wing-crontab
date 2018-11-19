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
	"encoding/json"
	"runtime"
	"sync/atomic"
	"models/user"
	"service"
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
	cronList *sync.Map
	lock     *sync.RWMutex
	status   int
	logModel *modelLog.DbLog
	cache    ListCronEntity
	statisticsModel *statistics.Statistics
	userModel       *user.User
	//Leader          int64
	redis           *redis.Client
	RedisKeyPrex    string
	ready           int64
	service *service.Service
}

func NewController(service *service.Service, redis *redis.Client, RedisKeyPrex string,
	logModel *modelLog.DbLog,
	statisticsModel *statistics.Statistics, userModel *user.User) *Controller {
	c := &Controller{
		service: service,
		serviceId: 0,
		cron:     cronV2.New(),
		cronList: new(sync.Map),
		lock:     new(sync.RWMutex),
		status:   0,
		logModel: logModel,
		cache:    nil,
		statisticsModel: statisticsModel,
		//Leader:   0,
		redis: redis,
		RedisKeyPrex: RedisKeyPrex,
		userModel: userModel,
		ready: 0,
	}
	go c.dispatch()
	return c
}

func (c *Controller) Ready() {
	atomic.StoreInt64(&c.ready, 1)
}

func (c *Controller) dispatch() {
	for {
		if 1 == atomic.LoadInt64(&c.ready) {
			break
		}
		log.Tracef("wait for xcrontab be ready")
		time.Sleep(time.Second)
	}
	var raw = make([]int64, 0)
	cpuNum := int64(runtime.NumCPU())
	gonum  := int64(0)
	for {
		if c.service.IsOffline() {
			log.Warnf("node is offline")
			time.Sleep(time.Second)
			continue
		}
		// 最大运行线程数量，取 max(定时任务数量, cpu数量)
		maxNum := int64(0)
		c.cronList.Range(func(key, value interface{}) bool {
			maxNum++
			return true
		})

		// 这里必须每次都重新拿长度数据，因为cronList可能会实时改变
		if cpuNum > maxNum {
			maxNum = cpuNum
		}

		qlen := c.redis.LLen(c.RedisKeyPrex).Val()
		log.Infof("当前待运行的定时任务: %v", qlen)
        // 如果队列太常，加个告警??

		// 如果当前正在执行的线程数量达到上限，则等待
		if atomic.LoadInt64(&gonum) >= maxNum {
			log.Warnf("正在运行的协程数量(%v)达到上限%v，等待中...", atomic.LoadInt64(&gonum), maxNum)
			time.Sleep(time.Millisecond * 500)
			continue
		}

		data, err := c.redis.BLPop(time.Second*3, c.RedisKeyPrex).Result()
		if err != nil {
			log.Errorf("dispatch row.redis.BLPop fail, error=[%v]", err)
			continue
		}

		if len(data) < 2 {
			log.Warnf("dispatch data error: %+v", data)
			continue
		}

		err = json.Unmarshal([]byte(data[1]), &raw)
		if err != nil {
			log.Errorf("dispatch json.Unmarshal fail, error=[%v]", err)
			continue
		}
		if len(raw) < 2 {
			log.Errorf("dispatch raw len fail, error=[%v]", err)
			continue
		}
		serviceId := raw[0]
		id        := raw[1]
		irow, ok  := c.cronList.Load(id)

		if !ok {
			log.Errorf("dispatch id not exists fail, id=[%v]", id)
			continue
		}

		row, ok := irow.(*CronEntity)
		if !ok {
			log.Errorf("dispatch convert fail, id=[%v]", id)
			continue
		}
		atomic.AddInt64(&gonum, 1)
		go row.runCommand(serviceId, func() {
			atomic.AddInt64(&gonum, -1)
		})
	}
}

//func (c *Controller) SetLeader(isLeader bool) {
	//if isLeader {
	//	atomic.StoreInt64(&c.Leader, 1)
	//} else {
	//	atomic.StoreInt64(&c.Leader, 0)
	//}
	//c.cronList.Range(func(key, value interface{}) bool {
	//	//v, ok := value.(*CronEntity)
	//	//if ok {
	//		//v.SetLeader(isLeader)
	//	//}
	//	return true
	//})
//}

func (c *Controller) SetServiceId(serviceId int64) {
	c.serviceId = serviceId
	c.cronList.Range(func(key, value interface{}) bool {
		v, ok := value.(*CronEntity)
		if ok {
			v.SetServiceId(serviceId)
		}
		return true
	})
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
	var err error
	uinfo, _ := c.userModel.GetUserInfo(ce.UserId)
	blameInfo, _ := c.userModel.GetUserInfo(ce.Blame)
	entity := newCronEntity(c.service, c.redis, c.RedisKeyPrex, ce,
		uinfo, blameInfo, c.onRun)
	entity.SetServiceId(c.service.ID)
	//entity.SetLeader(atomic.LoadInt64(&c.Leader) == 1)

	c.cronList.Store(entity.Id, entity)
	entity.CronId, err = c.cron.AddJob(entity.CronSet, entity)
	if err != nil {
		log.Errorf("%+v", err)
		return entity, err
	}
	return entity, nil
}

func (c *Controller) Delete(id int64) (*CronEntity, error) {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	c.cronList.Delete(id)
	e, ok := ie.(*CronEntity)
	if !ok {
		return nil, errors.New(fmt.Sprintf("id convert fail, id=[%v]", id))
	}
	c.cron.Remove(e.CronId)
	return e, nil
}

func (c *Controller) Update(id int64, cronSet, command string,
	remark string, stop bool, startTime,
	endTime string, isMutex bool, blame int64) (*CronEntity, error) {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}

	c.cronList.Delete(id)
	e, ok := ie.(*CronEntity)
	if !ok {
		return nil, errors.New(fmt.Sprintf("id convert fail, id=[%v]", id))
	}
	c.cron.Remove(e.CronId)

	uinfo, _ := c.userModel.GetUserInfo(e.UserId)
	blameInfo, _ := c.userModel.GetUserInfo(blame)
	entity := newCronEntity(c.service, c.redis, c.RedisKeyPrex, &cron.CronEntity{
		Id:        id,
		CronSet:   cronSet,
		Command:   command,
		Remark:    remark,
		Stop:      stop,
		StartTime: startTime,
		EndTime:   endTime,
		IsMutex:   isMutex,
		Blame: blame,
	}, uinfo, blameInfo, c.onRun)
	entity.SetServiceId(c.service.ID)
	//entity.SetLeader(atomic.LoadInt64(&c.Leader) == 1)

	var err error
	entity.CronId, err = c.cron.AddJob(entity.CronSet, entity)
	if err != nil {
		log.Errorf("Updatec.cron.AddJob fail, error=[%v]", err)
	}
	c.cronList.Store(entity.Id, entity)
	return e, nil
}

func (c *Controller) Get(id int64) (*CronEntity, error)  {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return nil, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e, ok := ie.(*CronEntity)
	if !ok {
		return nil, errors.New(fmt.Sprintf("id convert fail, id=[%v]", id))
	}
	return e, nil
}

// timeout 超时，单位秒
func (c *Controller) RunCommand(id int64, timeout int64) ([]byte, int, error)  {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return nil, 0, errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	if timeout < 1 {
		timeout = 3
	}
	e, ok := ie.(*CronEntity)
	if !ok {
		return nil, 0, errors.New(fmt.Sprintf("id convert fail, id=[%v]", id))
	}
	return e.runCommandWithTimeout(time.Duration(timeout) * time.Second)
}

func (c *Controller) Kill(id int64, processId int) {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return
	}
	e, ok := ie.(*CronEntity)
	if !ok {
		return
	}
	e.Kill(processId)
}

func (c *Controller) ProcessIsRunning(id int64, processId int) bool {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return false
	}
	e, ok := ie.(*CronEntity)
	if !ok {
		return false
	}
	return e.ProcessIsRunning(processId)
}

// 已处理线程安全问题
// 所有内容使用只读cache
func (c *Controller) GetList() ListCronEntity {
	l := 0
	c.cronList.Range(func(key, value interface{}) bool {
		l++
		return true
	})
	if c.cache == nil || len(c.cache) != l {
		c.cache = make(ListCronEntity, l)
	}
	i := 0

	c.cronList.Range(func(key, value interface{}) bool {
		v, ok := value.(*CronEntity)
		if ok {
			c.cache[i] = v.Clone()
			i++
		}
		return true
	})

	sort.Sort(c.cache)
	return c.cache
}

func (c *Controller) Stop(id int64, stop bool) error {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e, ok := ie.(*CronEntity)
	if !ok {
		return errors.New(fmt.Sprintf("id convert fail, id=[%v]", id))
	}
	e.setStop(stop)
	return nil
}

func (c *Controller) Mutex(id int64, mutex bool) error {
	ie, ok := c.cronList.Load(id)
	if !ok {
		return errors.New(fmt.Sprintf("id does not exists, id=[%v]", id))
	}
	e, ok := ie.(*CronEntity)
	if !ok {
		return errors.New(fmt.Sprintf("id convert fail, id=[%v]", id))
	}
	e.setMutex(mutex)
	return nil
}

func (c *Controller) onRun(dispatchServer, runServer int64, cronId int64, processId int, state, output string, useTime int64, remark, startTime string) {
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
	var ids = make([]int64, 0)
	c.cronList.Range(func(key, value interface{}) bool {
		id, ok := key.(int64)
		if ok {
			ids = append(ids, id)
		}
		return true
	})

	//fmt.Fprintf(os.Stderr, "%+v\r\n", ids)
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
		ir, ok := c.cronList.Load(id)//[id]
		if ok {
			r, ok := ir.(*CronEntity)
			if ok {
				r.setAvgMax(avgUseTime, maxUseTime)
			}
		}
	}
}