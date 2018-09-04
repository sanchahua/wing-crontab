package manager

import (
	"github.com/emicklei/go-restful"
	"gitlab.xunlei.cn/xllive/common/log"
	"strings"
	"net/http"
	//"github.com/json-iterator/go"
	//"github.com/json-iterator/go/extra"
	//"github.com/json-iterator/go/extra"
	//"github.com/json-iterator/go"
	"encoding/json"
	cron2 "cron"
	"library/time"
	time2 "time"
	"fmt"
)

//func init() {
//	extra.RegisterFuzzyDecoders()
//}
//var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
	//log.Tracef("outJson success, json=[%+v]", string(raw))
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
			log.Errorf("ParseForm request.ReadEntity fail, error=[%v]", err)
			if len(request.Request.Form) > 0 {
				for _, v := range request.Request.Form {
					if len(v) > 0 {
						err = json.Unmarshal([]byte(v[0]), &raw)
						if err != nil {
							log.Errorf("ParseForm json.Unmarshal fail, error=[%v]", err)
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
			} else {
				log.Errorf("ParseForm json.Unmarshal fail, error=[%v]", err)
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
		log.Errorf("ParseForm json.Marshal fail, error=[%v]", err)
		return nil, err
	}
	err = json.Unmarshal(jd, &raw)
	if err != nil {
		log.Errorf("ParseForm json.Marshal fail, error=[%v]", err)
		return nil, err
	}
	return raw, nil
}

func searchStop(data []*cron2.CronEntity, stop string) []*cron2.CronEntity {
	var newData []*cron2.CronEntity = nil
	if stop == "1" {
		newData = make([]*cron2.CronEntity, 0)
		for _, v := range data {
			if v.Stop {
				newData = append(newData, v)
			}
		}
		return newData
	}

	if stop == "0" {
		newData = make([]*cron2.CronEntity, 0)
		for _, v := range data {
			if !v.Stop {
				newData = append(newData, v)
			}
		}
		return newData
	}

	return  data
}

func searchMutex(newData []*cron2.CronEntity, mutex string) []*cron2.CronEntity {
	var newData2 []*cron2.CronEntity = nil
	if mutex == "1" {
		newData2 = make([]*cron2.CronEntity, 0)
		for _, v := range newData {
			if v.IsMutex {
				newData2 = append(newData2, v)
			}
		}
		return newData2
	}

	if mutex == "0" {
		newData2 = make([]*cron2.CronEntity, 0)
		for _, v := range newData {
			if !v.IsMutex {
				newData2 = append(newData2, v)

			}
		}
		return newData2
	}
	return newData
}

func searchTimeout(newData2 []*cron2.CronEntity, timeout string) []*cron2.CronEntity {
	var newData3 []*cron2.CronEntity = nil//:= make([]*cron2.CronEntity, 0)

	if timeout == "1" {
		newData3 = make([]*cron2.CronEntity, 0)
		for _, v := range newData2 {
			st := time.StrToTime(v.StartTime)
			et := time.StrToTime(v.EndTime)
			n := time2.Now().Unix()
			if n < st || n > et {
				newData3 = append(newData3, v)
			}
		}
		return newData3
	}

	if timeout == "0" {
		newData3 = make([]*cron2.CronEntity, 0)
		for _, v := range newData2 {
			st := time.StrToTime(v.StartTime)
			et := time.StrToTime(v.EndTime)
			n := time2.Now().Unix()
			if st <= n && n <= et {
				newData3 = append(newData3, v)

			}
		}
		return newData3
	}
	return newData2
}

func searchKeyword(newData3 []*cron2.CronEntity, keyword string) []*cron2.CronEntity {
	var newData4 []*cron2.CronEntity = nil
	keyword = strings.Trim(keyword, " ")
	if keyword != "" {
		newData4 = make([]*cron2.CronEntity, 0)
		for _, v := range newData3 {
			// 根据command模糊查询
			// 查询id
			// 查询定时设置
			if strings.Index(v.Command, keyword) >= 0 ||
				fmt.Sprintf("%v", v.Id) == keyword ||
				keyword == v.CronSet {
				newData4 = append(newData4, v)
			}
		}
		return newData4
	}
	return  newData3
}