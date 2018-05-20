package crontab

import (
	log "github.com/sirupsen/logrus"
	cronv2 "gopkg.in/robfig/cron.v2"
	"fmt"
	"os"
	"models/cron"
)

type CronEntity struct {
	// 数据库的基本属性
	Id int64              `json:"id"`
	CronSet string        `json:"cron_set"`
	Command string        `json:"command"`
	Remark string         `json:"remark"`
	Stop bool             `json:"stop"`
	CronId cronv2.EntryID `json:"cron_id"`//runtime cron id
	StartTime int64       `json:"start_time"`
	EndTime int64         `json:"end_time"`
	IsMutex bool          `json:"is_mutex"`

	onwillrun OnWillRunFunc `json:"-"`
	filter IFilter          `json:"-"`
}
type CronEntityMiddleWare func(entity *CronEntity) IFilter

func newCronEntity(entity *cron.CronEntity, onwillrun OnWillRunFunc) *CronEntity {
	e := &CronEntity{
		Id:        entity.Id,      //int64        `json:"id"`
		CronSet:   entity.CronSet, // string  `json:"cron_set"`
		Command:   entity.Command, // string  `json:"command"`
		Remark:    entity.Remark,  //string   `json:"remark"`
		Stop:      entity.Stop,    //bool       `json:"stop"`
		CronId:    0,              //int64    `json:"cron_id"`
		onwillrun: onwillrun,//c.onwillrun,
		StartTime: entity.StartTime,
		EndTime:   entity.EndTime,
		IsMutex:   entity.IsMutex,
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
		log.Debugf("%+v was stop", row.Id)
		return
	}
	//roundbin to target server and run command
	row.onwillrun(row.Id, row.Command, row.IsMutex)
	fmt.Fprintf(os.Stderr, "\r\n########## only leader do this %+v\r\n\r\n", *row)
}
