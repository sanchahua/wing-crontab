package agent

import (
	"library/data"
	"library/agent"
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

func (queueNomal *QEs) dispatch(id int64, address string, send func(data []byte) (int, error), c chan *SendData, success func(num uint32)){
	queueNormal := (*queueNomal)[id]
	itemI, ok, _ := queueNormal.Get()
	if !ok || itemI == nil {
		//log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
		return
	}
	item := itemI.(*runItem)
	sendData := pack(item, address)//c.ctx.Config.BindAddress)
	success(queueNormal.Quantity())
	c <- newSendData(agent.CMD_RUN_COMMAND, sendData, send, item.id, item.isMutex)
}
