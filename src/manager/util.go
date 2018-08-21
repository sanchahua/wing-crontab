package manager

import (
	"encoding/json"
	"github.com/emicklei/go-restful"
	"gitlab.xunlei.cn/xllive/common/log"
)

func (m *CronManager) outJson(w *restful.Response, code int, msg string, data interface{}) (error) {
	res := make(map[string] interface{})
	res["code"] = code
	res["message"] = msg
	res["data"] = data
	raw, err := json.Marshal(res)
	if err != nil {
		return err
	}
	n, err := w.Write(raw)
	if err != nil {
		return err
	}
	if n != len(raw) {
		return ErrSendNotComplete
	}
	log.Tracef("outJson success, json=[%+v]", string(raw))
	return nil
}
