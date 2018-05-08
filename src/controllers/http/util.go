package http

import (
	"net/http"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

var httpErrors map[int] string = map[int] string {
	200 : "ok",
	201 : "request param error",
	202 : "error happened",
	203 : "method does not support",
}
func systemError(err string) []byte {
	return []byte(fmt.Sprintf("{\"code\":101, \"message\":\"%s\", \"data\":\"\"}", err))
}
func output(code int, msg string, data interface{}) ([]byte, error) {
	res := make(map[string] interface{})
	res["code"] = code
	res["message"] = msg
	res["data"] = data
	return json.Marshal(res)
}

func ParseForm(req *http.Request) {
	req.ParseForm()
	//if strings.ToLower(req.Header.Get("Content-Type")) == "application/json" {
	// 处理传统意义上表单的参数，这里添加body内传输的json解析支持
	// 解析后的值默认追加到表单内部
	var data map[string]interface{}
		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			log.Errorf("%v", err)
		} else {
			for k, dv := range data {
				_, ok := req.Form[k]
				if !ok {
					req.Form[k] = []string{fmt.Sprintf("%v", dv)}
				} else {
					req.Form[k] = append(req.Form[k], fmt.Sprintf("%v", dv))
				}
			}
		}
	//}
	for k, v  := range req.Form {
		log.Debugf("%s=>%v", k, v)
	}
}
