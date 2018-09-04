package msg

import (
	"encoding/json"
	"testing"
)

var S = &Server{
	UrlPrefix: "http://msg.live/msg-server/apiRequest",
}

func TestGetUserRoomStat(t *testing.T) {
	var userIdList []uint64
	userIdList = append(userIdList, 601212133)
	userIdList = append(userIdList, 601212131)
	result, err := S.GetUserRoomStat(userIdList)
	if err != nil {
		t.Fatalf("get userRoomStat error, %v", err)
	}
	for _, res := range result {
		if res.Stat < 0 {
			t.Fatalf("stat is invalid")
		}
	}
}

func TestKick(t *testing.T) {
	var userId uint64 = 601212133
	result, err := S.Kick(userId)
	if err != nil {
		t.Fatalf("get kick error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestInRoom(t *testing.T) {
	var userId uint64 = 601212133
	roomId := "1"
	result, err := S.InRoom(userId, roomId)
	if err != nil {
		t.Fatalf("get inRoom error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestOutRoom(t *testing.T) {
	var userId uint64 = 601212133
	roomId := "1"
	result, err := S.InRoom(userId, roomId)
	if err != nil {
		t.Fatalf("get outRoom error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestSendAll(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	result, err := S.SendAll(data)
	if err != nil {
		t.Fatalf("get sendAll error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestMergedMsgRoom(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	var dataList []map[string]interface{} = make([]map[string]interface{}, 22)
	dataList = append(dataList, data)
	dataList = append(dataList, data)
	var userId uint64 = 601212133
	roomId := "1"
	result, err := S.SendMergedMsgRoom(userId, roomId, dataList)
	if err != nil {
		t.Fatalf("get SendMergedMsgRoom error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestSendRoom(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	var userId uint64 = 601212133
	roomId := "1"
	result, err := S.SendRoom(userId, roomId, data)
	if err != nil {
		t.Fatalf("get SendRoom error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestSendRooms(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	var roomIds []string = make([]string, 2)
	roomId := "1"
	roomId1 := "601212133"
	roomIds = append(roomIds, roomId)
	roomIds = append(roomIds, roomId1)
	result, err := S.SendRooms(roomIds, data)
	if err != nil {
		t.Fatalf("get SendRooms error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestSendMergedMsgUser(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	var dataList []map[string]interface{} = make([]map[string]interface{}, 22)
	dataList = append(dataList, data)
	dataList = append(dataList, data)
	var userId uint64 = 601212133
	var toUserId uint64 = 601212132
	result, err := S.SendMergedMsgUser(userId, toUserId, dataList)
	if err != nil {
		t.Fatalf("get SendMergedMsgUser error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestSendUser(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	var userId uint64 = 601212133
	var toUserId uint64 = 601212
	result, err := S.SendUser(userId, toUserId, data)
	if err != nil {
		t.Fatalf("get SendUser error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestSysMsg(t *testing.T) {
	var data map[string]interface{} = make(map[string]interface{})
	data["test"] = "rewew"
	var userId uint64 = 601212133
	roomId := "1212"
	flag := 1
	msg := "102j102"
	nickName := "heipacker"
	sendType := 1
	link := "http://live.xunlei.com/35189695"
	linkText := "12212"
	linkColor := "12212"
	var ext map[string]string = make(map[string]string)
	ext["12"] = "12"
	result, err := S.SysMsg(flag, msg, nickName, userId, roomId, sendType, link, linkText, linkColor, ext)
	if err != nil {
		t.Fatalf("get SysMsg error, %v", err)
	}
	if !result {
		t.Fatalf("result is invalid")
	}
}

func TestTicker(t *testing.T) {
	var userId uint64 = 601212133
	platform := "mqtt"
	result, err := S.Ticker(userId, platform)
	if err != nil {
		t.Fatalf("get SysMsg error, %v", err)
	}
	if result.Ticker == "" {
		t.Fatalf("result is invalid")
	}
	t.Logf("ticker result %v", result.Ticker)
}

func TestUserList(t *testing.T) {
	userId := "2"
	result, err := S.UserList(userId)
	if err != nil {
		t.Fatalf("get UserList error, %v", err)
	}
	if !(result.UserIdList == nil || len(result.UserIdList) == 0) {
		t.Fatalf("result is invalid")
	}
	if result != nil {
		marshal, err := json.Marshal(result)
		if err != nil {

		}
		t.Logf("ticker result %v", marshal)
	}
}
