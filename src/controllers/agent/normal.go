package agent

import (
	"library/data"
	log "github.com/sirupsen/logrus"
)

type QEs map[int64]*data.EsQueue
func (queueNomal *QEs) append(item *runItem){
	normal, ok := (*queueNomal)[item.id]
	if !ok {
		normal = data.NewQueue(maxQueueLen)
		(*queueNomal)[item.id] = normal
	}
	//item := &runItem{id: id, command: command, isMutex: isMutex,}
	ok, num := normal.Put(item)
	log.Debugf("queue len %v", num)
	if !ok {
		log.Errorf("put error %v, %v", ok, num)
	}
	log.Debugf("add item to queueMutex, current len %v", len(*queueNomal))
}
