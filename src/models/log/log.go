package log

import (
	"errors"
	"database/sql"
)
var (
	updateFailError = errors.New("更新失败")
)
// log 表实体类 entry
/**
 `id` int(11) NOT NULL AUTO_INCREMENT,
 `cron_id` int(11) NOT NULL DEFAULT '0',
 `time` bigint(20) NOT NULL COMMENT '命令运行的时间',
 `output` longtext NOT NULL COMMENT '执行命令输出',
 `use_time` bigint(20) NOT NULL COMMENT '执行命令耗时，单位为毫秒',
 `run_server` varchar(1024) NOT NULL DEFAULT '' COMMENT '该命令在那个节点上被执行（服务器）'
*/

type LogEntity struct {
	Id int64         `json:"id"`
	CronId int64     `json:"cron_id"`
	Time int64       `json:"time"`
	Output string    `json:"output"`
	UseTime int64    `json:"use_time"`
	DispatchServer string `json:"dispatch_server"`
	RunServer string `json:"run_server"`
}
type ILog interface {
	GetList(cronId int64, search string, dispatchServer, runServer string, page int64, limit int64) ([]*LogEntity, int64, error)
	Get(rid int64) (*LogEntity, error)
	Add(cronId int64, output string, useTime int64, dispatchTime int64, dispatchServer, runServer string, rtime int64) (*LogEntity, error)
	Delete(id int64) (*LogEntity, error)
	DeleteFormCronId(cronId int64) ([]*LogEntity, error)
}
type Middleware func(ILog) ILog

func NewLog(handler *sql.DB) ILog {
	var db ILog
	{
		db = newDbLog(handler)
	}
	return db
}
