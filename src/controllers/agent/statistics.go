package agent

type Statistics struct {
	sendTimes int64
	totalUseTime int64
	startTime int64
}

func (s *Statistics) getAvg() int64 {
	if s.sendTimes > 0 {
		return int64(s.totalUseTime/s.sendTimes)
	}
	return 0
}
