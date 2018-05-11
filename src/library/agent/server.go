package agent

import (
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
	"context"
)

type TcpService struct {
	Address string               // 监听ip
	lock *sync.Mutex
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
		lock:             new(sync.Mutex),
		statusLock:       new(sync.Mutex),
		wg:               new(sync.WaitGroup),
		listener:         nil,
		agents:           nil,
		status:           0,
		buffer:           make([]byte, 0),
		onevents:         make([]OnNodeEventFunc, 0),
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
					NodeClose(tcp.agents.remove),
					SetOnNodeEvent(tcp.onevents...),
				)
			log.Infof("new connect %v", conn.RemoteAddr().String())
			log.Infof("#####################nodes len before %v", len(tcp.agents))
			tcp.agents.append(node)
			log.Infof("#####################nodes len after %v", len(tcp.agents))
			//tcp.Clients() // debug
			go node.readMessage()
		}
	}()
}

func (tcp *TcpService) Clients() TcpClients {
	log.Debugf("get clients %v", len(tcp.agents))
	return tcp.agents
}

func (tcp *TcpService) Close() {
	log.Debugf("tcp service closing, waiting for buffer send complete.")
	tcp.lock.Lock()
	defer tcp.lock.Unlock()
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
