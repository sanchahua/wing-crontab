package manager

import (
	"github.com/emicklei/go-restful"
	"strconv"
	"math"
)

// http://localhost:38001/log/list/{cron_id}/{page}/{limit}
// http://localhost:38001/log/list/0/0/0
func (m *CronManager) logs(request *restful.Request, w *restful.Response) {
	scronId := request.PathParameter("cron_id")
	spage   := request.PathParameter("page")
	slimit  := request.PathParameter("limit")

	cronId, _ := strconv.ParseInt(scronId, 10, 64)
	page, _   := strconv.ParseInt(spage, 10, 64)
	limit, _  := strconv.ParseInt(slimit, 10, 64)

	data, total, page, limit, _ := m.logModel.GetList(cronId, page, limit)

	totalPage := int64(math.Ceil(float64(total/limit)))
	nextPage := page+1
	if nextPage > totalPage {
		nextPage = 1
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
