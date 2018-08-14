package open_falcon_reporter

import (
	"time"
	"gitlab.xunlei.cn/xlsoa/common/statistics"
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/sender"
)

func newTimingReporter(defaultStep int64) *timingReporter{
	reporter := &timingReporter{
		defaultStep: defaultStep,
		descMap: make(map[string]*reportDesc),
		collector: statistics.NewTimingCollector(defaultStep),
		reportChData: make(chan*reportChData, 2048),

		allMetaData : make(chan []*open_falcon_sender.JsonMetaData, 1),
		notifyGetAll: make(chan bool, 1),
	}
	return reporter
}

type timingReporter struct {
	descMap map[string]*reportDesc
	collector *statistics.TimingCollector
	defaultStep int64
	reportChData chan *reportChData

	notifyGetAll chan bool
	allMetaData chan []*open_falcon_sender.JsonMetaData
}

func (reporter *timingReporter) start() {
	go func() {
		for  {
			select {
			case chData := <- reporter.reportChData :
				keyStr := chData.key.KeyStr()
				reporter.createReportDescIfNotExist(keyStr, chData.key, reporter.defaultStep)
				value, _ := chData.value.(time.Duration)
				reporter.collector.Add(keyStr, value)

			case <- reporter.notifyGetAll:
				allMetaData := make([]*open_falcon_sender.JsonMetaData, 0 )
				all := reporter.collector.GetAllTimingResult()
				for k, v := range all {
					if it, ok := reporter.descMap[k]; ok {
						allMetaData = append(allMetaData, it.JsonMetaData(v.Avg))
					}
				}
				reporter.allMetaData <- allMetaData
			}
		}
	}()
}

func (reporter *timingReporter) add(key *ReportOpenFalConKey, value time.Duration) {
	select {
	case reporter.reportChData <- &reportChData{key: key, value:value} :
	default :
	}
}

func (reporter *timingReporter) getAllMetaData() []*open_falcon_sender.JsonMetaData{
	reporter.notifyGetAll <- true
	return <-reporter.allMetaData
}

func (reporter *timingReporter) createReportDescIfNotExist(keyStr string, key *ReportOpenFalConKey, step int64) {
	if _, ok := reporter.descMap[keyStr]; !ok {
		desc := &reportDesc{
			key : key,
			step: step,
			counterType: "GAUGE",
		}
		reporter.descMap[keyStr] = desc
	}
}