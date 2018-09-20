package manager

import (
	"github.com/emicklei/go-restful"
	"strconv"
)

func (m *CronManager) stop(request *restful.Request, w *restful.Response, stop bool) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}

	err = m.cronModel.Stop(id, stop)
	if err != nil {
		m.outJson(w, HttpErrorCronModelStopFalseFail, "m.cronModel.Stop fail", nil)
		return
	}
	err = m.cronController.Stop(id, stop)
	if err != nil {
		m.outJson(w, HttpErrorCronControllerStopFalseFail, "m.cronController.Stop fail", nil)
		return
	}
	if stop {
		m.broadcast(EV_STOP, id)
	} else {
		m.broadcast(EV_START, id)
	}
	m.outJson(w, HttpSuccess, "success", nil)
}

func (m *CronManager) mutex(request *restful.Request, w *restful.Response, mutex bool) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}

	err = m.cronModel.Mutex(id, mutex)
	if err != nil {
		m.outJson(w, HttpErrorCronModelMutexFalseFail, "m.cronModel.Mutex fail", nil)
		return
	}
	err = m.cronController.Mutex(id, mutex)
	if err != nil {
		m.outJson(w, HttpErrorCronControllerMutexFail, "m.cronController.Mutex fail", nil)
		return
	}
	if mutex {
		m.broadcast(EV_ENABLE_MUTEX, id)
	} else {
		m.broadcast(EV_DISABLE_MUTEX, id)
	}
	m.outJson(w, HttpSuccess, "success", nil)
}

