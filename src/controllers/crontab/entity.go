package crontab

import (
	log "github.com/sirupsen/logrus"
	cronv2 "gopkg.in/robfig/cron.v2"
	"fmt"
	"os"
)

type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	CronId cronv2.EntryID    `json:"cron_id"`//runtime cron id
	onwillrun OnWillRunFunc `json:"-"`
	StartTime int64 `json:"start_time"`
	EndTime int64 `json:"end_time"`
	IsMutex bool  `json:"is_mutex"`
}
type CronEntityMiddleWare func(entity *CronEntity) IFilter

func (row *CronEntity) Run() {
	//start := time.Now()

	m := StopMiddleware()(row)
	m  = TimeMiddleware(m)(row)
	if m.Check() {
		// 外部注入，停止执行定时任务支持
		log.Debugf("%+v was stop", row.Id)
		return
	}

	//roundbin to target server and run command
	row.onwillrun(row.Id, row.Command, row.IsMutex)
	fmt.Fprintf(os.Stderr, "\r\n########## only leader do this %+v\r\n\r\n", *row)
}
