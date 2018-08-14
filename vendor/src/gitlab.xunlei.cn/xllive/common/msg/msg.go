package msg

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"encoding/json"
	log "github.com/cihub/seelog"
	"strconv"
	"strings"
	"time"
	"errors"
)

const (
	USER = iota
	ROOM
	ALL
)

var defaultUserId uint64 = 0

var client = &http.Client{}

func init() {
	client.Timeout = 30 * time.Second
}

type Server struct {
	UrlPrefix string
}

func NewServer(urlPrefix string) (*Server) {
	if urlPrefix == "" {
		urlPrefix = "http://msg.live/msg-server/apiRequest"
	}
	var w = &Server{
		UrlPrefix: urlPrefix,
	}
	return w
}

func (s *Server) doRequest(method string, params map[string]interface{}) (string, error) {
	marshal, err := json.Marshal(params)
	if err != nil {
		log.Errorf("marshal params %v, error %v", params, err)
		return "", err
	}
	log.Debugf("request %v params %v", s.UrlPrefix+method, string(marshal))
	queryEscape := url.QueryEscape(string(marshal))
	req, err := http.NewRequest("POST", s.UrlPrefix+method, strings.NewReader("msg="+queryEscape))
	if err != nil {
		log.Errorf("new request error %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if status := resp.StatusCode; status < 200 || status >= 300 {
		log.Warnf("status code %v is not 200", resp.StatusCode)
		return "", errors.New("status code is " + strconv.Itoa(status))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type UserGroupResp struct {
	UserId string `json:"userid"`
	RoomId string `json:"roomid"`
	Stat   int    `json:"stat"`
}

//得到用户的长连接状态
func (s *Server) GetUserRoomStat(userIdList []uint64) ([]UserGroupResp, error) {
	if userIdList == nil || len(userIdList) == 0 {
		return nil, errors.New("userIdList is empty")
	}
	var uniqueUserIdList []string
	userIdMap := make(map[uint64]bool)
	for _, userId := range userIdList {
		if userIdMap[userId] {
			continue
		}
		userIdMap[userId] = true
		uniqueUserIdList = append(uniqueUserIdList, strconv.FormatUint(userId, 10))
	}
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "getUserRoomStat"
	bodyParams["userids"] = uniqueUserIdList

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return nil, err
	}
	type RespData struct {
		Result int             `json:"result"`
		Msg    string          `json:"msg"`
		Data   []UserGroupResp `json:"data"`
	}
	ret := &RespData{
		Result: 0,
		Msg:    "",
		Data:   []UserGroupResp{},
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return nil, err
	} else if ret.Result != 0 {
		return nil, errors.New(ret.Msg)
	}
	return ret.Data, nil
}

type DataIgnoreRespData struct {
	Result int         `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

// 用户下线（令牌作废）
func (s *Server) Kick(userId uint64) (bool, error) {
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "kick"
	bodyParams["userid"] = strconv.FormatUint(userId, 10)

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return false, err
	}
	ret := &DataIgnoreRespData{
		Result: 0,
		Msg:    "",
		Data:   new(interface{}),
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return false, err
	} else if ret.Result == 0 {
		return true, nil
	}
	return false, nil
}

// 进入房间
func (s *Server) InRoom(userId uint64, roomId string) (bool, error) {
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "inroom"
	bodyParams["userid"] = strconv.FormatUint(userId, 10)
	bodyParams["roomid"] = roomId

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return false, err
	}
	ret := &DataIgnoreRespData{
		Result: 0,
		Msg:    "",
		Data:   new(interface{}),
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return false, err
	} else if ret.Result == 0 {
		return true, nil
	}
	return false, nil
}

// 出房间
func (s *Server) OutRoom(userId uint64, roomId string) (bool, error) {
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "outroom"
	bodyParams["userid"] = strconv.FormatUint(userId, 10)
	bodyParams["roomid"] = roomId

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return false, err
	}
	ret := &DataIgnoreRespData{
		Result: 0,
		Msg:    "",
		Data:   new(interface{}),
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return false, err
	} else if ret.Result == 0 {
		return true, nil
	}
	return false, nil
}

type TickerResp struct {
	UserId string `json:"userid"`
	Ticker string `json:"ticker"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}

//取用户令牌 platform 平台 "mqtt"（代表app） 或 "websocket"（代表h5）
func (s *Server) Ticker(userId uint64, platform string) (*TickerResp, error) {
	data := make(map[string]interface{})
	data["cmd"] = "ticker"
	data["userid"] = strconv.FormatUint(userId, 10)
	data["platform"] = platform

	body, err := s.doRequest("", data)
	if err != nil {
		return nil, err
	}
	type RespData struct {
		Result int         `json:"result"`
		Msg    string      `json:"msg"`
		Data   *TickerResp `json:"data"`
	}
	ret := &RespData{
		Result: 0,
		Msg:    "",
		Data:   &TickerResp{},
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return nil, err
	}
	if ret.Result != 0 {
		return nil, errors.New(ret.Msg)
	}
	return ret.Data, nil
}

type UserListResp struct {
	Count      int      `json:"count"`
	UserIdList []string `json:"userids"`
}

//得到房间中当前所有在线的用户
func (s *Server) UserList(roomId string) (*UserListResp, error) {
	data := make(map[string]interface{})
	data["cmd"] = "userlist"
	data["roomid"] = roomId

	body, err := s.doRequest("", data)
	if err != nil {
		return nil, err
	}
	type RespData struct {
		Result int           `json:"result"`
		Msg    string        `json:"msg"`
		Data   *UserListResp `json:"data"`
	}
	ret := &RespData{
		Result: 0,
		Msg:    "",
		Data:   &UserListResp{},
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return nil, err
	} else if ret.Result != 0 {
		return nil, errors.New(ret.Msg)
	}
	return ret.Data, nil
}

// 向所有房间中的用户推消息（广播）
func (s *Server) SendAll(data map[string]interface{}) (bool, error) {
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "sendall"
	bodyParams["topic"] = "msg"

	if marshal, err := json.Marshal(data); err != nil {
		return false, err
	} else {
		bodyParams["data"] = string(marshal)
	}

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return false, err
	}
	ret := &DataIgnoreRespData{
		Result: 0,
		Msg:    "",
		Data:   new(interface{}),
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return false, err
	} else if ret.Result == 0 {
		return true, nil
	}
	return false, nil
}

// 向房间所有用户推合并的多个消息 （房间消息）
func (s *Server) SendMergedMsgRoom(userId uint64, roomId string, dataList []map[string]interface{}) (bool, error) {
	data := make(map[string]interface{})
	data["cmd"] = "onmsgmerge"
	data["lists"] = dataList
	return s.SendRoom(userId, roomId, data)
}

// 向房间所有用户推消息（房间消息）
func (s *Server) SendRoom(userId uint64, roomId string, data map[string]interface{}) (bool, error) {
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "sendroom"
	bodyParams["topic"] = "msg"
	bodyParams["userid"] = strconv.FormatUint(userId, 10)
	bodyParams["roomid"] = roomId

	data["range"] = "multicast"
	if marshal, err := json.Marshal(data); err != nil {
		return false, err
	} else {
		bodyParams["data"] = string(marshal)
	}

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return false, err
	}
	ret := &DataIgnoreRespData{
		Result: 0,
		Msg:    "",
		Data:   new(interface{}),
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return false, err
	} else if ret.Result == 0 {
		return true, nil
	}
	return false, nil
}

// 向多个房间发送消息,
func (s *Server) SendRooms(roomIds []string, data map[string]interface{}) (bool, error) {
	var ret = true
	for _, roomId := range roomIds {
		result, err := s.SendRoom(defaultUserId, roomId, data)
		if err != nil {
			return false, err
		}
		if !result {
			ret = result
		}
	}
	return ret, nil
}

// 用户与用户之间推合并消息
func (s *Server) SendMergedMsgUser(userId uint64, toUserId uint64, dataList []map[string]interface{}) (bool, error) {
	data := make(map[string]interface{})
	data["cmd"] = "onmsgmerge"
	data["lists"] = dataList
	return s.SendUser(userId, toUserId, data)
}

// 用户与用户之间发消息
func (s *Server) SendUser(userId uint64, toUserId uint64, data map[string]interface{}) (bool, error) {
	bodyParams := make(map[string]interface{})
	bodyParams["cmd"] = "senduser"
	bodyParams["topic"] = "msg"
	bodyParams["userid"] = strconv.FormatUint(userId, 10)
	bodyParams["touserid"] = strconv.FormatUint(toUserId, 10)

	data["range"] = "unicast"
	if marshal, err := json.Marshal(data); err != nil {
		return false, err
	} else {
		bodyParams["data"] = string(marshal)
	}

	body, err := s.doRequest("", bodyParams)
	if err != nil {
		return false, err
	}
	ret := &DataIgnoreRespData{
		Result: 0,
		Msg:    "",
		Data:   new(interface{}),
	}
	if err = json.Unmarshal([]byte(body), ret); err != nil {
		return false, err
	} else if ret.Result == 0 {
		return true, nil
	}
	return false, nil
}

/**
 * 系统消息管理
 * 0.主播离开
 * 1.主播回来
 * 2.关注主播
 * 3.系统公告
 * 4.分享信息
 * 5.禁言
 * 6.座驾续费
 * 7.主播完成任务
 * 8.红包中奖
 * 9.火箭炮领取全服通知
 * 10.主播升级信息
 * 11.公告
 * 12.禁言(主播守护版)
 */
func (s *Server) SysMsg(flag int, msg string, nickName string, userId uint64, roomId string, sendType int, link string, linkText string, linkColor string, ext map[string]string) (bool, error) {
	data := make(map[string]interface{})
	data["cmd"] = "onsysmsg"
	data["flag"] = flag
	data["nickname"] = nickName
	data["msg"] = msg
	data["userid"] = strconv.FormatUint(userId, 10)
	data["roomid"] = roomId
	data["link"] = link
	data["linkText"] = linkText
	data["linkColor"] = linkColor
	data["ext"] = ext

	if sendType == USER {
		return s.SendUser(defaultUserId, userId, data)
	} else if sendType == ROOM {
		return s.SendRoom(userId, roomId, data)
	} else if sendType == ALL {
		return s.SendAll(data)
	}
	return false, nil
}
