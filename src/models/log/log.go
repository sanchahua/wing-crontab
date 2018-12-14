package log

import (
	"database/sql"
	"errors"
)
//var (
//	updateFailError = errors.New("更新失败")
//)
// log 表实体类 entry
/**
CREATE TABLE `log` (
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0' COMMENT '定时任务id',
 `start_time` datetime NOT NULL COMMENT '命令开始执行的时间',
 `output` longtext NOT NULL COMMENT '执行命令输出',
 `use_time` bigint(20) NOT NULL COMMENT '执行命令耗时，单位为毫秒',
 `remark` varchar(1024) NOT NULL DEFAULT '' COMMENT '备注',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=706891 DEFAULT CHARSET=utf8
*/

type LogEntity struct {
	Id        int64       `json:"id"`
	DispatchServer int64  `json:"dispatch_server"`
	RunServer int64       `json:"run_server"`
	CronId    int64       `json:"cron_id"`
	ProcessId int         `json:"process_id"`
	StartTime string      `json:"start_time"`
	Output    string      `json:"output"`
	UseTime   int64       `json:"use_time"`
	Remark    string      `json:"remark"`
	State     string      `json:"state"`
	DispatchServerName string  `json:"dispatch_server_name"`
	RunServerName string       `json:"run_server_name"`
}

var (
	ErrIdInvalid = errors.New("id invalid")
	ErrNoRowsAffected =  errors.New("no rows affected")
	ErrorStartTimeEmpty = errors.New("starttime is empty")
)

func NewLog(handler *sql.DB) *DbLog {
	return newDbLog(handler)
}
