package http

import (
	"strconv"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"fmt"
	"models/cron"
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


type CronApi struct {
	cron cron.ICron
	hooks []ChangeCronHook
}

type ChangeCronHook func(event int, row *cron.CronEntity)
type CronApiOption func(http *CronApi)
func SetCronHook(f ChangeCronHook) CronApiOption {
	return func(http *CronApi) {
		http.hooks = append(http.hooks, f)
	}
}

func NewCronApi(
	cr cron.ICron,
	opts ...CronApiOption) *CronApi {
	h  := &CronApi{cron:cr, hooks:make([]ChangeCronHook, 0)}
	for _, f := range opts {
		f(h)
	}
	return h
}

func (server *CronApi) firedHooks(event int, row *cron.CronEntity) {
	go func() {
		for _, f := range server.hooks {
			f(event, row)
		}
	}()
}

//http://localhost:9990/cron/list
func (server *CronApi) list(request *restful.Request, w *restful.Response) {
	list, err := server.cron.GetList()
	if err != nil {
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
func (server *CronApi) stop(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, _ := strconv.ParseInt(string(sid), 10, 64)
	// todo 更新定时任务
	// 更新db记录
	row, err := server.cron.Stop(id)
	if err == nil {
		server.firedHooks(cron.EVENT_STOP, row)
	}
	log.Debugf("成功停止%d", id)
	out, _ := output(200, "ok", row)
	w.Write(out)
}

func (server *CronApi) start(request *restful.Request, w *restful.Response)  {
	sid := request.PathParameter("id")
	id, _ := strconv.ParseInt(string(sid), 10, 64)
	// todo 更新定时任务
	// 更新db记录
	row, err := server.cron.Start(id)
	if err == nil {
		server.firedHooks(cron.EVENT_START, row)
	}
	log.Debugf("成功开始%d", id)
	out, _ := output(200, "ok", row)
	w.Write(out)

}

// restful api 删除定时任务
// curl -X DELETE http://localhost:9990/cron/delete/1  这里的1是数据库id
//http://localhost:9990/cron/delete/1
func (server *CronApi) delete(request *restful.Request, w *restful.Response) {
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
		server.firedHooks(cron.EVENT_DELETE, row)
	}
}

// 更新定时任务
//http://localhost:9990/cron/update/1
func (server *CronApi) update(request *restful.Request, w *restful.Response) {
	sid       := request.QueryParameter("id")
	id, _     := strconv.ParseInt(string(sid), 10, 64)
	cronSet   := request.QueryParameter("cronSet")
	command   := request.QueryParameter("command")
	remark    := request.QueryParameter("remark")
	stop      := request.QueryParameter("stop")
	strStartTime := request.QueryParameter("start_time")
	strEndTime   := request.QueryParameter("end_time")
	strIsMutex   := request.QueryParameter("is_mutex")
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
	startTime, _ := strconv.ParseInt(strStartTime, 10, 64)
	endTime, _   := strconv.ParseInt(strEndTime, 10, 64)
	isMutex      := false
	if strIsMutex != "0" {
		isMutex = true
	}
	row, err     := server.cron.Update(id, cronSet, command, remark, stop == "1", startTime, endTime, isMutex)

	out, _ := output(200, "ok", row)
	w.Write(out)
	if err == nil {
		log.Debugf("成功更新%d", id)
		server.firedHooks(cron.EVENT_UPDATE, row)
	} else {
		log.Debugf("更新失败%d： %v", id, err)
	}
}

// 添加定时任务
// http://localhost:9990/cron/add?cronSet=0%20*/1%20*%20*%20*%20*&command=php%20-v&isMutex=0&remark=
func (server *CronApi) add(request *restful.Request, w *restful.Response) {
	cronSet   := request.QueryParameter("cronSet")
	command   := request.QueryParameter("command")
	remark    := request.QueryParameter("remark")
	stop         := request.QueryParameter("stop")
	strStartTime := request.QueryParameter("start_time")
	strEndTime   := request.QueryParameter("end_time")
	strIsMutex   := request.QueryParameter("is_mutex")


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
	isMutex      := false
	if strIsMutex != "0" {
		isMutex = true
	}
	startTime, _ := strconv.ParseInt(strStartTime, 10, 64)
	endTime, _   := strconv.ParseInt(strEndTime, 10, 64)
	row, err := server.cron.Add(cronSet, command, remark, stop == "1", startTime, endTime, isMutex)
	out, _ := output(200, httpErrors[200], row)
	w.Write(out)
	if err == nil {
		server.firedHooks(cron.EVENT_ADD, row)
	}
}