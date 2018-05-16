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
	log.Debugf("get clients %v", len(tcp.agents))
	//return tcp.agents
	start := time.Now()
	//start1 := time.Now()
	//clients := c.server.Clients()
	//log.Debugf("c.server.Clients use time: %+v", time.Since(start1))

	//start2 := time.Now()
	l := int64(len(tcp.agents))
	//log.Debugf("c.server.Clients use time => 2 : %+v", time.Since(start2))

	if l <= 0 {
		log.Debugf("clients empty")
		return
	}
	//start22 := time.Now()

	if tcp.index >= l {
		atomic.StoreInt64(&tcp.index, 0)
	}
	//log.Infof("clients %+v", l)
	//log.Debugf("c.server.Clients use time => 22 : %+v", time.Since(start22))

	tcp.lock.Lock()
	client := tcp.agents[tcp.index]
	tcp.lock.Unlock()
	//for key, client := range tcp.agents {
	//	if key != int(tcp.index) {
	//		continue
	//	}
		//start3 := time.Now()
		//log.Infof("dispatch %v=>%v to client[%v]", item.id, item.command, c.index)
		//client := clients[c.index]
		//atomic.AddInt64(&tcp.index, 1)
		//data := make([]byte, 8)
		//binary.LittleEndian.PutUint64(data, uint64(item.id))
		//
		//dataCommendLen := make([]byte, 8)
		//binary.LittleEndian.PutUint64(dataCommendLen, uint64(len(item.command)))
		//
		//data = append(data, dataCommendLen...)
		//data = append(data, []byte(item.command)...)
		//
		////data = append(data, dataDispatchServerLen...)
		//data = append(data, []byte(c.ctx.Config.BindAddress)...)
		//log.Debugf("c.server.Clients use time => 3 : %+v", time.Since(start3))
		sendData := Pack(CMD_RUN_COMMAND, data)

		start5   := time.Now()
		client.AsyncSend(sendData)
		log.Debugf("AsyncSend send use time: %+v", time.Since(start5))

	//}
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
