package agent

import (
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
	"context"
	"sync/atomic"
)

type TcpService struct {
	Address string               // 监听ip
	lock *sync.Mutex
	debuglock *sync.Mutex

	statusLock *sync.Mutex
	//ctx *app.Context
	listener *net.Listener
	wg *sync.WaitGroup
	agents TcpClients
	status int
	conn *net.TCPConn
	buffer []byte
	ctx context.Context
	onevents []OnNodeEventFunc
	index int64
}
type AgentServerOption func(s *TcpService)

func SetEventCallback(f ...OnNodeEventFunc) AgentServerOption {
	return func(s *TcpService) {
		s.onevents = append(s.onevents, f...)
	}
}
func NewAgentServer(ctx context.Context, address string, opts ...AgentServerOption) *TcpService {
	tcp := &TcpService{
		ctx:              ctx,
		Address:          address,
		// use for agents
		lock:             new(sync.Mutex),
		debuglock:        new(sync.Mutex),
		statusLock:       new(sync.Mutex),
		wg:               new(sync.WaitGroup),
		listener:         nil,
		agents:           nil,
		status:           0,
		buffer:           make([]byte, 0),
		onevents:         make([]OnNodeEventFunc, 0),
		index:            0,
	}
	go tcp.keepalive()
	for _, f := range opts {
		f(tcp)
	}
	return tcp
}

func (tcp *TcpService) Start() {
	go func() {
		listen, err := net.Listen("tcp", tcp.Address)
		if err != nil {
			log.Errorf("tcp service listen with error: %+v", err)
			return
		}
		tcp.listener = &listen
		log.Infof("agent service start with: %s", tcp.Address)
		for {
			conn, err := listen.Accept()
			select {
			case <-tcp.ctx.Done():
				return
			default:
			}
			if err != nil {
				log.Warnf("tcp service accept with error: %+v", err)
				continue
			}
			node := newNode(
					tcp.ctx,
					&conn,
					NodeClose(func(n *tcpClientNode) {
						tcp.lock.Lock()
						tcp.agents.remove(n)
						tcp.lock.Unlock()
					}),
					SetOnNodeEvent(tcp.onevents...),
				)
			log.Infof("new connect %v", conn.RemoteAddr().String())
			log.Infof("#####################nodes len before %v", len(tcp.agents))
			tcp.lock.Lock()
			tcp.agents.append(node)
			tcp.lock.Unlock()
			log.Infof("#####################nodes len after %v", len(tcp.agents))
			//tcp.Clients() // debug
			go node.readMessage()
		}
	}()
}

func (tcp *TcpService) RandSend(data []byte) {
	//tcp.debuglock.Lock()
	//defer tcp.debuglock.Unlock()
	start := time.Now()
	l := int64(len(tcp.agents))
	if l <= 0 {
		return
	}
	//log.Debugf("RandSend 1 use time : %+v", time.Since(start))
	//start2 := time.Now()

	if tcp.index >= l {
		atomic.StoreInt64(&tcp.index, 0)
	}
	//log.Debugf("RandSend 2 use time : %+v", time.Since(start2))

	//start3 := time.Now()
	iindex := int(tcp.index)
	for key, client := range tcp.agents {
		if key != iindex {
			continue
		}
		//startp := time.Now()
		sendData := Pack(CMD_RUN_COMMAND, data)
		//log.Debugf("pack use time: %+v", time.Since(startp))

		//start5   := time.Now()
		client.AsyncSend(sendData)
		//log.Debugf("AsyncSend send use time: %+v", time.Since(start5))

		//start4 := time.Now()
		atomic.AddInt64(&tcp.index, 1)
		//log.Debugf("RandSend 4 use time : %+v", time.Since(start4))
		break

	}
	//log.Debugf("RandSend 3 use time : %+v", time.Since(start3))
	log.Debugf("dispatch use time %+v", time.Since(start))
}

func (tcp *TcpService) Close() {
	log.Debugf("tcp service closing, waiting for buffer send complete.")
	if tcp.listener != nil {
		(*tcp.listener).Close()
	}
	tcp.agents.close()
	log.Debugf("tcp service closed.")
}

//func (tcp *TcpService) SendEvent(table string, data []byte) {
//	// 广播给agent client
//	// agent client 再发送给连接到当前service_plugin/tcp的客户端
//	packData := Pack(CMD_EVENT, data)
//	tcp.agents.asyncSend(packData)
//}

// 心跳
func (tcp *TcpService) keepalive() {
	for {
		select {
		case <-tcp.ctx.Done():
			return
		default:
		}
		if tcp.agents == nil {
			time.Sleep(time.Second * 3)
			continue
		}
		tcp.agents.asyncSend(packDataTickOk)
		time.Sleep(time.Second * 3)
	}
}
