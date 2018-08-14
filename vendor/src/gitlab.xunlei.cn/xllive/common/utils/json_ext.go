package utils

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"os"
	"encoding/gob"
	"encoding/json"
	"gitlab.xunlei.cn/xllive/common/json_omit"
	"github.com/go-yaml/yaml"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jsonGo = jsoniter.ConfigCompatibleWithStandardLibrary
func init() {
	extra.RegisterFuzzyDecoders()
}

/*
	发送HTTP请求，
	questUrl 请求的url
	questType 请求类型 GET POST PUT DELETE
	timeout 设置超时时间 秒
	data 请求的数据
	header 头部信息
	sendType 1:map, 2:json
	返回，查询后的响应结果
*/

func HttpPost(ConnId uint64, urlStr string, timeout time.Duration, params interface{}) ([]byte, error) {

	/// http request
	values := url.Values{}
	for dataKey, dataValue := range params.(map[string]string) {
		values.Set(dataKey, dataValue)
	}
	valueEncoded := values.Encode()
	reqBody := bytes.NewBufferString(valueEncoded)
	contentLength := len(valueEncoded)

	req, err := http.NewRequest("POST", urlStr, reqBody)
	if err != nil {
		return nil, err
	}

	/// http header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(contentLength))

	/// http client do
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	/// http read all
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func Contain(v string, arr []string) (int, bool) {
	for i, value := range arr {
		if v == value {
			return i + 1, true
		}
	}
	return 0, false
}

func DeepCopy(dst, src interface{}) error {
	gob.Register(map[string]interface{}{})
	var a json.Number = ""
	gob.Register(a)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func FileExist(filename string) bool {
	info, err := os.Stat(filename)
	if err == nil {
		if false == info.IsDir() {
			return true
		}
	}

	if os.IsExist(err) {
		return true
	}

	return false
}

func LoadYaml2Mapsi(configName string) (map[string]interface{}, error) {
	bytesCfg, err := ioutil.ReadFile(configName)
	if err != nil {
		return nil, err
	}

	var mapCfg map[string]interface{}
	if err = yaml.Unmarshal(bytesCfg, &mapCfg); err != nil {
		return nil, err
	}

	if len(mapCfg) == 0 {
		return nil, fmt.Errorf("mapCfg empty")
	}

	return mapCfg, nil
}

func LoadYaml(configName string, v interface{}) error {
	bytesCfg, err := ioutil.ReadFile(configName)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(bytesCfg, v); err != nil {
		return err
	}

	return nil
}

func LoadJson(configName string, v interface{}) error {
	bytesCfg, err := ioutil.ReadFile(configName)
	if err != nil {
		return err
	}

	if err = jsonGo.Unmarshal(bytesCfg, v); err != nil {
		return err
	}

	return nil
}

func LoadJson2Mapsi(configName string) (map[string]interface{}, error) {
	bytesCfg, err := ioutil.ReadFile(configName)
	if err != nil {
		return nil, err
	}

	var mapCfg map[string]interface{}
	if err = json.Unmarshal(bytesCfg, &mapCfg); err != nil {
		return nil, err
	}

	if len(mapCfg) == 0 {
		return nil, fmt.Errorf("mapCfg empty")
	}

	return mapCfg, nil
}

func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func JsonGoMarshal(v interface{}) ([]byte, error) {
	return jsonGo.Marshal(v)
}

func JsonOmitMarshal(v interface{}) ([]byte, error) {
	return json_omit.Marshal(v)
}

func JsonGoUnmarshal(data []byte, v interface{}) error {
	return jsonGo.Unmarshal(data, v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func JsonMarshalString(v interface{}) string {
	if v == nil {
		return ""
	}
	strJson, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(strJson)
}

/// 忽略关键字 omitempty 的作用，转换所有的字段
func JsonOmitMarshalString(v interface{}) string {
	if v == nil {
		return ""
	}
	strJson, err := json_omit.Marshal(v)
	if err != nil {
		return ""
	}
	return string(strJson)
}

func JsonGoMarshalString(v interface{}) string {
	if v == nil {
		return ""
	}
	strJson, err := jsonGo.Marshal(v)
	if err != nil {
		return ""
	}
	return string(strJson)
}

func JsonpbUnmarshal(data []byte, pb proto.Message) error {
	return jsonpb.Unmarshal(bytes.NewBuffer(data), pb)
}

func JsonpbMarshal(pb proto.Message, emit bool) ([]byte, error) {
	var w *bytes.Buffer = bytes.NewBuffer([]byte{})
	marshaler := &jsonpb.Marshaler{}
	marshaler.EmitDefaults = emit
	if err := marshaler.Marshal(w, pb); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}