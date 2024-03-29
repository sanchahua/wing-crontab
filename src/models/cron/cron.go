package cron

import (
	"errors"
	"database/sql"
)
var (
	ErrNoRowsChange = errors.New("no rows change")
	ErrIdInvalid = errors.New("id invalid")
	ErrNoRowsAffected =  errors.New("no rows affected")
	ErrCronSetInvalid = errors.New("cronSet invalid")
	ErrCommandInvalid = errors.New("command invalid")
	ErrEndTimeInvalid = errors.New("endTime invalid")
	ErrUserIdInvalid = errors.New("userid invalid")
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
	Id int64         `json:"id"`
	CronSet string   `json:"cron_set"`
	Command string   `json:"command"`
	Remark string    `json:"remark"`
	Stop bool        `json:"stop"`
	StartTime string `json:"start_time"`
	EndTime string   `json:"end_time"`
	IsMutex bool     `json:"is_mutex"`
	Blame int64      `json:"blame"`
	UserId int64     `json:"userid"`//添加者
}

func NewCron(handler *sql.DB) *DbCron {
	return newDbCron(handler)
}