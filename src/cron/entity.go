package cron

import (
	//log "github.com/cihub/seelog"
	log "gitlab.xunlei.cn/xllive/common/log"
	cronv2 "library/cron"
	"models/cron"
	"os/exec"
	"time"
	"sync/atomic"
	time2 "library/time"
	"sync"
	"encoding/json"
)

// 数据库的基本属性
type CronEntity struct {
	CronId     cronv2.EntryID   `json:"cron_id"`
	Id         int64            `json:"id"`
	CronSet    string           `json:"cron_set"`
	Command    string           `json:"command"`
	Remark     string           `json:"remark"`
	Stop       bool             `json:"stop"`
	StartTime  int64            `json:"start_time"`
	EndTime    int64            `json:"end_time"`
	IsMutex    bool             `json:"is_mutex"`
	filter     IFilter          `json:"-"`
	// 当前正在同事运行的进程数
	ProcessNum int64            `json:"process_num"`
	onRun      OnRunCommandFunc `json:"-"`
	runid      int64            `json:"-"`
	isRunning  int64            `json:"-"`
	lock       *sync.RWMutex    `json:"-"`
	copy       *CronEntity      `json:"-"`
}
type FilterMiddleWare func(entity *CronEntity) IFilter
type OnRunCommandFunc func(cron_id int64, output string, usetime int64, remark, startTime string)
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
	if 0 == atomic.LoadInt64(&row.isRunning) {
		go row.runCommand()
	}
}

func (row *CronEntity) toJson() (string,error) {
	row.lock.RLock()
	d, e := json.Marshal(row)
	row.lock.RUnlock()
	return string(d), e
}

func (row *CronEntity) Update(cronSet, command string, remark string, stop bool, startTime, endTime int64, isMutex bool) {
	row.lock.Lock()
	row.CronSet   = cronSet
	row.Command   = command
	row.Stop      = stop
	row.Remark    = remark
	row.StartTime = startTime
	row.EndTime   = endTime
	row.IsMutex   = isMutex
	row.lock.Unlock()
}
func (row *CronEntity) clone() *CronEntity {
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

	atomic.StoreInt64(&row.isRunning, 1)
	defer atomic.StoreInt64(&row.isRunning, 0)

	startTime := time2.GetDayTime()
	log.Tracef( "##########################%v, %v=>[%+v,%v] start run##########################", rid, processNum, row.Id, row.Command)
	start := time.Now().UnixNano()/1000000
	cmd = exec.Command("bash", "-c", row.Command)
	res, err := cmd.CombinedOutput()
	if err != nil {
		res = append(res, []byte("  error: " + err.Error())...)
		log.Errorf("runCommand fail, id=[%v], command=[%v], error=[%+v]", row.Id, row.Command, err)
	}
	log.Tracef( "##########################%v, %v=>[%+v,%v] run end##########################", rid, processNum, row.Id, row.Command)
	atomic.AddInt64(&row.ProcessNum, -1)

	useTime := int64(time.Now().UnixNano()/1000000 - start)
	// todo 程序退出时，定时任务的日志可能会失败，因为这个时候数据库已经关闭，这个问题需要处理一下
	row.onRun(row.Id, string(res), useTime, row.Command, startTime)
}
