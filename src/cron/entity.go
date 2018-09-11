package cron

import (
	//log "github.com/cihub/seelog"
	"gitlab.xunlei.cn/xllive/common/log"
	cronv2 "library/cron"
	"models/cron"
	"os/exec"
	"time"
	"sync/atomic"
	time2 "library/time"
	"sync"
	"encoding/json"
	"bytes"
	"context"
	"errors"
	"os"
	"fmt"
)

// 数据库的基本属性
type CronEntity struct {
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
}
var ErrTimeout = errors.New("timeout")
var ErrUnknown = errors.New("unknown error")

type FilterMiddleWare func(entity *CronEntity) IFilter
type OnRunCommandFunc func(cron_id int64, processId int, state, output string, usetime int64, remark, startTime string)
func newCronEntity(entity *cron.CronEntity, onRun OnRunCommandFunc) *CronEntity {
	e := &CronEntity{
		Id:         entity.Id,
		CronSet:    entity.CronSet,
		Command:    entity.Command,
		Remark:     entity.Remark,
		Stop:       entity.Stop,
		CronId:     0,
		StartTime:  entity.StartTime,
		EndTime:    entity.EndTime,
		IsMutex:    entity.IsMutex,
		onRun:      onRun,
		ProcessNum: 0,
		runid:      0,
		lock:       new(sync.RWMutex),
		copy:       nil,
		Process:    make(map[int]*os.Process),
	}
	// 这里是标准的停止运行过滤器
	// 如果stop设置为true
	// 如果不在指定运行时间范围之内
	e.filter = StopMiddleware()(e)
	e.filter = TimeMiddleware(e.filter)(e)
	return e
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

func (row *CronEntity) Run() {

	if row.filter.Stop() {
		// 外部注入，停止执行定时任务支持
		log.Tracef("%+v was stop", row.Id)
		return
	}

	// 不需要互斥运行
	if !row.IsMutex {
		go row.runCommand()
		return
	}

	// 如果需要互斥运行
	// 判断是否正在运行，0代表不是正在运行中
	if 0 == atomic.LoadInt64(&row.ProcessNum) {
		go row.runCommand()
	}
}

func (row *CronEntity) toJson() (string,error) {
	row.lock.RLock()
	d, e := json.Marshal(row)
	row.lock.RUnlock()
	return string(d), e
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
	row.copy.ProcessNum = atomic.LoadInt64(&row.ProcessNum)
	return row.copy
}

func (row *CronEntity) runCommand() {

	row.lock.Lock()
	processNum := atomic.AddInt64(&row.ProcessNum, 1)
	row.lock.Unlock()

	var cmd *exec.Cmd
	var err error
	rid := atomic.AddInt64(&row.runid, 1)

	startTime := time2.GetDayTime()
	log.Tracef( "##start run: %v, %v=>[%+v,%v]", rid, processNum, row.Id, row.Command)
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

	row.onRun(row.Id, processId, state, output, 0, row.Command, startTime)
	err = cmd.Wait()

	res := b.Bytes()

	//res, err := cmd.CombinedOutput()
	state = StateSuccess
	if err != nil {
		state = StateFail
		res = append(res, []byte("  error: " + err.Error())...)
		log.Errorf("runCommand fail, id=[%v], command=[%v], error=[%+v]", row.Id, row.Command, err)
	}
	log.Tracef( "##%v run end: %v, %v=>[%+v,%v]", rid, processNum, row.Id, row.Command, processId)
	atomic.AddInt64(&row.ProcessNum, -1)

	useTime := int64(time.Now().UnixNano()/1000000 - start)
	// todo 程序退出时，定时任务的日志可能会失败，因为这个时候数据库已经关闭，这个问题需要处理一下
	// 即安全退出问题，kill -9没办法了
	row.onRun(row.Id, processId, state, string(res), useTime, row.Command, startTime)
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

