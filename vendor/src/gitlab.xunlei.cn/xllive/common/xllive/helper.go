package xllive

import (
	"fmt"
	"crypto/md5"
	"strconv"
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

func HashId(id int64, max int64) int64 {
	idstr   := []byte(fmt.Sprintf("%v", id))
	has     := md5.Sum(idstr)
	md5str  := fmt.Sprintf("%x", has)
	v1      := md5str[:2]
	v2      := md5str[len(md5str)-2:]
	i16, _  := strconv.ParseInt(v1+v2, 16, 64)
	return i16 % max + 1
}

func ParseRoomId(roomid string) (id int64, userid int64, err error) {
	temp := strings.Split(roomid, "_")
	if len(temp) != 2 {
		return 0, 0, errors.New("roomid["+roomid+"] error")
	}
	//var err error
	id, err = strconv.ParseInt(temp[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	userid, err = strconv.ParseInt(temp[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return id, userid, nil
}

func ParseBody(body string, out interface{}) error {
	d, err := base64.StdEncoding.DecodeString(body)
	fmt.Println(string(d))
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
		fmt.Println("from ok")
		// 正常的表单
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
		fmt.Println(m)
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
