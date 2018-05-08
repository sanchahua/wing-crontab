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
 `is_mutex` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否需要互斥运行，1互斥，0非互斥，默认为0，非互斥，即可并发运行',
 `stop` tinyint(4) NOT NULL DEFAULT '0' COMMENT '1停止执行，0非，0为默认值',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '定时任务的备注信息',
 `lock_limit` int(11) NOT NULL DEFAULT '0' COMMENT '最长锁定时长，单位为秒',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8
*/

type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	IsMutex bool    `json:"is_mutex"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	LockLimit int64 `json:"lock_limit"`            //最长锁定时间，单位为秒，由数据库配置自定义指定
}

type ICron interface {
	GetList() ([]*CronEntity, error)
	Get(id int64) (*CronEntity, error)
	Add(cronSet, command string, isMutex bool, remark string, lockLimit int64, stop bool) (*CronEntity, error)
	Update(id int64, cronSet, command string, isMutex bool, remark string, lockLimit int64, stop bool) (*CronEntity,error)
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