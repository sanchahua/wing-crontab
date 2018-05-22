package agent

import (
	"library/data"
	"library/agent"
)

type QEs map[int64]*data.EsQueue
func (queueNomal *QEs) append(item *runItem){
	normal, ok := (*queueNomal)[item.id]
	if !ok {
		normal = data.NewQueue(maxQueueLen)
		(*queueNomal)[item.id] = normal
	}
	normal.Put(item)
}

func (queueNomal *QEs) dispatch(id int64, address string, send func(data []byte), c chan *SendData){
	queueNormal := (*queueNomal)[id]
	itemI, ok, _ := queueNormal.Get()
	if !ok || itemI == nil {
		//log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
		return
	}
	item := itemI.(*runItem)
	sendData := pack(item, address)//c.ctx.Config.BindAddress)
	c <- newSendData(agent.CMD_RUN_COMMAND, sendData, send, item.id, item.isMutex, item.logId)
}
