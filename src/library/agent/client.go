package agent

import (
	"net"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"encoding/json"
	"sync"
	"context"
	wstring "library/string"
)

type dataItem struct {
	cmd int
	content []byte
}
const dataChannelLen=10000
type AgentClient struct {
	ctx context.Context
	buffer  []byte
	bufferLock *sync.Mutex
	conn     *net.TCPConn
	statusLock *sync.Mutex
	status int
	getLeader GetLeaderFunc
	sendQueue map[string]*SendData
	sendQueueLock *sync.Mutex
	onCommand OnCommandFunc
	dataChannel chan *dataItem
}

type GetLeaderFunc     func()(string, int, error)
type ClientOption      func(tcp *AgentClient)
type OnCommandFunc     func(content []byte)

func SetGetLeader(f GetLeaderFunc) ClientOption {
	return func(tcp *AgentClient) {
		tcp.getLeader = f
	}
}

func SetOnCommand(f OnCommandFunc) ClientOption {
	return func(tcp *AgentClient) {
		tcp.onCommand = f
	}
}

type SendData struct {
	Unique string `json:"unique"`
	Data []byte `json:"data"`
	Status int `json:"status"`
	Time int64 `json:"time"`
	SendTimes int `json:"send_times"`
}

func newSendData(data []byte) *SendData {
	return &SendData{
		Unique:wstring.RandString(128),
		Data: data,
		Status: 0,
		Time: 0,
		SendTimes:0,
	}

}

func (d *SendData) encode() []byte {
	b, e := json.Marshal(d)
	if e != nil {
		return nil
	}
	return b
}

// client 用来接收 agent server 分发的定时任务事件
// 接收到事件后执行指定的定时任务
// onleader 触发后，如果是leader，client停止
// 如果不是leader，client查询到leader的服务地址，连接到server
func NewAgentClient(ctx context.Context, opts ...ClientOption) *AgentClient {
	c := &AgentClient{
		ctx:           ctx,
		buffer:        make([]byte, 0),
		conn:          nil,
		statusLock:    new(sync.Mutex),
		status:        0,
		sendQueue:     make(map[string]*SendData),
		sendQueueLock: new(sync.Mutex),
		bufferLock:    new(sync.Mutex),
		dataChannel:   make(chan *dataItem, dataChannelLen),
	}
	for _, f := range opts {
		f(c)
	}
	go c.keepalive()
	go c.sendService()
	//cpu := runtime.NumCPU()
	//for i := 0;i<cpu;i++ {
	//	go c.onData()
	//}
	return c
}

// must send success
func (tcp *AgentClient) Send(data []byte) {
	d := newSendData(data)
	tcp.sendQueueLock.Lock()
	tcp.sendQueue[d.Unique] = d
	tcp.sendQueueLock.Unlock()
}

func (tcp *AgentClient) sendService() {
	for {
		select {
			case <-tcp.ctx.Done():
				log.Debugf("keepalive exit 1")
				return
			default:
		}
		tcp.statusLock.Lock()
		if tcp.conn == nil || tcp.status & agentStatusConnect <= 0 {
			tcp.statusLock.Unlock()
			//log.Infof("keepalive continue")
			time.Sleep(3 * time.Second)
			continue
		}
		tcp.statusLock.Unlock()

		tcp.sendQueueLock.Lock()
		for _, d := range tcp.sendQueue {
			// status > 0 is sending
			if d.Status > 0 && (time.Now().Unix() - d.Time) <= 3 {
				continue
			}
			log.Infof("try to send %+v", *d)
			d.Status = 1
			d.SendTimes++

			if d.SendTimes >= 36 {
				delete(tcp.sendQueue, d.Unique)
				log.Warnf("send timeout(36s), delete %+v", *d)
				continue
			}
			d.Time   = time.Now().Unix()
			sd      := d.encode()
			log.Infof("try to send %+v", sd)

			data    := Pack(CMD_CRONTAB_CHANGE, sd)
			dl      := len(data)
			n, err  := tcp.conn.Write(data)

			if err != nil {
				log.Errorf("[agent - client] agent keepalive error: %d, %v", n, err)
				tcp.statusLock.Lock()
				tcp.disconnect()
				tcp.statusLock.Unlock()
			} else if n != dl {
				log.Errorf("[agent - client] %s send not complete", tcp.conn.RemoteAddr().String())
			}
		}
		tcp.sendQueueLock.Unlock()
		time.Sleep(time.Second * 1)
	}
}

func (tcp *AgentClient) keepalive() {
	data := Pack(CMD_TICK, []byte(""))
	dl   := len(data)
	for {
		select {
			case <-tcp.ctx.Done():
				log.Debugf("keepalive exit 1")
				return
			default:
		}
		tcp.statusLock.Lock()
		if tcp.conn == nil || tcp.status & agentStatusConnect <= 0 {
			tcp.statusLock.Unlock()
			//log.Infof("keepalive continue")
			time.Sleep(3 * time.Second)
			continue
		}
		tcp.statusLock.Unlock()
		n, err := tcp.conn.Write(data)
		if err != nil {
			log.Errorf("[agent - client] agent keepalive error: %d, %v", n, err)
			tcp.statusLock.Lock()
			tcp.disconnect()
			tcp.statusLock.Unlock()
		} else if n != dl {
			log.Errorf("[agent - client] %s send not complete", tcp.conn.RemoteAddr().String())
		}
		//log.Infof("client keepalive")
		time.Sleep(3 * time.Second)
	}
	log.Debugf("keepalive exit 2")
}

func (tcp *AgentClient) OnLeader(leader bool) {
	go func() {
		log.Debugf("==============agent client OnLeader %v===============", leader)
		var ip string
		var port int
		for {
			ip, port, _ = tcp.getLeader()
			if ip == "" || port <= 0 {
				log.Warnf("ip or port empty: %v, %v, wait for init", ip, port)
				time.Sleep(time.Second * 1)
				continue
			}
			break
		}
		log.Infof("leader %v:%v", ip, port)
		tcp.start(ip, port)
	}()
}

func (tcp *AgentClient) connect(ip string, port int) {
	if tcp.conn != nil {
		tcp.statusLock.Lock()
		tcp.disconnect()
		tcp.statusLock.Unlock()
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Errorf("start agent with error: %+v", err)
		tcp.conn = nil
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Errorf("start agent with error: %+v", err)
		tcp.conn = nil
		return
	}
	tcp.conn = conn
}

func (tcp *AgentClient) start(serviceIp string, port int) {
	var readBuffer [128]byte
	go func() {
		if serviceIp == "" || port == 0 {
			log.Warnf("ip or port empty %s:%d", serviceIp, port)
			return
		}

		tcp.statusLock.Lock()
		if tcp.status & agentStatusConnect > 0 {
			tcp.statusLock.Unlock()
			return
		}
		tcp.statusLock.Unlock()

		for {
			select {
				case <-tcp.ctx.Done():
					return
				default:
			}

			tcp.connect(serviceIp, port)
			if tcp.conn == nil {
				time.Sleep(time.Second * 3)
				// if connect error, try to get leader agein
				for {
					serviceIp, port, _ = tcp.getLeader()
					if serviceIp == "" || port <= 0 {
						log.Warnf("ip or port empty: %v, %v, wait for init", serviceIp, port)
						time.Sleep(time.Second * 1)
						continue
					}
					break
				}
				continue
			}
			tcp.statusLock.Lock()
			if tcp.status & agentStatusConnect <= 0 {
				tcp.status |= agentStatusConnect
			}
			tcp.statusLock.Unlock()

			log.Debugf("====================agent client connect to leader %s:%d====================", serviceIp, port)

			for {
				start := time.Now()
				if tcp.conn == nil {
					log.Errorf("============================tcp conn nil")
					break
				}
				size, err := tcp.conn.Read(readBuffer[0:])
				log.Debugf("read message use time %v", time.Since(start))
				//log.Debugf("read buffer len: %d, cap:%d", len(readBuffer), cap(readBuffer))
				if err != nil || size <= 0 {
					log.Warnf("agent read with error: %+v", err)
					tcp.statusLock.Lock()
					tcp.disconnect()
					tcp.statusLock.Unlock()
					break
				}
				//log.Debugf("agent receive %d bytes: %+v, %s", size, readBuffer[:size], string(readBuffer[:size]))
				start = time.Now()
				tcp.onMessage(readBuffer[:size])
				log.Debugf("#################################on message use time %+v", time.Since(start))
				select {
				case <-tcp.ctx.Done():
					return
				default:
				}
			}
		}
	}()
}

//func (tcp *AgentClient) onData() {
//	for {
//		select {
//		case data, ok := <- tcp.dataChannel:
//			if !ok {
//				return
//			}
//
//			switch data.cmd {
//			case CMD_TICK:
//				//keepalive
//				//log.Info("keepalive")
//			case CMD_CRONTAB_CHANGE:
//				unique := string(data.content)
//				log.Infof("%v send ok, delete from send queue", unique)
//				//tcp.sendQueueLock.Lock()
//				delete(tcp.sendQueue, unique)
//				//tcp.sendQueueLock.Unlock()
//			case CMD_RUN_COMMAND:
//				//id := binary.LittleEndian.Uint64(content[:8])
//				//log.Debugf("id == (%v) === (%v) ", id, content[:8])
//				//log.Debugf("content == (%v) === (%v) ", string(content[8:]), content[:8])
//				start := time.Now()
//				tcp.onCommand(data.content)//int64(id), string(content[8:]))
//				log.Debugf("onCommand use time %v", time.Since(start))
//			default:
//			}
//		}
//	}
//}

func (tcp *AgentClient) onMessage(msg []byte) {
	//start := time.Now()
	tcp.buffer = append(tcp.buffer, msg...)
	//log.Debugf("#################append use time %v", time.Since(start))

	for {
		//start = time.Now()
		bufferLen := len(tcp.buffer)
		if bufferLen < 6 {
			return
		}
		if bufferLen > MAX_PACKAGE_LEN {
			log.Errorf("buffer len is max then the limit %+v", MAX_PACKAGE_LEN)
			tcp.buffer = make([]byte, 0)
			return
		}
		//log.Debugf("#################check use time %v", time.Since(start))

		//start = time.Now()
		cmd, content, err := Unpack(&tcp.buffer)
		if err != nil {
			return
		}
		//log.Debugf("#################Unpack use time %v", time.Since(start))
		//start = time.Now()

		if !hasCmd(cmd) {
			log.Errorf("cmd %d dos not exists", cmd)
			tcp.buffer = make([]byte, 0)
			return
		}
		//log.Debugf("#################hasCmd use time %v", time.Since(start))

		//tcp.dataChannel <- &dataItem{
		//	cmd : cmd,
		//	content:content,
		//}
		//start = time.Now()

		switch cmd {
		case CMD_TICK:
			//keepalive
			//log.Info("keepalive")
		case CMD_CRONTAB_CHANGE:
			unique := string(content)
			//log.Infof("%v send ok, delete from send queue", unique)
			//tcp.sendQueueLock.Lock()
			delete(tcp.sendQueue, unique)
			//tcp.sendQueueLock.Unlock()
		case CMD_RUN_COMMAND:
			//id := binary.LittleEndian.Uint64(content[:8])
			//log.Debugf("id == (%v) === (%v) ", id, content[:8])
			//log.Debugf("content == (%v) === (%v) ", string(content[8:]), content[:8])
			//start2 := time.Now()
			tcp.onCommand(content)//int64(id), string(content[8:]))
			//log.Debugf("#################onCommand use time %v", time.Since(start2))
		default:
		}
		//log.Debugf("#################switch use time %v", time.Since(start))

	}
}

func (tcp *AgentClient) disconnect() {
	if tcp.conn == nil || tcp.status & agentStatusConnect <= 0 {
		log.Debugf("agent is in disconnect status")
		return
	}
	log.Warnf("====================agent disconnect====================")
	tcp.conn.Close()
	tcp.conn = nil
	if tcp.status & agentStatusConnect > 0 {
		tcp.status ^= agentStatusConnect
	}
}

