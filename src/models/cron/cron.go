package cron

import (
	"errors"
	"database/sql"
)
var (
	updateFailError = errors.New("更新失败")
)
// cron 表实体类 entry
/**
CREATE TABLE `cron` (
 `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
 `cron_set` varchar(128) NOT NULL DEFAULT '' COMMENT '定时任务配置，如：* * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周',
 `command` varchar(2048) NOT NULL DEFAULT '' COMMENT '定时任务执行的命令',
 `stop` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1停止执行，0非，0为默认值',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '定时任务的备注信息',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8
*/

type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	StartTime int64 `json:"start_time"`
	EndTime int64   `json:"end_time"`
}
const (
	EVENT_STOP   = 1
	EVENT_START  = 2
	EVENT_UPDATE = 3
	EVENT_ADD    = 4
	EVENT_DELETE = 5
)

type ICron interface {
	GetList() ([]*CronEntity, error)
	Get(id int64) (*CronEntity, error)
	Add(cronSet, command string, remark string, stop bool, startTime, endTime int64) (*CronEntity, error)
	Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64) (*CronEntity,error)
	Stop(id int64) (*CronEntity, error)
	Start(id int64) (*CronEntity, error)
	Delete(id int64) (*CronEntity, error)
}
type Middleware func(ICron) ICron


func NewCron(handler *sql.DB) ICron {
	var db ICron
	{
		db = newDbCron(handler)
		db = loggingMiddleware()(db)
	}
	return db
}