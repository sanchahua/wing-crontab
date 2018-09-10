package manager

import (
	"github.com/emicklei/go-restful"
	"strconv"
	"math"
	"time"
	time2 "library/time"
	"fmt"
	"net/url"
)

// http://localhost:38001/log/list/{cron_id}/{page}/{limit}
// http://localhost:38001/log/list/0/0/0
func (m *CronManager) logs(request *restful.Request, w *restful.Response) {
	scronId := request.PathParameter("cron_id")
	searchFail := request.PathParameter("search_fail")
	spage   := request.PathParameter("page")
	slimit  := request.PathParameter("limit")

	keyword := request.QueryParameter("keyword")
	keyword, _ = url.QueryUnescape(keyword)
	fmt.Println("keyword", keyword)

	cronId, _ := strconv.ParseInt(scronId, 10, 64)
	page, _   := strconv.ParseInt(spage, 10, 64)
	limit, _  := strconv.ParseInt(slimit, 10, 64)

	data, total, page, limit, _ := m.logModel.GetList(cronId, searchFail == "1", page, limit, keyword)

	totalPage := int64(math.Ceil(float64(total/limit)))
	nextPage := page+1
	if nextPage > totalPage {
		nextPage = 1
	}

	for _, row := range data {
		// 如果是是开始状态，并且当前进程还在运行，计算出距离开始执行的时间
		// 填充到UseTime
		if row.State == "start" && m.cronController.ProcessIsRunning(row.CronId, row.ProcessId) {
			row.UseTime = (time.Now().Unix() - time2.StrToTime(row.StartTime))*1000
		}
	}

	m.outJson(w, HttpSuccess, "success", map[string]interface{}{
		"data":      data,
		"total":     total,
		"totalPage": totalPage,
		"nextPage":  nextPage,
		"page":      page,
		"limit":     limit,
	})
}

//cronLogDetail
func (m *CronManager) cronLogDetail(request *restful.Request, w *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseInt(string(sid), 10, 64)
	if err != nil || id <= 0 {
		m.outJson(w, HttpErrorIdInvalid, "id错误", nil)
		return
	}

	data, _ := m.logModel.Get(id)
	m.outJson(w, HttpSuccess, "success", data)
}