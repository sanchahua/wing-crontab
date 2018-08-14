package open_falcon_reporter

import (
	"gitlab.xunlei.cn/xlsoa/common/statistics"
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/sender"
)

type percentReportChData struct {
	chData *reportChData
	isNumerator bool
}

type percentReporter struct {
	descMap map[string]*reportDesc
	collector *statistics.PercentCollector
	reportCh chan *percentReportChData
	defaultStep int64

	notifyGetAll chan bool
	allMetaData chan []*open_falcon_sender.JsonMetaData
}

func newPercentReporter(defaultStep int64) *percentReporter{
	reporter := &percentReporter{
		defaultStep: defaultStep,
		descMap: make(map[string]*reportDesc),
		collector: statistics.NewPercentCollector(defaultStep),
		reportCh: make(chan *percentReportChData, 2048),

		allMetaData : make(chan []*open_falcon_sender.JsonMetaData, 1),
		notifyGetAll: make(chan bool, 1),
	}
	return reporter
}

func (reporter *percentReporter) start() {

	go func() {
		for  {
			select {
			case chData := <-reporter.reportCh:
				keyStr := chData.chData.key.KeyStr()
				reporter.createReportDescIfNotExist(keyStr, chData.chData.key, reporter.defaultStep)
				value, _ := chData.chData.value.(float64)
				reporter.collector.Inc(keyStr, chData.isNumerator, value)

			case <-reporter.notifyGetAll:
				allMetaData := make([]*open_falcon_sender.JsonMetaData, 0)
				all := reporter.collector.GetAllPercentResult()
				for k, v := range all {
					if it, ok := reporter.descMap[k]; ok {
						allMetaData = append(allMetaData, it.JsonMetaData(v.Percent))
					}
				}
				reporter.allMetaData <- allMetaData
			}
		}
	}()
}

func (reporter *percentReporter) incNumerator(key *ReportOpenFalConKey, value float64) {
	select {
	case reporter.reportCh <- &percentReportChData{
		chData: &reportChData{
			key:key,
			value: value,
		},
		isNumerator:true,
	} :

	default :
	}
}

func (reporter *percentReporter) incDenominator(key *ReportOpenFalConKey, value float64) {
	select {
	case reporter.reportCh <- &percentReportChData{
		chData: &reportChData{
			key:key,
			value: value,
		},
		isNumerator:false,
	}:

	default :
	}
}

func (reporter *percentReporter) getAllMetaData() []*open_falcon_sender.JsonMetaData{
	reporter.notifyGetAll <- true
	return <-reporter.allMetaData
}

func (reporter *percentReporter) createReportDescIfNotExist(keyStr string, key *ReportOpenFalConKey, step int64) {
	if _, ok := reporter.descMap[keyStr]; !ok {
		desc := &reportDesc{
			key : key,
			step: step,
			counterType: "GAUGE",
		}
		reporter.descMap[keyStr] = desc
	}
}