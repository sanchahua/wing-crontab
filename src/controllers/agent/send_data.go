package agent

import (
	wstring "library/string"
	"encoding/json"
)

type SendData struct {
	CronId int64     `json:"cron_id"`
	Unique string    `json:"unique"`
	Data []byte      `json:"data"`
	Status int       `json:"status"`
	Time int64       `json:"time"`
	SendTimes int    `json:"send_times"`
	Cmd int          `json:"cmd"`
	send sendFunc    `json:"-"`
	IsMutex bool     `json:"is_mutex"`
	MsgId int64      `json:"msg_id"`
	Address string   `json:"address"`
}

func newSendData(msgId int64, cmd int, data []byte, send sendFunc, id int64, isMutex bool, address string) *SendData {
	return &SendData{
		Unique:    wstring.RandString(64),
		Data:      data,
		Status:    0,
		Time:      0,
		SendTimes: 0,
		Cmd:       cmd,
		send:      send,
		CronId:    id,
		IsMutex:   isMutex,
		MsgId:     msgId,
		Address:   address,
	}

}

func (d *SendData) encode() []byte {
	b, e := json.Marshal(d)
	if e != nil {
		return nil
	}
	return b
}

func decodeSendData(data []byte) (*SendData, error) {
	var d SendData
	err := json.Unmarshal(data, &d)
	return &d, err
}

