package manager

import (
	"cron"
	mcron "models/cron"
	"database/sql"
	seelog "gitlab.xunlei.cn/xllive/common/log"
	"library/http"
	modelLog "models/log"
	shttp "net/http"
	_ "statik"
	"github.com/rakyll/statik/fs"
	"time"
	time2 "library/time"
	"models/statistics"
)

type CronManager struct {
	cronController *cron.Controller
	cronModel *mcron.DbCron
	logModel  *modelLog.DbLog
	httpServer *http.HttpServer
	logKeepDay int64
	statisticsModel *statistics.Statistics
}

func NewManager(db *sql.DB, listen string, logKeepDay int64) *CronManager {
	cronModel := mcron.NewCron(db)
	logModel  := modelLog.NewLog(db)
	statisticsModel := statistics.NewStatistics(db)
	cronController := cron.NewController(logModel, statisticsModel)
	m := &CronManager{
		cronController:cronController,
		cronModel:cronModel,
		logModel: logModel,
		logKeepDay: logKeepDay,
		statisticsModel: statisticsModel,
	}
	m.init()
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
		return nil
	}
	m.httpServer = http.NewHttpServer(
		listen,
		http.SetRoute("GET",  "/log/list/{cron_id}/{search_fail}/{page}/{limit}", m.logs),
		http.SetRoute("GET",  "/cron/list",                   m.cronList),
		http.SetRoute("GET",  "/cron/stop/{id}",              m.stopCron),
		http.SetRoute("GET",  "/cron/start/{id}",             m.startCron),
		http.SetRoute("GET",  "/cron/mutex/false/{id}",       m.mutexFalse),
		http.SetRoute("GET",  "/cron/mutex/true/{id}",        m.mutexTrue),
		http.SetRoute("GET",  "/cron/delete/{id}",            m.deleteCron),
		http.SetRoute("POST", "/cron/update/{id}",            m.updateCron),
		http.SetRoute("POST", "/cron/add",                    m.addCron),
		http.SetRoute("GET",  "/cron/info/{id}",              m.cronInfo),
		http.SetRoute("GET",  "/index",                       m.index),
		http.SetRoute("GET",  "/charts/{days}",               m.charts),
		http.SetRoute("GET",  "/cron/run/{id}/{timeout}",     m.cronRun),
		http.SetRoute("GET",  "/cron/kill/{id}/{process_id}", m.cronKill),
		http.SetRoute("GET",  "/cron/log/detail/{id}",        m.cronLogDetail),
		http.SetHandle("/ui/", shttp.StripPrefix("/ui/", shttp.FileServer(statikFS))),
	)
	m.httpServer.Start()
	go m.logManager()
	return m
}

func (m *CronManager) logManager() {
	logKeepDay := m.logKeepDay
	if logKeepDay < 1 {
		logKeepDay = 1
	}
	// 日志清理操作，每60秒执行一次
	for {
		m.logModel.DeleteByStartTime(time2.TimeFormat(time.Now().Unix()-logKeepDay*86400))
		time.Sleep(time.Second * 60)
	}
}

func (m *CronManager) init() {
	// 启动后，先进性服务注册
	// 然后获取节点数量，当前节点id=节点数量
	// 从数据库查询 定时任务id%节点数量==0的定时任务，载入当前的定时任务管理器


	// 节点数量变化产生的流程
	// 如果获取到节点退出事件
	// id也需要重新生成，注意并发加锁与id唯一性即可
	// 如果当前的节点id == 节点数量
	// 从数据库查询 定时任务id%节点数量==0的定时任务，载入当前的定时任务管理器
	// 如果当前的节点id < 节点数量
	// 从数据库查询 定时任务id%节点数量==节点id 的定时任务，载入当前的定时任务管理器
	// 最后判断，当前的定时任务管理器内的定时任务是否在查询列表内，不在则清除
	// 然后再判断返回列表的定时任务，如果该定时任务已存在当前的定时任务管理器，不予处理
	// 不存在，则加入到当前的定时任务管理器中

	list, err := m.cronModel.GetList()
	if err != nil {
		seelog.Errorf("init fail, m.cronModel.GetList fail, error=[%v]", err)
		return
	}
	for _, data := range list {
		m.cronController.Add(data)
	}
}

func (m *CronManager) Start() {
	m.cronController.StartCron()
}

func (m *CronManager) Stop() {
	m.cronController.StartCron()
}

