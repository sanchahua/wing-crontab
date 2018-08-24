package manager

import (
	"testing"
	"github.com/parnurzeal/gorequest"
	"net/url"
	"fmt"
	"github.com/pkg/errors"
	"encoding/json"
)

// xcrontab restful api黑盒测试

type Rsp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data *CronEntity `json:"data"`
}

type CronEntity struct {
	// 数据库的基本属性
	Id int64        `json:"id"`
	CronSet string  `json:"cron_set"`
	Command string  `json:"command"`
	Remark string   `json:"remark"`
	Stop bool       `json:"stop"`
	StartTime int64 `json:"start_time"`
	EndTime int64   `json:"end_time"`
	IsMutex bool    `json:"is_mutex"`
}

func httpGet(uri string) (*Rsp, error) {
	resp, body, errs := gorequest.New().Get(uri).End()
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error http code: ", resp.StatusCode))
	}
	//return []byte(body), nil
	var rsp Rsp
	err := json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

func httpPost(uri string, data url.Values) (*Rsp, error) {
	resp, body, errs := gorequest.New().Post(uri).Send(data.Encode()).End()
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error http code: ", resp.StatusCode))
	}
	//return []byte(body), nil
	var rsp Rsp
	err := json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

func addCron() (*CronEntity, error) {
	uri := "http://localhost:38001/cron/add"
	/**
	cronSet      := request.QueryParameter("cronSet")
	command      := request.QueryParameter("command")
	remark       := request.QueryParameter("remark")
	stop         := request.QueryParameter("stop")
	strStartTime := request.QueryParameter("start_time")
	strEndTime   := request.QueryParameter("end_time")
	strIsMutex   := request.QueryParameter("is_mutex")
	*/
	data := make(url.Values)
	data.Add("cron_set", "0 */1 * * * *")
	data.Add("command", "ls /")
	data.Add("remark", "")
	data.Add("stop", "0")
	data.Add("start_time", "0")
	data.Add("end_time", "0")
	data.Add("is_mutex", "0")

	body, err := httpPost(uri, data)
	if err != nil {
		//t.Errorf("%v", err)
		return nil, err
	}
	if body.Code != 2000 {
		//t.Errorf("%v", body.Message)
		return nil, errors.New(body.Message)
	}
	return body.Data, nil
}

func updateCron(id int64) error {
	uri := fmt.Sprintf("http://localhost:38001/cron/update/%d", id)
	data := make(url.Values)
	data.Add("cron_set", "0 */2 * * *")
	data.Add("command", "ls /home")
	data.Add("remark", "new remark")
	data.Add("stop", "1")
	data.Add("start_time", "1")
	data.Add("end_time", "2")
	data.Add("is_mutex", "1")

	body, err := httpPost(uri, data)
	if err != nil {
		return err
	}
	if body.Code != 2000 {
		return errors.New(body.Message)
	}
	return nil
}

func delCron(id int64) error {
	uri := fmt.Sprintf("http://localhost:38001/cron/delete/%d", id)
	rsp, err := httpGet(uri)
	if err != nil {
		return err
	}
	if rsp.Code != 2000 {
		return errors.New(rsp.Message)
	}
	return nil
}

func stopCron(id int64) error {
	uri := fmt.Sprintf("http://localhost:38001/cron/stop/%d", id)
	rsp, err := httpGet(uri)
	if err != nil {
		return err
	}
	if rsp.Code != 2000 {
		return errors.New(rsp.Message)
	}
	return nil
}

func startCron(id int64) error {
	uri := fmt.Sprintf("http://localhost:38001/cron/start/%d", id)
	rsp, err := httpGet(uri)
	if err != nil {
		return err
	}
	if rsp.Code != 2000 {
		return errors.New(rsp.Message)
	}
	return nil
}

func Test_AddCron(t *testing.T) {
	e, err := addCron()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	err = delCron(e.Id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_StopCron(t *testing.T) {
	e, err := addCron()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer delCron(e.Id)
	err = stopCron(e.Id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}

// 定时任务必须停止后才能开始
// 测试流程
// 1、添加一个定时任务
// 2、停止
// 3、开始
// 4、删除
func Test_StartCron(t *testing.T) {
	e, err := addCron()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer delCron(e.Id)
	err = stopCron(e.Id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	err = startCron(e.Id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}


func Test_UpdateCron(t *testing.T) {
	e, err := addCron()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	defer delCron(e.Id)
	err = updateCron(e.Id)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
}

func Test_GetLogs(t *testing.T) {
	res, err := httpGet("http://localhost:38001/log/list/0/0/0")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if res.Code != 2000 {
		t.Errorf("get log list fail: %v", res.Message)
	}
}

func Test_CronList(t *testing.T) {
	//http://localhost:38001/cron/list
	res, err := httpGet("http://localhost:38001/cron/list")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if res.Code != 2000 {
		t.Errorf("get log list fail: %v", res.Message)
	}
}