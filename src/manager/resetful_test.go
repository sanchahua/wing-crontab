package manager

import (
	"testing"
	"github.com/parnurzeal/gorequest"
	"net/url"
	"fmt"
	"errors"
	"encoding/json"
)

// wing-crontab restful api黑盒测试

type Rsp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data *CronEntity `json:"data"`
}

type RspList struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data []*CronEntity `json:"data"`
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

func get(uri string) ([]byte, error) {
	resp, body, errs := gorequest.New().Get(uri).End()
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error http code: %v", resp.StatusCode))
	}
	return []byte(body), nil
}

func httpGet(uri string) (*Rsp, error) {
	//resp, body, errs := gorequest.New().Get(uri).End()
	//if len(errs) > 0 && errs[0] != nil {
	//	return nil, errs[0]
	//} else if resp.StatusCode != 200 {
	//	return nil, errors.New(fmt.Sprintf("error http code: ", resp.StatusCode))
	//}
	body, err := get(uri)
	if err != nil {
		return nil, err
	}
	//return []byte(body), nil
	var rsp Rsp
	err = json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

func httpGetList(uri string) (*RspList, error) {
	//resp, body, errs := gorequest.New().Get(uri).End()
	//if len(errs) > 0 && errs[0] != nil {
	//	return nil, errs[0]
	//} else if resp.StatusCode != 200 {
	//	return nil, errors.New(fmt.Sprintf("error http code: ", resp.StatusCode))
	//}
	//return []byte(body), nil
	body, err := get(uri)
	if err != nil {
		return nil, err
	}
	var rsp RspList
	err = json.Unmarshal([]byte(body), &rsp)
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
		return nil, errors.New(fmt.Sprintf("error http code: %v", resp.StatusCode))
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
	data.Add("start_time", "2018-01-01 9:00:00")
	data.Add("end_time", "2018-11-01 9:00:00")
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
	data.Add("start_time", "2018-01-01 9:00:00")
	data.Add("end_time", "2018-12-01 9:00:00")
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


type AddUserRsp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data int64 `json:"data"`
}

func AddUserPost(uri string, data url.Values) (*AddUserRsp, error) {
	resp, body, errs := gorequest.New().Post(uri).Send(data.Encode()).End()
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error http code: %v", resp.StatusCode))
	}
	//return []byte(body), nil
	var rsp AddUserRsp
	err := json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}


type DelUserRsp struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func UpdateUserPost(uri string, data url.Values) (error) {
	resp, body, errs := gorequest.New().Post(uri).Send(data.Encode()).End()
	if len(errs) > 0 && errs[0] != nil {
		return errs[0]
	} else if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error http code: %v", resp.StatusCode))
	}
	//return []byte(body), nil
	var rsp DelUserRsp
	err := json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return err
	}
	if rsp.Code != 2000 {
		return errors.New(rsp.Message)
	}
	return nil
}

func DelUserPost(id int64) (error) {
	uri := fmt.Sprintf("http://localhost:38001/user/delete/%v", id)
	resp, body, errs := gorequest.New().Post(uri).End()
	if len(errs) > 0 && errs[0] != nil {
		return errs[0]
	} else if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error http code: %v", resp.StatusCode))
	}
	//return []byte(body), nil
	var rsp DelUserRsp
	err := json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return err
	}
	if rsp.Code != 2000 {
		return errors.New(rsp.Message)
	}
	return nil
}

type Entity struct {
	//SELECT `id`, `user_name`, `password`, `real_name`,
	//`phone`, `created`, `updated` FROM `users` WHERE 1
	Id       int64  `json:"id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
	Enable   bool   `json:"enable"`
}

type GetUserInfoRsp struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data *Entity `json:"data"`
}

func httpGetUserInfo(id int64) (*Entity, error) {
	body, err := get(fmt.Sprintf("http://localhost:38001/user/info/%v", id))
	if err != nil {
		return nil, err
	}
	//return []byte(body), nil
	var rsp GetUserInfoRsp
	err = json.Unmarshal([]byte(body), &rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Data, nil
}

func AddUser() (int64, error) {
	uri := "http://localhost:38001/user/register"
	data := make(url.Values)
	data.Add("username", "111")
	data.Add("password", "111")
	data.Add("real_name", "111")
	data.Add("phone", "111")


	body, err := AddUserPost(uri, data)
	if err != nil {
		return 0, err
	}
	if body.Code != 2000 {
		return 0, errors.New(body.Message)
	}
	return body.Data, nil
}

func UpdateUser(id int64) (error) {
	uri := fmt.Sprintf("http://localhost:38001/user/update/%v", id)
	data := make(url.Values)
	data.Add("username", "112")
	data.Add("password", "112")
	data.Add("real_name", "112")
	data.Add("phone", "112")
	return UpdateUserPost(uri, data)
}

// go test -v -test.run Test_AddUser
func Test_AddUser(t *testing.T) {
	userId, err := AddUser()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	_, err = AddUser()
	if err == nil {
		t.Errorf("%v", "add user check exists fail")
		return
	}

	info, err := httpGetUserInfo(userId)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if info.Id != userId ||
		info.UserName != "111" ||
		info.RealName != "111" || info.Phone != "111" {
		t.Errorf("%v", "info check fail")
		return
	}

	err = UpdateUser(userId)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	info, err = httpGetUserInfo(userId)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if info.Id != userId ||
		//info.Password != "112" ||
		info.UserName != "112" ||
		info.RealName != "112" ||
		info.Phone != "112" {
		t.Errorf("%v", "info check fail")
		return
	}

	err = DelUserPost(userId)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
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
	res, err := httpGetList("http://localhost:38001/cron/list")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if res.Code != 2000 {
		t.Errorf("get log list fail: %v", res.Message)
	}
}