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
)

type CronManager struct {
	cronController *cron.Controller
	cronModel *mcron.DbCron
	logModel  *modelLog.DbLog
	httpServer *http.HttpServer
}

func NewManager(db *sql.DB, listen string) *CronManager {
	cronModel := mcron.NewCron(db)
	logModel  := modelLog.NewLog(db)
	cronController := cron.NewController(logModel)
	m := &CronManager{
		cronController:cronController,
		cronModel:cronModel,
		logModel: logModel,
	}
	m.init()
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
		return nil
	}
	m.httpServer = http.NewHttpServer(
		listen,
		http.SetRoute("GET",  "/log/list/{cron_id}/{page}/{limit}", m.logs),
		http.SetRoute("GET",  "/cron/list",        m.cronList),
		http.SetRoute("GET",  "/cron/stop/{id}",   m.stopCron),
		http.SetRoute("GET",  "/cron/start/{id}",  m.startCron),
		http.SetRoute("GET",  "/cron/delete/{id}", m.deleteCron),
		http.SetRoute("POST", "/cron/update/{id}", m.updateCron),
		http.SetRoute("POST", "/cron/add",         m.addCron),
		http.SetHandle("/ui/", shttp.StripPrefix("/ui/", shttp.FileServer(statikFS))),
	)
	m.httpServer.Start()
	return m
}

func (m *CronManager) init() {
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

