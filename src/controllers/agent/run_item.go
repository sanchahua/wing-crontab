package agent

import "encoding/json"

type runItem struct {
	Id int64       `json:"id"`
	Command string `json:"command"`
	IsMutex bool   `json:"is_mutex"`
	SubWaitNum func() int64 `json:"-"`
}

func (r *runItem) encode() ([]byte, error) {
	return json.Marshal(r)
}

func decodeRunItem(data []byte) (*runItem, error) {
	var r runItem
	err := json.Unmarshal(data, &r)
	return &r, err
}