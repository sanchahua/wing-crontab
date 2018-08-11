package cron

// 定时任务实体对象
import (
	log "github.com/cihub/seelog"
	cronv2 "library/cron"
	"models/cron"
	"os/exec"
	"time"
)

// 数据库的基本属性
type CronEntity struct {
	CronId    cronv2.EntryID  `json:"cron_id"`
	Id        int64           `json:"id"`
	CronSet   string          `json:"cron_set"`
	Command   string          `json:"command"`
	Remark    string          `json:"remark"`
	Stop      bool            `json:"stop"`
	StartTime int64           `json:"start_time"`
	EndTime   int64           `json:"end_time"`
	IsMutex   bool            `json:"is_mutex"`
	//onWillRun OnWillRunFunc   `json:"-"`
	filter    IFilter         `json:"-"`
	//WaitNum   int64           `json:"wait_num"`
	IsRunning bool            `json:"is_running"`
	runChan chan struct{}     `json:"-"`
	exitChan chan struct{}    `json:"-"`
	onRun OnRunCommandFunc    `json:"-"`
}
type CronEntityMiddleWare func(entity *CronEntity) IFilter
type OnRunCommandFunc func(cron_id int64, output string, usetime int64, remark string)
const RunChanLen = 60
func newCronEntity(entity *cron.CronEntity, onRun OnRunCommandFunc) *CronEntity {
	e := &CronEntity{
		Id:        entity.Id,
		CronSet:   entity.CronSet,
		Command:   entity.Command,
		Remark:    entity.Remark,
		Stop:      entity.Stop,
		CronId:    0,
		//onWillRun: onWillRun,
		StartTime: entity.StartTime,
		EndTime:   entity.EndTime,
		IsMutex:   entity.IsMutex,
		runChan:   make(chan struct{}, RunChanLen),
		exitChan:  make(chan struct{}, 3),
		onRun:     onRun,
	}
	// 这里是标准的停止运行过滤器
	// 如果stop设置为true
	// 如果不在指定运行时间范围之内
	e.filter = StopMiddleware()(e)
	e.filter = TimeMiddleware(e.filter)(e)
	go e.run()
	return e
}

//func (row *CronEntity) SubWaitNum() int64 {
//	if atomic.LoadInt64(&row.WaitNum) <= 0 {
//		return 0
//	}
//	return atomic.AddInt64(&row.WaitNum, -1)
//}
//
//func (row *CronEntity) AddWaitNum() {
//	atomic.AddInt64(&row.WaitNum, 1)
//}
func (row *CronEntity) delete() {
	row.exitChan <- struct{}{}
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
		log.Debugf("%+v was stop", row.Id)
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
	//row.onWillRun(row.Id, row.Command, row.IsMutex, row.AddWaitNum, row.SubWaitNum)
	//fmt.Fprintf(os.Stderr, "\r\n########## only leader do this %+v\r\n\r\n", *row)
}

func (row *CronEntity) runCommand() {
	//for _, f := range c.onBefore {
	//	f(id, dispatchServer, runServer, []byte(""), 0)
	//}
	var cmd *exec.Cmd
	var err error
	start := time.Now().UnixNano()/1000000
	cmd = exec.Command("bash", "-c", row.Command)
	res, err := cmd.CombinedOutput()
	if err != nil {
		res = append(res, []byte("  error: " + err.Error())...)
		log.Errorf("runCommand fail, id=[%v], command=[%v], error=[%+v]", row.Id, row.Command, err)
	}
	log.Infof( "##########################[%+v,%v] was run##########################", row.Id, row.Command)
	//for _, f := range c.onAfter {
	//	f(id, dispatchServer, runServer, res, time.Since(start))
	//}
	usetime := int64(time.Now().UnixNano()/1000000 - start)
	row.onRun(row.Id, string(res), usetime, row.Command)
}
