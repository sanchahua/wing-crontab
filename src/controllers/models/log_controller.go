package models

import (
	"app"
	"database/sql"
	mlog "models/log"
	_ "github.com/go-sql-driver/mysql"
)

type LogController struct {
	db mlog.ILog
	handler *sql.DB
}

func NewLogController(ctx *app.Context, handler *sql.DB) *LogController {
	//dataSource := fmt.Sprintf(
	//	"%s:%s@tcp(%s:%d)/%s?charset=%s",
	//	ctx.Config.MysqlUser,//User,
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

	db := mlog.NewLog(handler)
	return &LogController{db:db, handler:handler}
}

// 获取所有的定时任务列表
func (db *LogController) GetList(cronId int64, search string, runServer string, page int64, limit int64) ([]*mlog.LogEntity, int64, error) {
	return db.db.GetList(cronId, search, runServer, page, limit)
}

// 根据指定id查询行
func (db *LogController) Get(rid int64) (*mlog.LogEntity, error) {
	return db.db.Get(rid)
}

func (db *LogController) Add(cronId int64, output string, useTime int64, runServer string) (*mlog.LogEntity, error) {
	return db.db.Add(cronId, output, useTime, runServer)
}


func (db *LogController) Delete(id int64) (*mlog.LogEntity, error) {
	return db.db.Delete(id)
}

func (db *LogController) DeleteFormCronId(cronId int64) ([]*mlog.LogEntity, error) {
	return db.db.DeleteFormCronId(cronId)
}

