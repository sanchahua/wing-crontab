package manager

import (
	"github.com/emicklei/go-restful"
	"time"
	"strconv"
	"gitlab.xunlei.cn/xllive/common/log"
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
