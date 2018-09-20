package manager

import (
	"github.com/emicklei/go-restful"
	"strconv"
	"gitlab.xunlei.cn/xllive/common/log"
	"models/cron"
	"fmt"
	"net/url"
)


// restful api，只支持post
// http post 添加定时任务
func (m *CronManager) addCron(request *restful.Request, w *restful.Response) {
	params, err := ParseForm(request)
	if err != nil || params == nil {
		m.outJson(w, HttpErrorParamInvalid, err.Error(), nil)
		return
	}
	if params.GetCronSet() == "" {
		m.outJson(w, HttpErrorParamCronSet, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		return
	}
	if len(params.GetCommand()) <= 0 {
		m.outJson(w, HttpErrorParamCommand, "参数错误", nil)
		return
	}

	st, err := params.GetStartTime()
	if err != nil {
		m.outJson(w, HttpErrorParamStartTime, "参数错误", err.Error())
		return
	}
	et, err := params.GetEndTime()
	if err != nil {
		m.outJson(w, HttpErrorParamEndTime, "参数错误", err.Error())
		return
	}
	// 添加到数据库
	id, err := m.cronModel.Add(params.Blame, params.GetCronSet(),
		params.Command, params.Remark, params.IsStop(),
		st, et, params.Mutex())
	if err != nil || id <= 0 {
		log.Errorf("addCron m.cronModel.Add fail, error=[%v]", err)
		m.outJson(w, HttpErrorCronModelAddFail, err.Error(), nil)
		return
	}
	cdata := &cron.CronEntity{
		Id:        id,
		CronSet:   params.CronSet,
		Command:   params.Command,
		Remark:    params.Remark,
		Stop:      params.IsStop(),
		StartTime: st,
		EndTime:   et,
		IsMutex:   params.Mutex(),
	}
	m.broadcast(EV_ADD, id)
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

// http://localhost:38001/cron/mutex/true/1656
func (m *CronManager) mutexTrue(request *restful.Request, w *restful.Response) {
	m.mutex(request, w, true)
}

// http://localhost:38001/cron/mutex/false/1656
func (m *CronManager) mutexFalse(request *restful.Request, w *restful.Response) {
	m.mutex(request, w, false)
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
	m.broadcast(EV_DELETE, id)
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
	params, err := ParseForm(request)
	if err != nil || params == nil {
		m.outJson(w, HttpErrorParamInvalid, err.Error(), nil)
		return
	}
	//cronSet      := request.QueryParameter("cron_set")
	//command      := request.QueryParameter("command")
	//remark       := request.QueryParameter("remark")
	//stop         := request.QueryParameter("stop")
	//strStartTime := request.QueryParameter("start_time")
	//strEndTime   := request.QueryParameter("end_time")
	//strIsMutex   := request.QueryParameter("is_mutex")

	if params.GetCronSet() == "" {
		m.outJson(w, HttpErrorParamCronSet, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		return
	}
	if len(params.GetCommand()) <= 0 {
		m.outJson(w, HttpErrorParamCommand, "参数错误", nil)
		return
	}

	st, err := params.GetStartTime()
	if err != nil {
		m.outJson(w, HttpErrorParamStartTime, "参数错误", err.Error())
		return
	}
	et, err := params.GetEndTime()
	if err != nil {
		m.outJson(w, HttpErrorParamEndTime, "参数错误", err.Error())
		return
	}

	err = m.cronModel.Update(id, params.GetCronSet(),
		params.GetCommand(), params.GetRemark(),
		params.IsStop(), st, et, params.Mutex(), params.Blame)
	if err != nil {
		m.outJson(w, HttpErrorCronModelUpdateFail, "更新失败", err.Error())
		return
	}
	m.broadcast(EV_UPDATE, id)
	m.cronController.Update(id, params.GetCronSet(), params.GetCommand(),
		params.GetRemark(), params.IsStop(),
		st, et, params.Mutex(), params.Blame)
	m.outJson(w, HttpSuccess, "success", nil)
}

// http://localhost:38001/cron/list
func (m *CronManager) cronList(request *restful.Request, w *restful.Response) {
	data := m.cronController.GetList()

	///cron/list?stop=&mutex=&timeout=&keyword=
	stop := request.QueryParameter("stop")
	fmt.Println("stop", stop)
	mutex := request.QueryParameter("mutex")
	fmt.Println("mutex", mutex)
	timeout := request.QueryParameter("timeout")
	fmt.Println("timeout", timeout)
	keyword := request.QueryParameter("keyword")
	keyword, _ = url.QueryUnescape(keyword)
	fmt.Println("keyword", keyword)


	data = searchStop(data, stop)
	data = searchMutex(data, mutex)
	data = searchTimeout(data, timeout)
	data = searchKeyword(data, keyword)

	m.outJson(w, HttpSuccess, "success", data)
}

func (m *CronManager) cronInfo(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}
	d, err := m.cronController.Get(id)
	if err != nil {
		m.outJson(w, HttpErrorCronControllerGetFail, "查询失败", err.Error())
		return
	}
	m.outJson(w, HttpSuccess, "success", d.Clone())
}

//cronRun
func (m *CronManager) cronRun(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}
	stimeout := request.PathParameter("timeout")
	timeout, err := strconv.ParseInt(string(stimeout), 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorTimeoutInvalid, "timeout设置错误：" + err.Error(), nil)
		return
	}
	res, processId, err := m.cronController.RunCommand(id, timeout)
	if err != nil {
		m.outJson(w, HttpErrorCronControllerRunCommandFail, fmt.Sprintf("运行失败，进程ID：%v, 返回错误："+err.Error(), processId), nil)
		return
	}
	m.outJson(w, HttpSuccess, "success", fmt.Sprintf("进程ID：%v, 返回：%v", processId, string(res)))
}

//cronKill
func (m *CronManager) cronKill(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}
	sprocess_id := request.PathParameter("process_id")
	process_id, err := strconv.ParseInt(string(sprocess_id), 10, 64)
	if err != nil || process_id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "process_id错误", nil)
		return
	}
	m.cronController.Kill(id, int(process_id))//id, timeout)
	m.broadcast(EV_KILL, id, process_id)
	m.outJson(w, HttpSuccess, "success", nil)

}