package open_falcon_reporter

import (
	"time"
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/sender"
	"sync"
)

var DefaultReportStep int64 = 60

var Reporter *OpenFalconReporter = newOpenFalconReporter()

type OpenFalconReporter struct {
	counter *counterReporter
	gauge *gaugeReporter
	timing *timingReporter
	percent *percentReporter
	status *statusReporter

	senderQueue *open_falcon_sender.SafeLinkedList

	startMutex *sync.Mutex
	isStart bool
}

func newOpenFalconReporter() *OpenFalconReporter {
	return &OpenFalconReporter{
		startMutex: new(sync.Mutex),
		counter: newCounterReporter(DefaultReportStep),
		gauge: newGaugeReporter(DefaultReportStep),
		timing: newTimingReporter(DefaultReportStep),
		percent: newPercentReporter(DefaultReportStep),
		status: newStatusReporter(DefaultReportStep),
		senderQueue : open_falcon_sender.MetaDataQueue,
	}
}

func (reporter *OpenFalconReporter) runReporter() {
	if reporter.counter != nil {
		reporter.counter.start()
	}
	if reporter.gauge != nil {
		reporter.gauge.start()
	}
	if reporter.timing != nil {
		reporter.timing.start()
	}
	if reporter.percent != nil {
		reporter.percent.start()
	}
	if reporter.status != nil {
		reporter.status.start()
	}
}

func (reporter *OpenFalconReporter) SetStatus(endpoint, metric string, tags []string, status int) {
	reportKey := &ReportOpenFalConKey{
		Endpoint: endpoint,
		Metric: metric,
		Tags: tags,
	}
	reporter.status.setStatus(reportKey, status)
}

func (reporter *OpenFalconReporter) IncCounter(endpoint, metric string, tags []string, value float64) {
	reportKey := &ReportOpenFalConKey{
		Endpoint: endpoint,
		Metric: metric,
		Tags: tags,
	}
	reporter.counter.inc(reportKey, value)
}

func (reporter *OpenFalconReporter) IncNumerator(endpoint, metric string, tags []string, value float64) {
	reportKey := &ReportOpenFalConKey{
		Endpoint: endpoint,
		Metric: metric,
		Tags: tags,
	}
	reporter.percent.incNumerator(reportKey, value)
}

func (reporter *OpenFalconReporter) IncDenominator(endpoint, metric string, tags []string, value float64) {
	reportKey := &ReportOpenFalConKey{
		Endpoint: endpoint,
		Metric: metric,
		Tags: tags,
	}
	reporter.percent.incDenominator(reportKey, value)
}

func (reporter *OpenFalconReporter) IncGauge(endpoint, metric string, tags []string, value float64) {
	reportKey := &ReportOpenFalConKey{
		Endpoint: endpoint,
		Metric: metric,
		Tags: tags,
	}
	reporter.gauge.inc(reportKey, value)
}

func (reporter *OpenFalconReporter) AddTiming(endpoint, metric string, tags []string, value time.Duration) {
	reportKey := &ReportOpenFalConKey{
		Endpoint: endpoint,
		Metric: metric,
		Tags: tags,
	}
	reporter.timing.add(reportKey, value)
}

func (reporter *OpenFalconReporter) Start(postUrl string) {
	reporter.startMutex.Lock()
	if reporter.isStart {
		reporter.startMutex.Unlock()
		return
	}

	reporter.isStart = true
	reporter.startMutex.Unlock()

	reporter.runReporter()
	open_falcon_sender.PostPushUrl = postUrl //
	open_falcon_sender.StartSender()

	go func() {
		for {
			time.Sleep(time.Minute)
			reporter.pushCounter()
			reporter.pushGauge()
			reporter.pushTiming()
			reporter.pushPercent()
			reporter.pushStatus()
		}
	}()
}

func (reporter *OpenFalconReporter) pushCounter() {
	all := reporter.counter.getAllMetaData()
	for _, data := range all {
		reporter.senderQueue.PushFront(data)
	}
}

func (reporter *OpenFalconReporter) pushGauge() {
	all := reporter.gauge.getAllMetaData()
	for _, data := range all {
		reporter.senderQueue.PushFront(data)
	}
}

func (reporter *OpenFalconReporter) pushTiming() {
	all := reporter.timing.getAllMetaData()
	for _, data := range all {
		reporter.senderQueue.PushFront(data)
	}
}

func (reporter *OpenFalconReporter) pushPercent() {
	all := reporter.percent.getAllMetaData()
	for _, data := range all {
		reporter.senderQueue.PushFront(data)
	}
}

func (reporter *OpenFalconReporter) pushStatus() {
	all := reporter.status.getAllMetaData()
	for _, data := range all {
		reporter.senderQueue.PushFront(data)
	}
}
