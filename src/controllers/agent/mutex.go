package agent

import (
	"library/data"
	log "github.com/sirupsen/logrus"
	"time"
	"library/agent"
	"fmt"
	"os"
)

type Mutex struct {
	isRuning bool
	queue *data.EsQueue
	start int64
	sta *Statistics
	isMutex bool
}

type QMutex map[int64]*Mutex
func (queueMutex *QMutex) append(item *runItem) {
	mutex, ok := (*queueMutex)[item.id]
	if !ok {
		mutex = &Mutex{
			isRuning: false,
			queue:    data.NewQueue(maxQueueLen),
			start:    0,
			sta:      &Statistics{},
			isMutex:  item.isMutex,
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

func (queueMutex *QMutex) dispatch(gindexMutex *int64,
	address string, send func(data []byte), c chan *SendData) {
	start := time.Now()
	{
		indexMutex := int64(-1)
		if *gindexMutex >= int64(len(*queueMutex)-1) {
			*gindexMutex = 0
		}
		log.Debugf("queue mutex len %v", *queueMutex)
		for id, queue := range *queueMutex {
			log.Debugf("queue mutex queue len %v", queue.queue.Quantity())
			indexMutex++
			// 如果有未完成的任务，跳过
			// 这里的正在运行应该有一个超时时间
			// 一般情况下用不着，仅仅为了预防，提高可靠性
			// 最多锁定60秒

			// 获取平均原型周期 + 60s最为超时标准
			var timeout = queue.getTimeout()
			log.Debugf("%v queueMutex avg timeout is %v", id, timeout)

			//c.statisticsLock.Lock()
			//var timeout int64 = 60 * 1000
			//sta, ok := c.statistics[id]
			//if ok {
			//	avg := sta.getAvg()
			//	if avg > 0 {
			//		timeout = avg * 3
			//		if timeout > avg + 60 * 1000 {
			//			timeout = avg + 60 * 1000
			//		} else if timeout < 300 {
			//			timeout = 1000
			//		}
			//	}
			//}
			//c.statisticsLock.Unlock()

			if queue.isRuning && (int64(time.Now().UnixNano()/1000000) - queue.start) < timeout {
				log.Debugf("================%v still running", id)
				continue
			}
			if indexMutex >= *gindexMutex {

				(*gindexMutex)++
				itemI, ok, _ := queue.queue.Get()
				if !ok || itemI == nil {
					//log.Warnf("queue get empty, %+v, %+v, %+v", ok, itemI)
					continue
				}
				queue.isRuning = true
				queue.start = int64(time.Now().UnixNano()/1000000)//time.Now().Unix()
				item := itemI.(*runItem)
				//分发互斥定时任务
				sendData := pack(item, address)//c.ctx.Config.BindAddress)

				c <- newSendData(agent.CMD_RUN_COMMAND, sendData, /*node.AsyncSend*/send, item.id, item.isMutex)
				//log.Debugf("###########dispatch mutex : %+v", *d)
				//c.sendQueueLock.Lock()
				//c.sendQueue[d.Unique] = d
				//c.sendQueueLock.Unlock()
				//c.sendQueueChan <- d
				//c <- d
				break
			}

		}
	}
	fmt.Fprintf(os.Stderr, "OnPullCommand mutex use time %v\n", time.Since(start))

}

func (queueMutex *QMutex) setRunning(id int64, running bool) {
	m ,ok := (*queueMutex)[id]
	if ok {
		log.Debugf("##################set %v running is %v", id, running)
		m.isRuning = running
	} else {
		log.Errorf("%v does not exists")
	}
}

func (queue *Mutex) getTimeout() int64 {

	var timeout int64 = 60 * 1000
	avg := queue.sta.getAvg()
	if avg > 0 {
		timeout = avg * 3
		if timeout > avg + 60 * 1000 {
			timeout = avg + 60 * 1000
		} else if timeout < 300 {
			timeout = 1000
		}
	}
	return timeout
}
