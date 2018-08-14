package open_falcon_reporter

import (
	"gitlab.xunlei.cn/xlsoa/common/statistics"
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/sender"
)

type counterReporter struct {
	descMap map[string]*reportDesc
	collector *statistics.CounterCollector
	reportChData chan *reportChData
	defaultStep int64

	notifyGetAll chan bool
	allMetaData chan []*open_falcon_sender.JsonMetaData
}

func newCounterReporter(defaultStep int64) *counterReporter{
	reporter := &counterReporter{
		defaultStep: defaultStep,
		descMap: make(map[string]*reportDesc),
		collector: statistics.NewCounterCollector(),
		reportChData: make(chan *reportChData, 2048),

		allMetaData : make(chan []*open_falcon_sender.JsonMetaData, 1),
		notifyGetAll: make(chan bool, 1),
	}
	return reporter
}

func (reporter *counterReporter) start() {
	go func() {
		for {
			select {
			case chData := <- reporter.reportChData :
				keyStr := chData.key.KeyStr()
				reporter.createReportDescIfNotExist(keyStr, chData.key, reporter.defaultStep)
				value, _ := chData.value.(float64)
				reporter.collector.Inc(keyStr, value)

			case <- reporter.notifyGetAll:
				allMetaData := make([]*open_falcon_sender.JsonMetaData, 0 )
				all := reporter.collector.GetAllCounter()
				for k, v := range all {
					if it, ok := reporter.descMap[k]; ok {
						allMetaData = append(allMetaData, it.JsonMetaData(v))
					}
				}
				reporter.allMetaData <- allMetaData
			}
		}
	}()
}

func (reporter *counterReporter) inc(key *ReportOpenFalConKey, value float64) {
	select {
	case reporter.reportChData <- &reportChData{key: key, value:value} :
	default :
	}
}

func (reporter *counterReporter) getAllMetaData() []*open_falcon_sender.JsonMetaData{
	reporter.notifyGetAll <- true
	return <-reporter.allMetaData
}

func (reporter *counterReporter) createReportDescIfNotExist(keyStr string, key *ReportOpenFalConKey, step int64) {
	if _, ok := reporter.descMap[keyStr]; !ok {
		desc := &reportDesc{
			key : key,
			step: step,
			counterType: "COUNTER",
		}
		reporter.descMap[keyStr] = desc
	}
}