package manager

import (
	"github.com/emicklei/go-restful"
	"strconv"
	"gitlab.xunlei.cn/xllive/common/log"
	"encoding/json"
	"fmt"
)

func (m *CronManager) services(request *restful.Request, w *restful.Response) {
	sv, _ := m.service.GetServices()
	m.outJson(w, HttpSuccess, "ok", sv)
}

//nodeOffline
func (m *CronManager) nodeOffline(request *restful.Request, w *restful.Response) {
	//send event offline
	strServiceId := request.PathParameter("id")
	serviceId, err := strconv.ParseInt(strServiceId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorServiceIdInvalid, "service id invalid", err)
		return
	}
	m.service.SetOffline(serviceId, true)
	m.service.UpdateOffline(serviceId, 1)
	m.broadcastService(EV_OFFLINE, serviceId, 1)
	m.outJson(w, HttpSuccess, "ok", nil)
}

//nodeOnline
func (m *CronManager) nodeOnline(request *restful.Request, w *restful.Response) {
	// send event online
	strServiceId := request.PathParameter("id")
	serviceId, err := strconv.ParseInt(strServiceId, 10, 64)
	if err != nil {
		m.outJson(w, HttpErrorServiceIdInvalid, "service id invalid", err)
		return
	}
	m.service.SetOffline(serviceId, false)
	m.service.UpdateOffline(serviceId, 0)
	m.broadcastService(EV_OFFLINE, serviceId, 0)
	m.outJson(w, HttpSuccess, "ok", nil)
}

//func (m *CronManager) ServiceKeep(id int64) {
//	log.Infof("send service keep: %+v, EV_LEADER=%v", id, EV_LEADER)
//	m.broadcastService(EV_LEADER, m.service.ID)
//}

func (m *CronManager) broadcastService(ev, id int64, p...int64) {
	// 查询服务列表 逐个push redis队列广播通知数据变化
	services, err := m.service.GetServices()
	if err != nil {
		log.Errorf("broadcastService m.service.GetServices fail, error=[%v]", err)
		return
	}
	log.Tracef("broadcastService ev=[%v], id=[%v], p=[%v], serviceId=[%v]", ev, id, p, m.serviceId)
	for _, sv := range services {
		var data []byte
		var err error

		sendData := make([]int64, 0)
		sendData = append(sendData, ev)
		sendData = append(sendData, id)
		sendData = append(sendData, p...)


		//if ev == EV_KILL {
		data, err = json.Marshal(sendData)
		//} else {
		//	data, err = json.Marshal([]int64{ev, id})
		//}
		if err != nil {
			log.Errorf("broadcastOffline json.Marshal fail, error=[%v]", err)
			continue
		}

		watch := fmt.Sprintf("xcrontab/watch/event/%v", sv.ID)
		//这里还需要一个线程，watch定时任务的增删改查，用来改变自身的配置
		log.Tracef("push [%v] to [%v]", string(data), watch)
		err = m.redis.RPush(watch, string(data)).Err()
		if err != nil {
			log.Errorf("broadcastOffline m.redis.RPush fail, error=[%v]", err)
		}
	}
}