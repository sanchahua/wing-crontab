package http

import (
	//log "github.com/sirupsen/logrus"
	"models/cron"
	mlog "models/log"
	"library/http"
	"app"
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
	log mlog.ILog
	server *http.HttpServer
}

func NewHttpController(
	ctx *app.Context,
	cr cron.ICron,
	log mlog.ILog,
	opts ...CronApiOption) *HttpServer {

	logApi := NewLogApi(log)
	cronApi := NewCronApi(cr, opts...)
	h  := &HttpServer{cron:cr, log:log}
	h.server = http.NewHttpServer(
		ctx.Config.HttpBindAddress,
		http.SetRoute("GET",  "/log/list",         logApi.logs),

		http.SetRoute("GET",  "/cron/list",        cronApi.list),
		http.SetRoute("GET",  "/cron/stop/{id}",   cronApi.stop),
		http.SetRoute("GET",  "/cron/start/{id}",  cronApi.start),
		http.SetRoute("GET",  "/cron/delete/{id}", cronApi.delete),
		http.SetRoute("POST", "/cron/update",      cronApi.update),
		http.SetRoute("POST", "/cron/add",         cronApi.add),
	)
	return h
}

func (server *HttpServer) Start() {
	server.server.Start()
}

func (server *HttpServer) Close() {
	server.server.Close()
}