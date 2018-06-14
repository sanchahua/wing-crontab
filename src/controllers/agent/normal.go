package agent

import (
	"library/data"
)

type QEs map[int64]*data.EsQueue
func (queueNomal *QEs) append(item *runItem) bool {
	normal, ok := (*queueNomal)[item.id]
	if !ok {
		normal = data.NewQueue(maxQueueLen)
		(*queueNomal)[item.id] = normal
	}
	ok , _ = normal.Put(item)
	return ok
}

func (queueNomal *QEs) dispatch(msgId int64, id int64, address string, send sendFunc, c chan *SendData, success func(item *runItem)){
	queueNormal := (*queueNomal)[id]
	itemI, ok, _ := queueNormal.Get()
	if !ok || itemI == nil {
		return
	}
	item := itemI.(*runItem)
	success(item)
	c <- newSendData(msgId, CMD_RUN_COMMAND, item, send, item.id, item.isMutex, address)
}
