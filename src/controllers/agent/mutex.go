package agent

import (
	"library/data"
)

type Mutex struct {
	isRuning bool
	queue *data.EsQueue
	start int64
}

