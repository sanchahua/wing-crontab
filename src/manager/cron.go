package manager

import (
	"github.com/emicklei/go-restful"
	"strings"
	"strconv"
	"gitlab.xunlei.cn/xllive/common/log"
	"models/cron"
)

// restful api，只支持post
// http post 添加定时任务
func (m *CronManager) addCron(request *restful.Request, w *restful.Response) {
	cronSet      := request.QueryParameter("cron_set")
	command      := request.QueryParameter("command")
	remark       := request.QueryParameter("remark")
	stop         := request.QueryParameter("stop")
	strStartTime := request.QueryParameter("start_time")
	strEndTime   := request.QueryParameter("end_time")
	strIsMutex   := request.QueryParameter("is_mutex")

	cronSet = strings.Trim(cronSet, " ")
	//res := strings.Split(cronSet, " ")
	if cronSet == "" {
		m.outJson(w, HttpErrorParamCronSet, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		return
	}
	command = strings.Trim(command, " ")
	if len(command) <= 0 {
		m.outJson(w, HttpErrorParamCommand, "参数错误", nil)
		return
	}

	isMutex := false
	if strIsMutex != "0" && strIsMutex != "" {
		isMutex = true
	}
	startTime, _ := strconv.ParseInt(strStartTime, 10, 64)
	endTime, _   := strconv.ParseInt(strEndTime, 10, 64)
	// 添加到数据库
	id, err      := m.cronModel.Add(cronSet, command, remark, stop == "1", startTime, endTime, isMutex)
	if err != nil || id <= 0 {
		log.Errorf("addCron m.cronModel.Add fail, error=[%v]", err)
		m.outJson(w, HttpErrorCronModelAddFail, err.Error(), nil)
		return
	}
	cdata := &cron.CronEntity{
		Id:        id,
		CronSet:   cronSet,
		Command:   command,
		Remark:    remark,
		Stop:      stop == "1",
		StartTime: startTime,
		EndTime:   endTime,
		IsMutex:   isMutex,
	}
	// 添加定时任务到定时任务管理器
	m.cronController.Add(cdata)
	m.outJson(w, HttpSuccess, "success", cdata)
}

// http://localhost:38001/cron/start/1656
func (m *CronManager) startCron(request *restful.Request, w *restful.Response) {
	m.stop(request, w, false)
}

// http://localhost:38001/cron/stop/1656
func (m *CronManager) stopCron(request *restful.Request, w *restful.Response) {
	m.stop(request, w, true)
}

// http://localhost:38001/cron/delete/1656
func (m *CronManager) deleteCron(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}
	err = m.cronModel.Delete(id)
	if err != nil {
		m.outJson(w, HttpErrorCronModelDeleteFail, "删除失败", nil)
		return
	}
	m.cronController.Delete(id)
	m.outJson(w, HttpSuccess, "success", nil)
}

// http://localhost:38001/cron/update/1656
func (m *CronManager) updateCron(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}

	cronSet      := request.QueryParameter("cron_set")
	command      := request.QueryParameter("command")
	remark       := request.QueryParameter("remark")
	stop         := request.QueryParameter("stop")
	strStartTime := request.QueryParameter("start_time")
	strEndTime   := request.QueryParameter("end_time")
	strIsMutex   := request.QueryParameter("is_mutex")

	cronSet = strings.Trim(cronSet, " ")
	//res := strings.Split(cronSet, " ")
	if cronSet == "" {
		m.outJson(w, HttpErrorParamCronSet, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		return
	}
	command = strings.Trim(command, " ")
	if len(command) <= 0 {
		m.outJson(w, HttpErrorParamCommand, "参数错误", nil)
		return
	}

	isMutex := false
	if strIsMutex != "0" && strIsMutex != "" {
		isMutex = true
	}
	startTime, _ := strconv.ParseInt(strStartTime, 10, 64)
	endTime, _   := strconv.ParseInt(strEndTime, 10, 64)

	err = m.cronModel.Update(id, cronSet, command, remark, stop == "1", startTime, endTime, isMutex)
	if err != nil {
		m.outJson(w, HttpErrorCronModelUpdateFail, "更新失败", nil)
		return
	}
	m.cronController.Update(id, cronSet, command, remark, stop == "1", startTime, endTime, isMutex)
	m.outJson(w, HttpSuccess, "success", nil)
}

// http://localhost:38001/cron/list
func (m *CronManager) cronList(request *restful.Request, w *restful.Response) {
	data := m.cronController.GetList()
	m.outJson(w, HttpSuccess, "success", data)
}