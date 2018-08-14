package main

import (
	"time"
	"gitlab.xunlei.cn/xlsoa/common/open_falcon_sdk/reporter"
	"math/rand"
)

func stat()  {

	//统计计数类型，数值一直累加，上报给open falcon为COUNTER类型
	//模拟统计上报请求数
	open_falcon_reporter.Reporter.IncCounter("endpoint_myServiceName", "metric_request_cnt", []string{"tag0=t0", "tag1=t1"}, 1)

	//统计gauge类型，数值一直累加，上报给open falcon为GAUGE类型，一旦上报后清零
	//模拟统计上报错误数
	open_falcon_reporter.Reporter.IncGauge("endpoint_myServiceName", "metric_error_cnt", []string{"tag0=t0", "tag1=t1"}, 1)

	//统计时间类型，统计一定周期内的duration累加值，上报给open falcon为GAUGE类型，上报的数值为周期内的平均值
	//模拟统计响应时间，上报给open falcon为平均响应时间
	open_falcon_reporter.Reporter.AddTiming("endpoint_myServiceName", "metric_avg_response_time", []string{"tag0=t0", "tag1=t1"}, time.Millisecond*200)

	//统计百分比类型，统计一定周期内的duration累加值, 上报给open falcon为GAUGE类型，上报的数值为周期内的分子/分母的值
	//模拟统计上报错误率: 50%错误率(1次分子，2次分母)
	open_falcon_reporter.Reporter.IncNumerator("endpoint_myServiceName", "metric_error_rate", []string{"tag0=t0", "tag1=t1"}, 1) //累计分子的值
	open_falcon_reporter.Reporter.IncDenominator("endpoint_myServiceName", "metric_error_rate", []string{"tag0=t0", "tag1=t1"}, 1)//累计分母的值
	open_falcon_reporter.Reporter.IncDenominator("endpoint_myServiceName", "metric_error_rate", []string{"tag0=t0", "tag1=t1"}, 1)//累计分母的值

	//上报当前状态(该状态应该是一个外部模块定时更新), 该值会原封不动的上报给open falcon，上报给open falcon为GAUGE类型
	//模拟定时更新当前安全降级的状态
	SecurityDegradeActivated := 1
	go func() {//模拟定时变更状态
		SecurityDegradeActivated = rand.Int() % 2
		time.Sleep(time.Second * 50)
	}()
	//设置当前安全降级的状态，并自动上报
	open_falcon_reporter.Reporter.SetStatus("endpoint_myServiceName", "metric_SecurityDegradeActivated", []string{"tag0=t0", "tag1=t1"}, SecurityDegradeActivated)
}

func main()  {
	//启动open falcon的上报进程
	open_falcon_reporter.Reporter.Start("http://10.10.131.101:1988/v1/push")
	for i := 0; i<100000; i++ {
		stat()
		time.Sleep(time.Second)
	}
}
