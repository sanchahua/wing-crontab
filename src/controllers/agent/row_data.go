package agent

import (
	//log "github.com/sirupsen/logrus"
	"models/cron"
	"encoding/json"
)

type rowData struct {
	Event int `json:"event"`
	Row *cron.CronEntity `json:"row"`
}

func (r *rowData) encode() ([]byte, error) {
	return json.Marshal(r)
}

func decodeRowData(data []byte) (*rowData, error) {
	var d rowData
	err := json.Unmarshal(data, &d)
	return &d, err
}
