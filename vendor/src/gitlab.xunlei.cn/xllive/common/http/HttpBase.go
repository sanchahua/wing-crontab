package http

import (
	"time"
	"fmt"
	"github.com/valyala/fasthttp"
	"bytes"
	"net/http"
	"strconv"
	"io/ioutil"
	"gitlab.xunlei.cn/xllive/common/utils"
	"github.com/kataras/iris/core/errors"
	"encoding/json"
	"net/url"
)

var ErrRspDataNil = errors.New("rsp data nil")
var ErrHttpRspNil = errors.New("http rsp nil")

type HttpBaseRsp struct {
	Result  int32                     `json:"result,omitempty"`
	Message string                    `json:"message,omitempty"`
	Data    interface{}               `json:"data,omitempty"`
}

type HttpBase struct {

}

func NewHttpBase() (*HttpBase, error) {
	h := new(HttpBase)
	if err := h.Init(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *HttpBase)Init() error {
	return nil
}

func (h *HttpBase)Get(strUrl string, timeout time.Duration) ([]byte, error) {

	statusCode, body, err:= fasthttp.GetTimeout(nil, strUrl, timeout)
	if err != nil {
		return nil, err
	}

	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("statusCode=%d body='%s'", statusCode, string(body))
	}

	if len(body) == 0 {
		return nil, ErrHttpRspNil
	}

	return body, nil
}

func (h *HttpBase)Get_Data2Interface(strUrl string, timeout time.Duration, v interface{}) error {
	httpRsp, err := h.Get(strUrl, timeout)
	if err != nil {
		return fmt.Errorf("err='%v' url='%s'", err, strUrl)
	}

	return h.HttpRspUnmarshalData2Interface(httpRsp, v)
}

func (h *HttpBase)Post(strUrl string, timeout time.Duration, body string) ([]byte, error) {

	/// http request
	reqBody := bytes.NewBufferString(body)
	contentLength := strconv.Itoa(len(body))
	req, err := http.NewRequest("POST", strUrl, reqBody)
	if err != nil {
		return nil, err
	}

	/// http header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", contentLength)

	/// http client do
	client := &http.Client{Timeout: timeout}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	/// http read all
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	if len(rspBody) == 0 {
		return nil, ErrHttpRspNil
	}

	return rspBody, nil
}

func (h *HttpBase)HttpRspUnmarshalData2Interface(data []byte, v interface{}) error {
	rsp := HttpBaseRsp{Data:v}
	if err := utils.JsonGoUnmarshal(data, &rsp); err != nil {
		return err
	}

	if rsp.Result != 0 {
		return fmt.Errorf("result invalid, rsp='%v'", rsp)
	}

	return nil
}

func (h *HttpBase)Post_Data2Interface(strUrl string, timeout time.Duration, body string, v interface{}) error {

	httpRsp, err := h.Post(strUrl, timeout, body)
	if err != nil {
		return err
	}

	return h.HttpRspUnmarshalData2Interface(httpRsp, v)
}

func (m *HttpBase) RpcInnerCallfn(rsp interface{}, model, fun string, args ...interface{}) error {

	bytesArgs, err := utils.JsonMarshal(args)
	if err != nil {
		return err
	}

	strArgs := string(bytesArgs)
	strUrl := fmt.Sprintf("http://rpc.live/v1/inner/callfn?model=%s&fun=%s&args=%s", model, fun, strArgs)
	if err := m.Get_Data2Interface(strUrl, 5 * time.Second, rsp); err != nil {
		return err
	}

	return nil
}

/*
	向消息服务器发送消息的入口
*/
func (h *HttpBase)PostMsg(strUrl string, timeout time.Duration, reqParams, rspData interface{}) error {

	bytesReqParams, err := json.Marshal(reqParams)
	if err != nil {
		return err
	}

	strReqParams := string(bytesReqParams)
	escapeReqParams := url.QueryEscape(strReqParams)
	values := url.Values{}
	values.Set("msg", escapeReqParams)
	strReqBody := values.Encode()
	for i := 0; i < 3; i++ {
		if err = h.Post_Data2Interface(strUrl, timeout, strReqBody, rspData); err != nil {
			return nil
		}
	}
	return err
}
