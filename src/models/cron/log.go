package cron

import (
	log "github.com/sirupsen/logrus"
	"time"
)
//日志中间件实现

type Logger struct {
	next ICron
}

func loggingMiddleware() Middleware {
	return func(next ICron) ICron {
		return Logger{next}
	}
}

// 获取所有的定时任务列表
func (db Logger) GetList() ([]*CronEntity, error) {
	log.Infof("GetList() was called")
	start := time.Now()
	d, e := db.next.GetList()
	log.Infof("GetList use time: %+v", time.Since(start))
	return d, e
}

// 根据指定id查询行
func (db Logger) Get(pid int64) (*CronEntity, error) {
	log.Infof("Get(%v) was called", pid)
	start := time.Now()
	d, e := db.next.Get(pid)
	log.Infof("Get use time: %+v", time.Since(start))
	return d, e
}

func (db Logger) Add(cronSet, command string, remark string, stop bool) (*CronEntity, error) {
	log.Infof("Add(\"%v\", \"%v\", \"%v\", %v) was called", cronSet, command, remark, stop)
	start := time.Now()
	d, e := db.next.Add(cronSet, command, remark, stop)
	log.Infof("Add use time: %+v", time.Since(start))
	return d, e
}

func (db Logger) Update(id int64, cronSet, command string, remark string, stop bool) (*CronEntity,error) {
	log.Infof("Update(%v, \"%v\", \"%v\", \"%v\", %v) was called", id, cronSet, command, remark, stop)
	start := time.Now()
	d, e := db.next.Update(id, cronSet, command, remark, stop)
	log.Infof("Update use time: %+v", time.Since(start))
	return d, e
}

func (db Logger) Stop(id int64) (*CronEntity, error) {
	log.Infof("Stop(%v) was called", id)
	start := time.Now()
	d, e := db.next.Stop(id)
	log.Infof("Stop use time: %+v", time.Since(start))
	return d, e
}


func (db Logger) Start(id int64) (*CronEntity, error) {
	log.Infof("Start(%v) was called", id)
	start := time.Now()
	d, e := db.next.Start(id)
	log.Infof("Start use time: %+v", time.Since(start))
	return d, e
}

func (db Logger) Delete(id int64) (*CronEntity, error) {
	log.Infof("Delete(%v) was called", id)
	start := time.Now()
	d, e := db.next.Delete(id)
	log.Infof("Delete use time: %+v", time.Since(start))
	return d, e
}
