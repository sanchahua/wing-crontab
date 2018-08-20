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
	runChan    chan struct{}    `json:"-"`
	exitChan   chan struct{}    `json:"-"`
	onRun      OnRunCommandFunc `json:"-"`
}
type FilterMiddleWare func(entity *CronEntity) IFilter
type OnRunCommandFunc func(cron_id int64, output string, usetime int64, remark, startTime string)
const (
	RunChanLen = 60
)
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
		runChan:    make(chan struct{}, RunChanLen),
		exitChan:   make(chan struct{}, 3),
		onRun:      onRun,
		ProcessNum: 0,
	}
	// 这里是标准的停止运行过滤器
	// 如果stop设置为true
	// 如果不在指定运行时间范围之内
	e.filter = StopMiddleware()(e)
	e.filter = TimeMiddleware(e.filter)(e)
	go e.run()
	return e
}

func (row *CronEntity) delete() {
	close(row.exitChan)
	close(row.runChan)
}

func (row *CronEntity) run() {
	for {
		select {
		case <-row.runChan:
			row.runCommand()
		case <-row.exitChan:
			return
		}
	}
}

func (row *CronEntity) Run() {

	if row.filter.Stop() {
		// 外部注入，停止执行定时任务支持
		log.Tracef("%+v was stop", row.Id)
		return
	}

	// 如果需要互斥运行
	if row.IsMutex {
		if len(row.runChan) < RunChanLen {
			row.runChan <- struct{}{}
		}
		return
	}

	// 不需要互斥运行
	go row.runCommand()
}

func (row *CronEntity) runCommand() {
	processNum := atomic.AddInt64(&row.ProcessNum, 1)
	var cmd *exec.Cmd
	var err error
	startTime := time2.GetDayTime()
	start := time.Now().UnixNano()/1000000
	cmd = exec.Command("bash", "-c", row.Command)
	res, err := cmd.CombinedOutput()
	if err != nil {
		res = append(res, []byte("  error: " + err.Error())...)
		log.Errorf("runCommand fail, id=[%v], command=[%v], error=[%+v]", row.Id, row.Command, err)
	}
	log.Tracef( "##########################%v=>[%+v,%v] was run##########################", processNum, row.Id, row.Command)
	atomic.AddInt64(&row.ProcessNum, -1)
	useTime := int64(time.Now().UnixNano()/1000000 - start)
	row.onRun(row.Id, string(res), useTime, row.Command, startTime)
}
