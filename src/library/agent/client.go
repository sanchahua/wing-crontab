package agent

import (
	"net"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"sync"
	"context"
	"errors"
)

type Client struct {
	ctx context.Context
	buffer  []byte
	conn     *net.TCPConn
	connLock *sync.Mutex
	statusLock *sync.Mutex
	status int
	getLeader GetLeaderFunc
	onEvents []OnClientEventFunc
	asyncWriteChan chan []byte
	checkChan chan address
}
type address struct {
	ip string
	port int
}
type GetLeaderFunc     func()(string, int, error)
type ClientOption      func(tcp *Client)
type OnCommandFunc     func(content []byte)
type OnClientEventFunc func(tcp *Client, event int, content []byte)

func SetGetLeader(f GetLeaderFunc) ClientOption {
	return func(tcp *Client) {
		tcp.getLeader = f
	}
}

func SetOnClientEvent(f ...OnClientEventFunc) ClientOption {
	return func(tcp *Client) {
		tcp.onEvents = append(tcp.onEvents, f...)
	}
}

const asyncWriteChanLen = 10000

// client 用来接收 agent server 分发的定时任务事件
// 接收到事件后执行指定的定时任务
// onleader 触发后，如果是leader，client停止
// 如果不是leader，client查询到leader的服务地址，连接到server
func NewClient(ctx context.Context, opts ...ClientOption) *Client {
	c := &Client{
		ctx:           ctx,
		buffer:        make([]byte, 0),
		conn:          nil,
		statusLock:    new(sync.Mutex),
		status:        0,
		onEvents:      make([]OnClientEventFunc, 0),
		asyncWriteChan:make(chan []byte, asyncWriteChanLen),
		connLock:      new(sync.Mutex),
		checkChan:     make(chan address),
	}
	for _, f := range opts {
		f(c)
	}
	go c.keepalive()
	go c.asyncWrite()
	go c.check()
	return c
}

// 直接发送
func (tcp *Client) AsyncWrite(data []byte) {
	tcp.asyncWriteChan <- data
}

var notConnect = errors.New("not connect")

func (tcp *Client) Write(data []byte) (int, error) {
	if tcp.status & agentStatusConnect <= 0 {
		return 0, notConnect
	}
	return tcp.conn.Write(data)
}

func (tcp *Client) asyncWrite() {
	for {
		select {
		case data, ok := <- tcp.asyncWriteChan:
			if !ok {
				return
			}

			n, err := tcp.Write(data)
			if err != nil {
				log.Errorf("send failure: %+v", err)
			}
			if n < len(data) {
				log.Errorf("send not complete")

			}
		}
	}
}

func (tcp *Client) keepalive() {
	data := Pack(CMD_TICK, []byte(""))
	for {
		tcp.Write(data)
		time.Sleep(3 * time.Second)
	}
}

func (tcp *Client) connect(ip string, port int) {
	if ip == "" || port <= 0 {
		return
	}
	tcp.disconnect()
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Errorf("start agent with error: %+v", err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Errorf("start agent with error: %+v", err)
		return
	}
	if tcp.status & agentStatusConnect <= 0 {
		tcp.status |= agentStatusConnect
	}
	tcp.conn = conn
}

func (tcp *Client) check() {
	//var cc = ""
	var c = make(chan struct{})
	go func() {
		for {
			c <- struct{}{}
			time.Sleep(time.Second * 3)
		}
	}()

	var serviceIp = ""
	var port = 0

	for {
		select {
		case ad, ok := <- tcp.checkChan:
			if !ok {
				return
			}

			if serviceIp != "" && port > 0 {
				if serviceIp != ad.ip || port != ad.port {
					log.Warnf("leader change found")
					//如果服务地址端口发生改变
					tcp.disconnect()
					serviceIp = ad.ip
					port = ad.port
				}
			} else {
				serviceIp = ad.ip
				port = ad.port
				tcp.start()
			}

		case <- c :
			s, p, _ := tcp.getLeader()
			if serviceIp != "" && port > 0 {
				if serviceIp != s || port != p {
					log.Warnf("self check, leader change found")
					//如果服务地址端口发生改变
					tcp.disconnect()
					serviceIp = s//ad.ip
					port = p//ad.port
				}
			}

		}

	}
}

func (tcp *Client) Start(serviceIp string, port int)  {
	tcp.checkChan <- address{serviceIp, port}
}

func (tcp *Client) start() {
	if tcp.status & agentStatusConnect > 0 {
		return
	}

	go func() {
		for {
			select {
				case <-tcp.ctx.Done():
					return
				default:
			}
			serviceIp, port, _ := tcp.getLeader()
			tcp.connect(serviceIp, port)
			if tcp.status & agentStatusConnect <= 0 {
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

			//tcp.checkChan <- fmt.Sprintf("%v:%v", serviceIp, port)

			log.Debugf("====================agent client connect to leader %s:%d====================", serviceIp, port)

			for {
				if tcp.status & agentStatusConnect <= 0  {
					break
				}
				readBuffer := make([]byte, 4096)
				size, err  := tcp.conn.Read(readBuffer)

				if err != nil || size <= 0 {
					log.Warnf("agent read with error: %+v", err)
					tcp.disconnect()
					break
				}
				tcp.onMessage(readBuffer[:size])

				select {
				case <-tcp.ctx.Done():
					return
				default:
				}
			}
		}
	}()
}

func (tcp *Client) onMessage(msg []byte) {

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Unpack recover##########%+v, %+v", err, tcp.buffer)
			tcp.buffer = make([]byte, 0)
		}
	}()

	tcp.buffer = append(tcp.buffer, msg...)
	for {
		cmd, content, pos, err := Unpack(tcp.buffer)
		if err != nil {
			log.Errorf("%v", err)
			tcp.buffer = make([]byte, 0)
			return
		}
		if cmd <= 0 {
			return
		}
		if len(tcp.buffer) >= pos {
			tcp.buffer = append(tcp.buffer[:0], tcp.buffer[pos:]...)
		} else {
			tcp.buffer = make([]byte, 0)
			log.Errorf("pos %v error, len is %v, data is: %+v", pos, len(tcp.buffer), tcp.buffer)
			return
		}
		if !hasCmd(cmd) {
			log.Errorf("cmd %d dos not exists", cmd)
			tcp.buffer = make([]byte, 0)
			return
		}
		for _, f := range tcp.onEvents {
			f(tcp, cmd, content)
		}
	}
}

func (tcp *Client) disconnect() {
	if tcp.status & agentStatusConnect <= 0 {
		log.Debugf("agent is in disconnect status")
		return
	}
	log.Warnf("====================agent disconnect====================")
	tcp.conn.Close()
	if tcp.status & agentStatusConnect > 0 {
		tcp.status ^= agentStatusConnect
	}
}

