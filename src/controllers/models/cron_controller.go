package models

import (
	"database/sql"
	"models/cron"
	"app"
)

type CronController struct {
	cr cron.ICron
	handler *sql.DB
}

func NewCronController(ctx *app.Context, handler *sql.DB) *CronController {
	//dataSource := fmt.Sprintf(
	//	"%s:%s@tcp(%s:%d)/%s?charset=%s",
	//	ctx.Config.MysqlUser,
	//	ctx.Config.MysqlPassword,
	//	ctx.Config.MysqlHost,
	//	ctx.Config.MysqlPort,
	//	ctx.Config.MysqlDatabase,
	//	ctx.Config.MysqlCharset,
	//)
	//handler, err := sql.Open("mysql", dataSource)
	//if err != nil {
	//	log.Panicf("链接数据库错误：%+v", err)
	//}
	////设置最大空闲连接数
	//handler.SetMaxIdleConns(8)
	////设置最大允许打开的连接
	//handler.SetMaxOpenConns(8)
	cr := cron.NewCron(handler)
	return &CronController{cr:cr,handler:handler}
}

func (db *CronController) Close() {
	//db.handler.Close()
}

// 获取所有的定时任务列表
func (db *CronController) GetList() ([]*cron.CronEntity, error) {
	return db.cr.GetList()
}

// 根据指定id查询行
func (db *CronController) Get(pid int64) (*cron.CronEntity, error) {
	return db.cr.Get(pid)
}

func (db *CronController) Add(cronSet, command string, remark string, stop bool, startTime, endTime int64) (*cron.CronEntity, error) {
	return db.cr.Add(cronSet, command, remark, stop, startTime, endTime)
}

func (db *CronController) Update(id int64, cronSet, command string, remark string, stop bool, startTime, endTime int64) (*cron.CronEntity,error) {
	return db.cr.Update(id, cronSet, command, remark, stop, startTime, endTime)
}

func (db *CronController) Stop(id int64) (*cron.CronEntity, error) {
	return db.cr.Stop(id)
}

func (db *CronController) Start(id int64) (*cron.CronEntity, error) {
	return db.cr.Start(id)
}

func (db *CronController) Delete(id int64) (*cron.CronEntity, error) {
	return db.cr.Delete(id)
}

