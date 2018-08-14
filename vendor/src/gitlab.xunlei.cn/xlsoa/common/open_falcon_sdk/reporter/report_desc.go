package open_falcon_reporter

import (
	"fmt"
	"strings"
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/sender"
)

type ReportOpenFalConKey struct {
	Endpoint    string
	Metric      string
	Tags        []string
}

type reportChData struct {
	key *ReportOpenFalConKey
	value interface{}
}

func (key *ReportOpenFalConKey) KeyStr() string {
	return fmt.Sprintf("Key{%s:%s:%s}", key.Endpoint, key.Metric, strings.Join(key.Tags, ","))
}

func (key *ReportOpenFalConKey) String() string {
	return fmt.Sprintf("Key{Endpoint=%s,Metric=%s,Metric=%s}", key.Endpoint, key.Metric, strings.Join(key.Tags, ","))
}

type reportDesc struct {
	key *ReportOpenFalConKey
	step        int64
	counterType string
	tags string
}

func (item *reportDesc) JsonMetaData(value interface{}) *open_falcon_sender.JsonMetaData{
	if item.tags == "" {
		item.tags = strings.Join(item.key.Tags, ",")
	}
	if item.step == 0 {
		item.step = 60
	}
	return open_falcon_sender.MakeMetaData(item.key.Endpoint, item.key.Metric, item.tags, value, item.counterType, item.step)
}