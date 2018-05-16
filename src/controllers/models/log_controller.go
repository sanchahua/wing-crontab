package models

import (
	"app"
	"database/sql"
	mlog "models/log"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	log "github.com/sirupsen/logrus"
	"time"
)

type LogController struct {
	db mlog.ILog
	handler *sql.DB
	addChannel chan *addItem
}

type addItem struct {
	cronId int64
	output string
	useTime int64
	dispatchTime int64
	dispatchServer string
	runServer string
	rtime int64
}
const addChannelLen = 10000
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
	c := &LogController{db:db, handler:handler, addChannel:make(chan *addItem, addChannelLen)}
	cpu := runtime.NumCPU()
	for i := 0; i < cpu; i++ {
		go c.asyncAdd()
	}
	return c
}

func (db *LogController) asyncAdd() {
	for {
		select {
		case data, ok := <- db.addChannel:
			if !ok {
				return
			}
			db.db.Add(data.cronId, data.output, data.useTime, data.dispatchTime, data.dispatchServer, data.runServer, data.rtime)
		}
	}
}

func (db *LogController) AsyncAdd(cronId int64, output string, useTime int64, dispatchTime int64, dispatchServer, runServer string, rtime int64) {
	for {
		if len(db.addChannel) < cap(db.addChannel) {
			break
		}
		db.db.Add(cronId, output, useTime, dispatchTime, dispatchServer, runServer, rtime)
		log.Warnf("AsyncAdd cache full, %v, %v", len(db.addChannel) , cap(db.addChannel))
		return
	}
	db.addChannel <- &addItem{
		cronId:cronId,
		output :output,
		useTime :useTime,
		dispatchServer:dispatchServer,
		runServer :runServer,
		rtime:time.Now().Unix(),
		dispatchTime:dispatchTime,
	}//db.db.Add(cronId, output, useTime, dispatchServer, runServer)
}

func (db *LogController) Add(cronId int64, output string, useTime int64, dispatchTime int64, dispatchServer, runServer string, rtime int64) (*mlog.LogEntity, error) {
	return db.db.Add(cronId, output, useTime, dispatchTime, dispatchServer, runServer, rtime)
}

// 获取所有的定时任务列表
func (db *LogController) GetList(cronId int64, search string, dispatchServer, runServer string, page int64, limit int64) ([]*mlog.LogEntity, int64, error) {
	return db.db.GetList(cronId, search, dispatchServer, runServer, page, limit)
}

// 根据指定id查询行
func (db *LogController) Get(rid int64) (*mlog.LogEntity, error) {
	return db.db.Get(rid)
}

func (db *LogController) Delete(id int64) (*mlog.LogEntity, error) {
	return db.db.Delete(id)
}

func (db *LogController) DeleteFormCronId(cronId int64) ([]*mlog.LogEntity, error) {
	return db.db.DeleteFormCronId(cronId)
}

