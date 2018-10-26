package manager

import (
	"github.com/emicklei/go-restful"
)

func (m *CronManager) services(request *restful.Request, w *restful.Response) {
	sv, _ := m.service.GetServices()
	m.outJson(w, HttpSuccess, "ok", sv)
}

//nodeOffline
func (m *CronManager) nodeOffline(request *restful.Request, w *restful.Response) {
	//send event offline
	m.service.Offline(true)
	m.outJson(w, HttpSuccess, "ok", nil)
}

//nodeOnline
func (m *CronManager) nodeOnline(request *restful.Request, w *restful.Response) {
	// send event online
	m.service.Offline(false)
	m.outJson(w, HttpSuccess, "ok", nil)
}