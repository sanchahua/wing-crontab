package http

import (
	"strconv"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"fmt"
	"models/cron"
	"library/http"
	"app"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

//查看数据库配置列表
//http://localhost:9990/cron/list
//停止指定定时任务
//http://localhost:9990/cron/stop/1
//开始指定定时任务
//http://localhost:9990/cron/start/1
//删除指定定时任务
//http://localhost:9990/cron/delete/1
//更新指定定时任务
//http://localhost:9990/cron/update/4?cronSet=*%20*%20*%20*%20*%20*&command=php%20-v&isMutex=1&remark=1234&lockLimit=60
//新增定时任务
//http://localhost:9990/cron/add?cronSet=0%20*\/1%20*%20*%20*%20*&command=php%20-v&isMutex=0&remark=&lockLimit=60
//curl -X POST http://localhost:9990/cron/add -d '{"cronSet":"*/1 * * * * *","command":"curl -X POST http://live.xunlei.com/","isMutex":0,"remark":"","lockLimit":60 }'
//强制解锁执行id
//http://localhost:9990/cron/unlock/1


type HttpServer struct {
	cron cron.ICron
	server *http.HttpServer
	db *sql.DB
	hooks []ChangeHook
}
const (
	EVENT_STOP   = 1
	EVENT_START  = 2
	EVENT_UPDATE = 3
	EVENT_ADD    = 4
	EVENT_DELETE = 5
)
type ChangeHook func(event int, row *cron.CronEntity)
type HttpControllerOption func(http *HttpServer)
func SetHook(f ChangeHook) HttpControllerOption {
	return func(http *HttpServer) {
		http.hooks = append(http.hooks, f)
	}
}
func NewHttpController(ctx *app.Context, opts ...HttpControllerOption) *HttpServer {

	//config, _ := app.GetMysqlConfig()
	dataSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s",
		ctx.Config.MysqlUser,
		ctx.Config.MysqlPassword,
		ctx.Config.MysqlHost,
		ctx.Config.MysqlPort,
		ctx.Config.MysqlDatabase,
		ctx.Config.MysqlCharset,
	)
	handler, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Panicf("链接数据库错误：%+v", err)
	}
	//设置最大空闲连接数
	handler.SetMaxIdleConns(8)
	//设置最大允许打开的连接
	handler.SetMaxOpenConns(8)
	cr := cron.NewCron(handler)
	h  := &HttpServer{cron:cr, db:handler, hooks:make([]ChangeHook, 0)}
	h.server = http.NewHttpServer(
		ctx.Config.HttpBindAddress,
		http.SetRoute("GET",  "/cron/list",        h.list),
		http.SetRoute("GET",  "/cron/stop/{id}",   h.stop),
		http.SetRoute("GET",  "/cron/start/{id}",  h.start),
		http.SetRoute("GET",  "/cron/delete/{id}", h.delete),
		http.SetRoute("POST", "/cron/update",      h.update),
		http.SetRoute("POST", "/cron/add",         h.add),
		//http.SetRoute("GET",  "/cron/unlock/{id}", h.unlock),
		//http.SetRoute("GET",  "/cron/lock/{id}",   h.lock),
	)

	for _, f := range opts {
		f(h)
	}
	return h
}

func (server *HttpServer) Start() {
	server.server.Start()
}

func (server *HttpServer) Close() {
	server.server.Close()
	server.db.Close()
}

func (server *HttpServer) firedHooks(event int, row *cron.CronEntity) {
	go func() {
		for _, f := range server.hooks {
			f(event, row)
		}
	}()
}

//http://localhost:9990/cron/list
func (server *HttpServer) list(request *restful.Request, w *restful.Response) {
	list, err := server.cron.GetList()
	if err == nil {
		data, _ := output(200, httpErrors[200], err)
		w.Write(data)
		return
	}
	data, err := output(200, httpErrors[200], list)
	log.Debugf("josn: %v, %v", list, data)
	if err == nil {
		w.Write(data)
	} else {
		w.Write(systemError("编码json发生错误"))
	}
}

// 停止定时任务
//http://localhost:9990/cron/stop/1
func (server *HttpServer) stop(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, _ := strconv.ParseInt(string(sid), 10, 64)
	// todo 更新定时任务
	// 更新db记录
	row, err := server.cron.Stop(id)
	if err == nil {
		server.firedHooks(EVENT_STOP, row)
	}
	log.Debugf("成功停止%d", id)
	out, _ := output(200, "ok", row)
	w.Write(out)
}

func (server *HttpServer) start(request *restful.Request, w *restful.Response)  {
	sid := request.PathParameter("id")
	id, _ := strconv.ParseInt(string(sid), 10, 64)
	// todo 更新定时任务
	// 更新db记录
	row, err := server.cron.Start(id)
	if err == nil {
		server.firedHooks(EVENT_START, row)
	}
	log.Debugf("成功开始%d", id)
	out, _ := output(200, "ok", row)
	w.Write(out)

}

// restful api 删除定时任务
// curl -X DELETE http://localhost:9990/cron/delete/1  这里的1是数据库id
//http://localhost:9990/cron/delete/1
func (server *HttpServer) delete(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, _ := strconv.ParseInt(string(sid), 10, 64)
	log.Debugf("====删除===================%d", id)
	row, err := server.cron.Delete(id)
	if row == nil {
		out, _ := output(200, fmt.Sprintf("%v does not exists", id), nil)
		w.Write(out)
	} else {
		out, _ := output(200, "ok", row)
		w.Write(out)
	}
	if err == nil {
		server.firedHooks(EVENT_DELETE, row)
	}
}

// 更新定时任务
//http://localhost:9990/cron/update/1
func (server *HttpServer) update(request *restful.Request, w *restful.Response) {
	sid       := request.QueryParameter("id")
	id, _     := strconv.ParseInt(string(sid), 10, 64)
	cronSet   := request.QueryParameter("cronSet")
	command   := request.QueryParameter("command")
	remark    := request.QueryParameter("remark")
	stop      := request.QueryParameter("stop")

	if len(cronSet) <= 0 || len(command) <= 0 || len(remark) <= 0 {
		out, _ := output(201, "参数错误", nil)
		w.Write(out)
		return
	}
	res := strings.Split(cronSet, " ")
	if len(res) != 6 {
		out, _ := output(201, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		w.Write(out)
		return
	}
	row, err := server.cron.Update(id, cronSet, command, remark, stop == "1")
	log.Debugf("成功更新%d", id)
	out, _ := output(200, "ok", row)
	w.Write(out)
	if err == nil {
		server.firedHooks(EVENT_UPDATE, row)
	}
}

// 添加定时任务
// http://localhost:9990/cron/add?cronSet=0%20*/1%20*%20*%20*%20*&command=php%20-v&isMutex=0&remark=
func (server *HttpServer) add(request *restful.Request, w *restful.Response) {
	cronSet := request.QueryParameter("cronSet")
	command := request.QueryParameter("command")
	remark := request.QueryParameter("remark")
	stop := request.QueryParameter("stop")

	if len(cronSet) <= 0 || len(command) <= 0 {
		out, _ := output(201, "参数错误", nil)
		w.Write(out)
		return
	}
	res := strings.Split(cronSet, " ")
	if len(res) != 6 {
		out, _ := output(201, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		w.Write(out)
		return
	}

	row, err := server.cron.Add(cronSet, command, remark, stop == "1")
	out, _ := output(200, httpErrors[200], row)
	w.Write(out)
	if err == nil {
		server.firedHooks(EVENT_ADD, row)
	}
}

//强制解锁某定时任务
//http://localhost:9990/cron/unlock/1
//func (server *HttpServer) unlock(request *restful.Request, w *restful.Response) {
//	sid       := request.QueryParameter("id")
//	id, _     := strconv.ParseInt(string(sid), 10, 64)
//	out, _ := output(201, httpErrors[201], nil)
//	if id <= 0 {
//		w.Write(out)
//		return
//	}
//
//	 log.Debugf("强制解锁%d", id)
//	 out, _ = output(200, "ok", nil)
//	 w.Write(out)
//}
//
////http://localhost:9990/cron/lock
//func (server *HttpServer) lock(request *restful.Request, w *restful.Response) {
//	sid       := request.QueryParameter("id")
//	id, _     := strconv.ParseInt(string(sid), 10, 64)
//	log.Debug("lock %v", id)
//	w.Write([]byte("ok"))
//}




