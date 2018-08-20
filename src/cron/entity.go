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
	onRun      OnRunCommandFunc `json:"-"`
	runid      int64            `json:"-"`
	isRunning  int64              `json:"-"`
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
	}
	// 这里是标准的停止运行过滤器
	// 如果stop设置为true
	// 如果不在指定运行时间范围之内
	e.filter = StopMiddleware()(e)
	e.filter = TimeMiddleware(e.filter)(e)
	return e
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

func (row *CronEntity) runCommand() {
	processNum := atomic.AddInt64(&row.ProcessNum, 1)
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
	row.onRun(row.Id, string(res), useTime, row.Command, startTime)
}
