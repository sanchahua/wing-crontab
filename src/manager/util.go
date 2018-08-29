package manager

import (
	"encoding/json"
	"github.com/emicklei/go-restful"
	"gitlab.xunlei.cn/xllive/common/log"
	"strings"
	"net/http"
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

func ContentTypeIsJson(header http.Header) bool {
	for key, value := range header {
		if strings.ToLower(key) == "content-type" {
			if len(value) > 0 {
				if strings.Index(value[0], "application/json") >= 0 {
					return true
				}
			}
			break
		}
	}
	return false
}

// 已兼容各种各样奇葩一般的表单提交方式
func ParseForm(request *restful.Request) (*httpParamsEntity, error) {
	raw := new(httpParamsEntity)
	if request.Request.Form == nil {
		request.Request.ParseMultipartForm(32 << 20)
	}
	// 兼容application/json从body读取json数据
	if ContentTypeIsJson(request.Request.Header) {
		err := request.ReadEntity(raw)
		if err != nil {
			if len(request.Request.Form) > 0 {
				for _, v := range request.Request.Form {
					if len(v) > 0 {
						err = json.Unmarshal([]byte(v[0]), &raw)
						if err != nil {
							return nil, err
						}
					}
				}
			}
			return nil, err
		}
		return raw, nil
	}
	var d = make(map[string]interface{})
	i := 0
	for k, v := range request.Request.Form {
		if 0 == i {
			i++
			err := json.Unmarshal([]byte(k), &raw)
			if err == nil {
				return raw, err
			}
		}
		if len(v) > 0 {
			d[k] = v[0]
		} else {
			d[k] = ""
		}
	}
	jd, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jd, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}