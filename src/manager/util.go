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
		log.Errorf("outJson json.Marshal fail, error=[%v]", err)
		return err
	}
	n, err := w.Write(raw)
	if err != nil {
		log.Errorf("outJson w.Write fail, error=[%v]", err)
		return err
	}
	if n != len(raw) {
		log.Errorf("outJson ErrSendNotComplete fail, error=[%v]", ErrSendNotComplete)
		return ErrSendNotComplete
	}
	log.Tracef("outJson success, json=[%+v]", string(raw))
	return nil
}
