package manager

import (
	"cron"
	mcron "models/cron"
	"database/sql"
	seelog "gitlab.xunlei.cn/xllive/common/log"
	"library/http"
)

type CronManager struct {
	cronController *cron.Controller
	cronModel *mcron.DbCron
	httpServer *http.HttpServer
}

func NewManager(db *sql.DB) *CronManager {
	cronModel := mcron.NewCron(db)
	cronController := cron.NewController(db)
	m := &CronManager{
		cronController:cronController,
		cronModel:cronModel,
	}
	m.init()
	m.httpServer = http.NewHttpServer(
		"0.0.0.0:98001",
		//http.SetRoute("GET",  "/log/list",         logApi.logs),
		//http.SetRoute("GET",  "/cron/list",        cronApi.list),
		//http.SetRoute("GET",  "/cron/stop/{id}",   cronApi.stop),
		//http.SetRoute("GET",  "/cron/start/{id}",  cronApi.start),
		//http.SetRoute("GET",  "/cron/delete/{id}", cronApi.delete),
		//http.SetRoute("POST", "/cron/update",      cronApi.update),
		http.SetRoute("POST", "/cron/add",         m.addCron),
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

