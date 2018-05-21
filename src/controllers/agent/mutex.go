package agent

import (
	"library/data"
	log "github.com/sirupsen/logrus"
)

type Mutex struct {
	isRuning bool
	queue *data.EsQueue
	start int64
}

type QMutex map[int64]*Mutex
func (queueMutex *QMutex) append(item *runItem) {
	mutex, ok := (*queueMutex)[item.id]
	if !ok {
		mutex = &Mutex{
			isRuning: false,
			queue:    data.NewQueue(maxQueueLen),
			start:    0,
		}
		(*queueMutex)[item.id] = mutex
	}

	ok, num := mutex.queue.Put(item)
	log.Debugf("queue len %v", num)

	if !ok {
		log.Errorf("put error %v, %v", ok, num)
	}
	log.Debugf("add item to queueMutex, current len %v", len(*queueMutex))
}

