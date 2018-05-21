package agent

import (
	"net"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"sync"
	"context"
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
	connLock *sync.Mutex
	statusLock *sync.Mutex
	status int
	getLeader GetLeaderFunc
	dataChannel chan *dataItem
	onEvents []OnClientEventFunc
	asyncWriteChan chan []byte
}

type GetLeaderFunc     func()(string, int, error)
type ClientOption      func(tcp *AgentClient)
type OnCommandFunc     func(content []byte)
type OnClientEventFunc func(tcp *AgentClient, event int, content []byte)

func SetGetLeader(f GetLeaderFunc) ClientOption {
	return func(tcp *AgentClient) {
		tcp.getLeader = f
	}
}

func SetOnClientEvent(f ...OnClientEventFunc) ClientOption {
	return func(tcp *AgentClient) {
		tcp.onEvents = append(tcp.onEvents, f...)
	}
}

const asyncWriteChanLen = 10000

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
		bufferLock:    new(sync.Mutex),
		dataChannel:   make(chan *dataItem, dataChannelLen),
		onEvents:      make([]OnClientEventFunc, 0),
		asyncWriteChan:make(chan []byte, asyncWriteChanLen),
		connLock:      new(sync.Mutex),
	}
	for _, f := range opts {
		f(c)
	}
	go c.keepalive()
	go c.asyncWrite()
	return c
}

// 直接发送
func (tcp *AgentClient) Write(data []byte) {
	start := time.Now().Unix()
	for {
		if (time.Now().Unix() - start) > 3 {
			log.Errorf("asyncWriteChan full, wait timeout")
			return
		}
		if len(tcp.asyncWriteChan) < cap(tcp.asyncWriteChan) {
			break
		}
		log.Warnf("asyncWriteChan full")
	}
	tcp.asyncWriteChan <- data
}

func (tcp *AgentClient) asyncWrite() {
	for {
		select {
		case data, ok := <- tcp.asyncWriteChan:
			if !ok {
				return
			}
			//tcp.connLock.Lock()
			if tcp.conn != nil {
				//log.Debugf("##########send data: %+v", data)
				n, err := tcp.conn.Write(data)
				if err != nil {
					log.Errorf("send failure: %+v", err)
				}
				if n < len(data) {
					log.Errorf("send not complete")

				}
			}
			//tcp.connLock.Unlock()
		}
	}
}

func (tcp *AgentClient) keepalive() {
	data := Pack(CMD_TICK, []byte(""))
	for {
		tcp.Write(data)
		time.Sleep(3 * time.Second)
	}
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
	//tcp.connLock.Lock()
	//defer tcp.connLock.Unlock()
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
	var readBuffer [4096]byte
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
				//start := time.Now()
				if tcp.conn == nil {
					log.Errorf("============================tcp conn nil")
					break
				}
				//start3 := time.Now()
				size, err := tcp.conn.Read(readBuffer[0:])
				//fmt.Fprintf(os.Stderr, "read use time %v\n", time.Since(start3))

				if err != nil || size <= 0 {
					log.Warnf("agent read with error: %+v", err)
					tcp.statusLock.Lock()
					tcp.disconnect()
					tcp.statusLock.Unlock()
					break
				}
				tcp.onMessage(readBuffer[:size])

				select {
				case <-tcp.ctx.Done():
					return
				default:
				}
				//fmt.Fprintf(os.Stderr, "read message use time %v\n", time.Since(start))
			}
		}
	}()
}

func (tcp *AgentClient) onMessage(msg []byte) {
	tcp.bufferLock.Lock()
	tcp.buffer = append(tcp.buffer, msg...)
	tcp.bufferLock.Unlock()

	for {
		bufferLen := len(tcp.buffer)
		if bufferLen < 6 {
			return
		}
		if bufferLen > MAX_PACKAGE_LEN {
			log.Errorf("buffer len is max then the limit %+v", MAX_PACKAGE_LEN)
			tcp.bufferLock.Lock()
			tcp.buffer = make([]byte, 0)
			tcp.bufferLock.Unlock()
			return
		}
		tcp.bufferLock.Lock()
		cmd, content, err := Unpack(&tcp.buffer)
		tcp.bufferLock.Unlock()
		if err != nil {
			return
		}

		if !hasCmd(cmd) {
			log.Errorf("cmd %d dos not exists", cmd)
			tcp.bufferLock.Lock()
			tcp.buffer = make([]byte, 0)
			tcp.bufferLock.Unlock()
			return
		}

		for _, f := range tcp.onEvents {
			go f(tcp, cmd, content)
		}
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

