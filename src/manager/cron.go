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
	cronSet      := request.QueryParameter("cronSet")
	command      := request.QueryParameter("command")
	remark       := request.QueryParameter("remark")
	stop         := request.QueryParameter("stop")
	strStartTime := request.QueryParameter("start_time")
	strEndTime   := request.QueryParameter("end_time")
	strIsMutex   := request.QueryParameter("is_mutex")

	cronSet = strings.Trim(cronSet, " ")
	res := strings.Split(cronSet, " ")
	if len(res) != 6 {
		err := m.outJson(w, HttpErrorParamCronSet, "定时任务设置错误，格式为（秒 分 时 日 月 周），如： * * * * * *", nil)
		if err != nil {
			log.Errorf("addCron m.outJson fail, error=[%v]", err)
		}
		return
	}
	command = strings.Trim(command, " ")
	if len(command) <= 0 {
		err := m.outJson(w, HttpErrorParamCommand, "参数错误", nil)
		if err != nil {
			log.Errorf("addCron m.outJson fail, error=[%v]", err)
		}
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
		err = m.outJson(w, HttpErrorCronModelFail, err.Error(), nil)
		if err != nil {
			log.Errorf("addCron m.outJson fail, error=[%v]", err)
		}
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
	m.cronController.StopCron()
	m.cronController.Add(cdata)
	m.cronController.StartCron()
	err = m.outJson(w, HttpSuccess, "success", cdata)
	if err != nil {
		log.Errorf("addCron m.outJson fail, error=[%v]", err)
	}
}
