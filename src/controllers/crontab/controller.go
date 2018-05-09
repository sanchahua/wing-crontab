package crontab

import (
	"models/cron"
	log "github.com/sirupsen/logrus"
	cronv2 "gopkg.in/robfig/cron.v2"
)

type CrontabController struct {
	handler *cronv2.Cron
}
type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	CronId int64    `json:"cron_id"`//runtime cron id
}

func (row *CronEntity) Run() {
	if row.Stop {
		// 外部注入，停止执行定时任务支持
		log.Debugf("%+v was stop", row.Id)
		return
	}

	//roundbin to target server and run command
}


func NewCrontabController() *CrontabController {
	c := &CrontabController{
		handler: cronv2.New(),
	}
	return c
}
func (c *CrontabController) Start() {
	c.handler.Start()
}
func (c *CrontabController) Stop() {
	c.handler.Stop()
}

func (c *CrontabController) OnCrontabChange(event int, entity *cron.CronEntity) {
	switch event {
	case cron.EVENT_ADD:
		log.Infof("add crontab: %+v", entity)

	case cron.EVENT_DELETE:
		log.Infof("delete crontab: %+v", entity)
	case cron.EVENT_START:
		log.Infof("start crontab: %+v", entity)
	case cron.EVENT_STOP:
		log.Infof("stop crontab: %+v", entity)
	case cron.EVENT_UPDATE:
		log.Infof("update crontab: %+v", entity)
	}
}
