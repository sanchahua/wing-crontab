package crontab

import (
	"models/cron"
	log "github.com/sirupsen/logrus"
)

type CrontabController struct {

}

func NewCrontabController() *CrontabController {
	c := &CrontabController{}
	return c
}
func (c *CrontabController) Start() {}
func (c *CrontabController) Stop() {}

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
