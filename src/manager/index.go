package manager

import (
	"github.com/emicklei/go-restful"
	"time"
	"strconv"
	"gitlab.xunlei.cn/xllive/common/log"
	"encoding/json"
	"fmt"
)

// 首页 api
// 需要提供数据
// 1、定时任务数量
// 2、历史执行次数
// 3、今日执行次数
// 4、今日错误次数
func (m *CronManager) index(request *restful.Request, w *restful.Response) {
	//定时任务数量
	cronCount, _ := m.cronModel.GetCount()
	//查询历史执行次数
	historyCount, _ := m.statisticsModel.GetCount()
	//今日执行次数 今日错误次数
	dayCount, dayFailCount, _ := m.statisticsModel.GetDayCount(time.Now().Format("2006-01-02"))

	log.Tracef("index header: %+v", request.Request.Header)

	m.outJson(w, HttpSuccess, "success", map[string]int64{
		"cron_count": cronCount,
		"history_run_count": historyCount,
		"day_run_count": dayCount,
		"day_run_fail_count": dayFailCount,
	})
}

func (m *CronManager) charts(request *restful.Request, w *restful.Response) {
	daysStr := request.PathParameter("days")
	days, err := strconv.ParseInt(string(daysStr), 10, 64)
	if err != nil || days <= 0 {
		days = 7
	}
	charts, _ := m.statisticsModel.GetCharts(int(days))
	m.outJson(w, HttpSuccess, "success", charts)
}

func (m *CronManager) avgRunTimeCharts(request *restful.Request, w *restful.Response) {
	daysStr := request.PathParameter("days")
	days, err := strconv.ParseInt(string(daysStr), 10, 64)
	if err != nil || days <= 0 {
		days = 7
	}

	strids := request.QueryParameter("ids")
	var ids = make([]int64, 0)
	json.Unmarshal([]byte(strids), &ids)

	log.Tracef("strids=[%v], ids=[%v]", strids, ids)

	list := m.cronController.GetList()
	var legend = make([]string, 0)
	//for _, v := range list {
	//	legend = append(legend, v.Command)
	//}
	charts := make(map[string]interface{})
	var xAxis = make([]string, 0)
	for i := days; i >= 0; i-- {
		xAxis = append(xAxis, time.Unix(time.Now().Unix() - i*86400, 0).Format("2006-01-02"))
	}
	charts["legend"] = legend
	charts["xAxis"]  = xAxis

	var datas = make(map[string]map[int64]int64)
	for _, day := range xAxis {
		avt, _ := m.statisticsModel.GetAvgTime(day)
		datas[day] = avt
	}

	var series = make([]map[string]interface{}, 0)
	for _, v := range list {

		inArray := false
		for _, h := range ids {
			if h == v.Id {
				inArray = true
				break
			}
		}
		if !inArray {
			continue
		}

		var d = make(map[string]interface{})
		d["name"]  = fmt.Sprintf("[%v] %v", v.Id, v.Command)
		d["type"]  = "line"
		d["stack"] = "平均时长"
		var t = make([]int64, 0)
		for _, day := range xAxis {
			avt := datas[day]//, _ := m.statisticsModel.GetAvgTime(day)
			vt := int64(0)
			if vv, ok := avt[v.Id]; ok {
				vt = vv
			}
			t = append(t, vt)
		}
		d["data"] = t
		series = append(series, d)
	}

	charts["series"] = series
	m.outJson(w, HttpSuccess, "success", charts)
}
