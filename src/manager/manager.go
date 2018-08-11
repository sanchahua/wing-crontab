package manager

import (
	"cron"
	mcron "models/cron"
	"database/sql"
	"github.com/cihub/seelog"
)

type CronManager struct {
	cronController *cron.CronController
	cronModel *mcron.DbCron
}

func NewManager(db *sql.DB) *CronManager {
	cronModel := mcron.NewCron(db)
	cronController := cron.NewCronController(db)
	m := &CronManager{cronController, cronModel}
	m.init()
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

