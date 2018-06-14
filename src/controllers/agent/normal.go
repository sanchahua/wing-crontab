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

func (queueNomal *QEs) dispatch(id int64, success func(item *runItem)){
	queueNormal := (*queueNomal)[id]
	itemI, ok, _ := queueNormal.Get()
	if !ok || itemI == nil {
		return
	}
	item := itemI.(*runItem)
	success(item)
}
