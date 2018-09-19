package cron

import (
	//log "github.com/cihub/seelog"
	"gitlab.xunlei.cn/xllive/common/log"
	cronv2 "library/cron"
	"models/cron"
	"os/exec"
	"time"
	time2 "library/time"
	"sync"
	"bytes"
	"context"
	"errors"
	"os"
	"fmt"
	"github.com/go-redis/redis"
	"sync/atomic"
	"encoding/json"
)

// 数据库的基本属性
type CronEntity struct {
	ServiceId  int64              `json:"service_id"`
	CronId     cronv2.EntryID     `json:"cron_id"`
	Id         int64              `json:"id"`
	CronSet    string             `json:"cron_set"`
	Command    string             `json:"command"`
	Remark     string             `json:"remark"`
	Stop       bool               `json:"stop"`
	StartTime  string             `json:"start_time"`
	EndTime    string             `json:"end_time"`
	IsMutex    bool               `json:"is_mutex"`
	filter     IFilter            `json:"-"`
	// 当前正在同事运行的进程数
	ProcessNum int64              `json:"process_num"`
	onRun      OnRunCommandFunc   `json:"-"`
	runid      int64              `json:"-"`
	lock       *sync.RWMutex      `json:"-"`
	copy       *CronEntity        `json:"-"`
	Process    map[int]*os.Process        `json:"-"`
	Leader bool  `json:"-"`
	redis *redis.Client `json:"-"`
	redisKeyPrex string  `json:"-"`
	AvgRunTime int64 `json:"avg_run_time"`
	MaxRunTime int64 `json:"max_run_time"`
	exit bool `json:"-"`

}
var ErrTimeout = errors.New("timeout")
var ErrUnknown = errors.New("unknown error")
var GoId int64 = 0
const DefaultTimeout = 6 //秒

type FilterMiddleWare func(entity *CronEntity) IFilter
type OnRunCommandFunc func(dispatchServer, runServer int64, cron_id int64, processId int, state, output string, usetime int64, remark, startTime string)
func newCronEntity(
	redis *redis.Client,
	redisKeyPrex string,
	entity *cron.CronEntity,
	onRun OnRunCommandFunc,
) *CronEntity {
	e := &CronEntity{
		ServiceId:    0,
		Id:           entity.Id,
		CronSet:      entity.CronSet,
		Command:      entity.Command,
		Remark:       entity.Remark,
		Stop:         entity.Stop,
		CronId:       0,
		StartTime:    entity.StartTime,
		EndTime:      entity.EndTime,
		IsMutex:      entity.IsMutex,
		onRun:        onRun,
		ProcessNum:   0,
		runid:        0,
		lock:         new(sync.RWMutex),
		copy:         nil,
		Process:      make(map[int]*os.Process),
		redis:        redis,
		redisKeyPrex: redisKeyPrex,
		MaxRunTime:   10000,
		AvgRunTime:   10000,
		exit: false,
	}
	// 这里是标准的停止运行过滤器
	// 如果stop设置为true
	// 如果不在指定运行时间范围之内
	e.filter = StopMiddleware()(e)
	e.filter = TimeMiddleware(e.filter)(e)
	go e.dispatch(atomic.AddInt64(&GoId, 1))
	return e
}

func (row *CronEntity) Exit() {
	row.lock.Lock()
	row.exit = true
	row.lock.Unlock()
}

func (row *CronEntity) setStop(stop bool) {
	row.lock.Lock()
	row.Stop = stop
	row.lock.Unlock()
}

func (row *CronEntity) setMutex(mutex bool) {
	row.lock.Lock()
	row.IsMutex = mutex
	row.lock.Unlock()
}

func (row *CronEntity) setAvgMAx(avg, max int64) {
	row.lock.Lock()
	log.Tracef("%v set avg=%v, max=%v", row.Id, avg, max)
	row.AvgRunTime = avg
	row.MaxRunTime = max
	row.lock.Unlock()
}

func (row *CronEntity) addProcessNum()  {
	// 这里使用redis incr原子递增增加正在运行的并行进程数
	// 超时时间设置为该进程历史运行的平均时间，单位为秒
	key := fmt.Sprintf(row.redisKeyPrex+"/%v/process_num", row.Id)
	err := row.redis.Incr(key).Err()
	if err != nil {
		log.Errorf("addProcessNum row.redis.Incr fail, key=[%v], error=[%v]", key, err)
	}
	err = row.redis.Expire(key, DefaultTimeout * time.Second).Err()
	if err != nil {
		log.Errorf("addProcessNum row.redis.Expire fail, key=[%v], error=[%v]", key, err)
	}
}

func (row *CronEntity) SetServiceId(serviceId int64) {
	row.lock.Lock()
	row.ServiceId = serviceId
	row.lock.Unlock()
}

func (row *CronEntity) subProcessNum()  {
	// 这里使用redis incr原子递减减少正在运行的并行进程数
	// 超时时间设置为该进程历史运行的平均时间，单位为秒
	key := fmt.Sprintf(row.redisKeyPrex+"/%v/process_num", row.Id)
	err := row.redis.IncrBy(key, -1).Err()
	if err != nil {
		log.Errorf("subProcessNum row.redis.IncrBy fail, key=[%v], error=[%v]", key, err)
	}
	err = row.redis.Expire(key, DefaultTimeout * time.Second).Err()
	if err != nil {
		log.Errorf("subProcessNum row.redis.Expire fail, key=[%v], error=[%v]", key, err)
	}
}

func (row *CronEntity) getProcessNum() (int64, error)  {
	key := fmt.Sprintf(row.redisKeyPrex+"/%v/process_num", row.Id)
	num, err := row.redis.Get(key).Int64()
	if err == redis.Nil || err == nil {
		return num, nil
	}
	return num, err
}

// 定时任务管理引擎 接口
func (row *CronEntity) Run() {
	// 只有leader负责定时任务调度
	row.lock.RLock()
	if !row.Leader {
		row.lock.RUnlock()
		return
	}
	row.lock.RUnlock()
	if row.filter.Stop() {
		// 外部注入，停止执行定时任务支持
		return
	}
	// 推送到redis
	row.push()
}

func (row *CronEntity) dispatch(goid int64) {
	queue := fmt.Sprintf(row.redisKeyPrex+"/%v", row.Id)
	var raw = make([]int64, 0)

	for {
		row.lock.RLock()
		if row.exit {
			row.lock.RUnlock()
			return
		}

		row.lock.RUnlock()
		data, err := row.redis.BLPop(time.Second*3, queue).Result()
		if err != nil {
			if err != redis.Nil {
				log.Errorf("dispatch row.redis.BLPop fail, queue=[%v], error=[%v]", queue, err)
			}
			continue
		}

		if len(data) < 2 {
			continue
		}

		fmt.Println(data)
		err = json.Unmarshal([]byte(data[1]), &raw)
		if err != nil {
			log.Errorf("dispatch json.Unmarshal fail, queue=[%v], error=[%v]", queue, err)
			continue
		}
		if len(raw) < 2 {
			log.Errorf("dispatch raw len fail, queue=[%v], error=[%v]", queue, err)
			continue
		}
		serviceId := raw[0]
		id        := raw[1]
		if !row.IsMutex {
			log.Infof("##%v => %v was run", goid, id)
			// 不需要互斥运行
			go row.runCommand(serviceId)
		} else {
			// 必须严格互斥运行
			processNum, err := row.getProcessNum()
			if processNum > 0 {
				log.Infof("%v => %v has running process", goid, id)
			}
			if processNum <= 0 && err == nil {
				log.Infof("###%v => %v was run", goid, id)
				go row.runCommand(serviceId)
			}
			if err != nil {
				log.Errorf("dispatch row.getProcessNum fail, queue=[%v], error=[%v]", queue, err)
			}
		}
	}
}

func (row *CronEntity) push() {
	key := fmt.Sprintf(row.redisKeyPrex+"/%v", row.Id)
	log.Tracef("push %v to %v", row.Id, key)

	data, err := json.Marshal([]int64{row.ServiceId, row.Id})
	if err != nil {
		log.Errorf("push json.Marshal fail, [%v] to [%v/%v], error=[%v]", row.Id, row.redisKeyPrex, row.Id, err)
		return
	}
	err = row.redis.RPush(key, string(data)).Err()
	if err != nil {
		log.Errorf("push row.redis.RPush fail, [%v] to [%v/%v], error=[%v]", row.Id, row.redisKeyPrex, row.Id, err)
	}
}

func (row *CronEntity) SetLeader(isLeader bool) {
	row.lock.Lock()
	row.Leader = isLeader
	row.lock.Unlock()
}

func (row *CronEntity) Clone() *CronEntity {
	row.lock.RLock()
	defer row.lock.RUnlock()
	if row.copy == nil {
		row.copy = new(CronEntity)
	}
	row.copy.CronId     = row.CronId
	row.copy.Id         = row.Id
	row.copy.CronSet    = row.CronSet
	row.copy.Command    = row.Command
	row.copy.Remark     = row.Remark
	row.copy.Stop       = row.Stop
	row.copy.StartTime  = row.StartTime
	row.copy.EndTime    = row.EndTime
	row.copy.IsMutex    = row.IsMutex
	row.copy.ProcessNum, _ = row.getProcessNum()
	row.copy.AvgRunTime = row.AvgRunTime
	row.copy.MaxRunTime = row.MaxRunTime
	return row.copy
}

// serviceId 为leader服务id
func (row *CronEntity) runCommand(serviceId int64) {

	// todo 这里需要加上平行锁策略，防止死锁，过长锁定等等

	row.addProcessNum()

	var cmd *exec.Cmd
	var err error

	startTime := time2.GetDayTime()
	start := time.Now().UnixNano()/1000000
	cmd = exec.Command("bash", "-c", row.Command)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	err = cmd.Start()
	processId := 0
	if cmd != nil && cmd.Process != nil {
		processId = cmd.Process.Pid

		row.lock.Lock()
		row.Process[cmd.Process.Pid] = cmd.Process
		row.lock.Unlock()

		defer func() {
			row.lock.Lock()
			delete(row.Process, processId)
			row.lock.Unlock()
		}()
	}

	state := StateStart
	output := ""
	if err != nil {
		state = StateFail//StateStart+"-"+StateFail
		output = err.Error()
	}

	row.onRun(serviceId, row.ServiceId, row.Id, processId, state, output, 0, row.Command, startTime)
	err = cmd.Wait()
	res := b.Bytes()

	state = StateSuccess
	if err != nil {
		state = StateFail
		res = append(res, []byte("  error: " + err.Error())...)
		log.Errorf("runCommand fail, id=[%v], command=[%v], error=[%+v]", row.Id, row.Command, err)
	}
	row.subProcessNum()

	useTime := int64(time.Now().UnixNano()/1000000 - start)
	// todo 程序退出时，定时任务的日志可能会失败，因为这个时候数据库已经关闭，这个问题需要处理一下
	// 即安全退出问题，kill -9没办法了
	row.onRun(serviceId, row.ServiceId,row.Id, processId, state, string(res), useTime, row.Command, startTime)
}

func (row *CronEntity) runCommandWithTimeout(duration time.Duration) ([]byte, int, error) {
	ctx, _ := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "bash", "-c", row.Command)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Start()
	if err != nil {
		return nil, 0, err
	}

	c  := make(chan []byte)
	er := make(chan error)

	defer func() {
		close(c)
		close(er)
	}()

	processId := 0
	if cmd != nil && cmd.Process != nil {
		processId = cmd.Process.Pid
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("err:", err)
			}
		}()
		err := cmd.Wait()
		if err != nil {
			er <- err
		} else {
			c <- b.Bytes()
		}
	}()

	select {
	case r := <- c :
		return r, processId, nil
	case err := <- er:
		return nil, processId, err
	case <- time.After(duration):
		err := cmd.Process.Kill()
		if err != nil {
			return nil, processId, errors.New("timeout with error: "+ err.Error())
		}
		return nil, processId, ErrTimeout
	}
	return nil, processId, ErrUnknown
}

func (row *CronEntity) GetAllProcessId() []int {
	if row == nil {
		return nil
	}
	row.lock.RLock()
	defer row.lock.RUnlock()
	var process = make([]int, 0)
	for pid, _ := range row.Process {
		process = append(process, pid)
	}
	return process
}

func (row *CronEntity) ProcessIsRunning(processId int) bool {
	if row == nil {
		return false
	}
	row.lock.RLock()
	defer row.lock.RUnlock()
	if _, ok :=row.Process[processId]; ok {
		return true
	}
	return false
}

func (row *CronEntity) Kill(processId int) {
	if row == nil {
		return
	}
	row.lock.RLock()
	defer row.lock.RUnlock()
	if pro, ok :=row.Process[processId]; ok {
		pro.Kill()
		return
	}
}

type ListCronEntity []*CronEntity

func (c ListCronEntity) Len() int {
	return len(c)
}
func (c ListCronEntity) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
// 按照id倒叙排序
func (c ListCronEntity) Less(i, j int) bool {
	return c[i].Id > c[j].Id
}

