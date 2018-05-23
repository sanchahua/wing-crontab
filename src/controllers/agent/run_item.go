package agent

type runItem struct {
	id int64
	command string
	isMutex bool
	subWaitNum func() int64
}
