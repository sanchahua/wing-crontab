package agent

import (
	"library/data"
	log "github.com/sirupsen/logrus"
	"time"
	"library/agent"
	"fmt"
	"os"
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

func (queueNomal *QEs) dispatch(gindexNormal *int64, address string, send func(data []byte), c chan *SendData){
	start := time.Now()
	{
		index := int64(-1)
		if *gindexNormal >= int64(len(*queueNomal)-1) {
			*gindexNormal = 0
		}

		for _, queueNormal := range *queueNomal {
			index++
			if index != *gindexNormal {
				continue
			}
			(*gindexNormal)++
			itemI, ok, _ := queueNormal.Get()
			if !ok || itemI == nil {
				//log.Warnf("queue get empty, %+v, %+v, %+v", ok, num, itemI)
				continue
			}
			item := itemI.(*runItem)
			sendData := pack(item, address)//c.ctx.Config.BindAddress)

			c <- newSendData(agent.CMD_RUN_COMMAND, sendData, /*node.AsyncSend*/send, item.id, item.isMutex) //c.server.Broadcast)//
			//c.sendQueueLock.Lock()
			//c.sendQueue[d.Unique] = d
			//c.sendQueueLock.Unlock()
			//c.sendQueueChan <- d

			break
		}
	}
	fmt.Fprintf(os.Stderr, "OnPullCommand normal use time %v\n", time.Since(start))
}
