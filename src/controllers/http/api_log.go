package http

import (
	"strconv"
	log "github.com/sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"models/cron"
	mlog "models/log"
	"library/http"
)

type LogApi struct {
	cron cron.ICron
	log mlog.ILog
	server *http.HttpServer
}

func NewLogApi(
	log mlog.ILog) *LogApi {

	h  := &LogApi{log:log}
	return h
}


func (server *LogApi) logs(request *restful.Request, w *restful.Response) {
	strCronId      := request.QueryParameter("cron_id")
	cronId, _      := strconv.ParseInt(strCronId, 10, 64)
	search         := request.QueryParameter("search")
	dispatchServer := request.QueryParameter("dispatch_server")
	runServer      := request.QueryParameter("run_server")
	strPage        := request.QueryParameter("page")
	page, _        := strconv.ParseInt(strPage, 10, 64)
	strLimit       := request.QueryParameter("limit")
	limit, _       := strconv.ParseInt(strLimit, 10, 64)
	//cronId int64, search string, runServer string, page int64, limit int64

	list, num, err := server.log.GetList(cronId, search, dispatchServer, runServer, page, limit)
	if err != nil {
		data, _ := output(200, httpErrors[200], err)
		w.Write(data)
		return
	}
	data, err := output(200, httpErrors[200], map[string]interface{}{"list":list, "total":num})
	log.Debugf("josn: %v, %v", list, data)
	if err == nil {
		w.Write(data)
	} else {
		w.Write(systemError("编码json发生错误"))
	}
}