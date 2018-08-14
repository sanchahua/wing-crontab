package open_falcon_reporter

import (
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/sender"
)

type statusReporter struct {
	descMap map[string]*reportDesc
	reportChData chan *reportChData
	status map[string]int
	defaultStep int64

	notifyGetAll chan bool
	allMetaData chan []*open_falcon_sender.JsonMetaData
}

func newStatusReporter(defaultStep int64) *statusReporter{
	reporter := &statusReporter{
		descMap: make(map[string]*reportDesc),
		status: make(map[string]int),
		reportChData: make(chan *reportChData, 2048),
		defaultStep: defaultStep,

		allMetaData : make(chan []*open_falcon_sender.JsonMetaData, 1),
		notifyGetAll: make(chan bool, 1),
	}
	return reporter
}

func (reporter *statusReporter) start() {
	go func() {
		for  {
			select {
			case chData := <- reporter.reportChData :
				keyStr := chData.key.KeyStr()
				reporter.createReportDescIfNotExist(keyStr, chData.key)
				value, _ := chData.value.(int)
				reporter.status[keyStr] = value

			case <- reporter.notifyGetAll:
				allMetaData := make([]*open_falcon_sender.JsonMetaData, 0 )
				for k, v := range reporter.status {
					if it, ok := reporter.descMap[k]; ok {
						allMetaData = append(allMetaData, it.JsonMetaData(v))
					}
				}
				reporter.allMetaData <- allMetaData
			}
		}
	}()
}

func (reporter *statusReporter) setStatus(key *ReportOpenFalConKey, status int) {
	select {
	case reporter.reportChData <- &reportChData{key: key, value:status} :
	default :
	}
}

func (reporter *statusReporter) getAllMetaData() []*open_falcon_sender.JsonMetaData{
	reporter.notifyGetAll <- true
	return <-reporter.allMetaData
}

func (reporter *statusReporter) createReportDescIfNotExist(keyStr string, key *ReportOpenFalConKey) {
	if _, ok := reporter.descMap[keyStr]; !ok {
		desc := &reportDesc{
			key : key,
			step: reporter.defaultStep,
			counterType: "GAUGE",
		}
		reporter.descMap[keyStr] = desc
	}
}
