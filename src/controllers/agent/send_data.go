package agent

import (
	wstring "library/string"
	"encoding/json"
)

type SendData struct {
	CronId int64  `json:"cron_id"`
	Unique string `json:"unique"`
	Data []byte `json:"data"`
	Status int `json:"status"`
	Time int64 `json:"time"`
	SendTimes int `json:"send_times"`
	Cmd int `json:"cmd"`
	send sendFunc `json:"-"`
}

func newSendData(cmd int, data []byte, send sendFunc, id int64) *SendData {
	return &SendData{
		Unique:    wstring.RandString(128),
		Data:      data,
		Status:    0,
		Time:      0,
		SendTimes: 0,
		Cmd:       cmd,
		send:      send,
		CronId:    id,
	}

}

func (d *SendData) encode() []byte {
	b, e := json.Marshal(d)
	if e != nil {
		return nil
	}
	return b
}

