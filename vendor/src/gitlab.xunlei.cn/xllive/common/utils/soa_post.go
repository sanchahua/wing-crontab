package utils

import (
	"fmt"
	"strings"
	"errors"
	"net/url"
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"encoding/base64"
	"regexp"
)
func init() {
	extra.RegisterFuzzyDecoders()
}
var jsonext = jsoniter.ConfigCompatibleWithStandardLibrary


func ParseBody(body string, out interface{}) error {
	d, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return err
	}
	body , err = url.QueryUnescape(string(d))
	if err != nil {
		return err
	}

	// 如果body是个json
	err = jsonext.Unmarshal([]byte(body), out)
	if err == nil {
		return nil
	}

	// Content-Type: multipart/form-data; boundary=---------------------------9747672998524
	// 类型的表单解析支持
	//fmt.Println(body)
	im, _ := regexp.MatchString(`[\-]{1,}[\S\s]{1,}Content\-Disposition\:[\s]form\-data`, body)
	var m = make(map[string]interface{})
	if im {
		// 正常的表单
		//boundary := regexp.MustCompile(`[\-]{1,}[a-zA-Z0-9]{1,}`).FindString(body)
		boundary := regexp.MustCompile(`[\-]{1,}[a-zA-Z0-9\-_]{1,}`).FindString(body)

		if boundary == "" {
			return errors.New("body invalid")
		}
		stemp := strings.Split(body, boundary)
		for _, v := range stemp  {
			v = strings.Trim(v, " ")
			v = strings.Trim(v, "\r\n")
			vt := strings.Split(v, "\r\n")
			if len(vt)!=3 {
				continue
			}
			vt[0] = strings.Trim(vt[0], "\"")
			name := vt[0][strings.Index(vt[0], "\"")+1:]
			vt[2] = strings.Trim(vt[2], " ")
			m[name] = vt[2]
		}
	} else {
		// 尝试按照字符串解析
		values, err := url.ParseQuery(body)
		if err != nil {
			return err
		}
		fmt.Println(values)

		for k, v := range values {
			m[k] = nil
			if len(v) > 0 {
				m[k] = v[0]
			}
		}
	}
	d, err = jsonext.Marshal(m)
	if err != nil {
		return err
	}
	err = jsonext.Unmarshal(d, out)
	if err != nil {
		return err
	}
	return nil
}

