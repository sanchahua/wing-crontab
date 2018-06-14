package agent

import (
	"library/data"
	log "github.com/sirupsen/logrus"
	"time"
)

type Mutex struct {
	isRuning bool
	queue *data.EsQueue
	start int64
	sta *Statistics
	isMutex bool
}

type QMutex map[int64]*Mutex
func (queueMutex *QMutex) append(item *runItem) bool {
	mutex, ok := (*queueMutex)[item.Id]
	if !ok {
		mutex = &Mutex{
			isRuning: false,
			queue:    data.NewQueue(maxQueueLen),
			start:    0,
			sta:      &Statistics{},
			isMutex:  item.IsMutex,
		}
		(*queueMutex)[item.Id] = mutex
	}
	ok, _ = mutex.queue.Put(item)
	return ok
}

func (queueMutex *QMutex) dispatch(id int64, success func(*runItem)) {
	log.Infof("%+v mutex dispatch", id)
	queue := (*queueMutex)[id]
	var timeout = queue.getTimeout()
	tn := int64(time.Now().UnixNano()/1000000)
	if queue.isRuning && (tn - queue.start) < timeout {
		log.Debugf("================%v still running", id)
		return
	}
	itemI, ok, _ := queue.queue.Get()
	if !ok || itemI == nil {
		//log.Warnf("queue get empty, %+v, %+v, %+v", ok, itemI)
		return
	}
	queue.isRuning = true
	queue.start = tn//int64(time.Now().UnixNano()/1000000)//time.Now().Unix()
	item := itemI.(*runItem)
	//分发互斥定时任务
	//sendData := pack(item, address)//c.ctx.Config.BindAddress)
	success(item)
	//c <- newSendData(msgId, CMD_RUN_COMMAND, item, send, item.id, item.isMutex, address)
}

func (queueMutex *QMutex) setRunning(id int64, running bool) {
	m ,ok := (*queueMutex)[id]
	if ok {
		//log.Debugf("##################set %v running is %v", id, running)
		m.isRuning = running
	} else {
		log.Errorf("%v does not exists", id)
	}
}

func (queue *Mutex) setRunning(running bool) {
	queue.isRuning = running
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
