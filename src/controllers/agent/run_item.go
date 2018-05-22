package agent

type runItem struct {
	id int64
	command string
	isMutex bool
	logId int64
	setWaitNum func(int64)
}
