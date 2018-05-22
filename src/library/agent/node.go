package agent

import (
	"time"
	"net"
	log "github.com/sirupsen/logrus"
	"sync"
	"io"
	"context"
)

type TcpClientNode struct {
	conn *net.Conn   // 客户端连接进来的资源句柄
	sendQueue chan []byte // 发送channel
	sendFailureTimes int64       // 发送失败次数
	recvBuf []byte      // 读缓冲区
	connectTime int64       // 连接成功的时间戳
	recvBufLock *sync.Mutex
	status int
	wg *sync.WaitGroup
	lock *sync.Mutex          // 互斥锁，修改资源时锁定
	onclose []NodeFunc
	ctx context.Context
	onServerEvents []OnServerEventFunc
}

type OnPullCommandFunc func(node *TcpClientNode)

func setOnServerEvents(f ...OnServerEventFunc) NodeOption {
	return func(n *TcpClientNode) {
		n.onServerEvents = append(n.onServerEvents, f...)
	}
}

func newNode(ctx context.Context, conn *net.Conn, opts ...NodeOption) *TcpClientNode {
	node := &TcpClientNode{
		conn:             conn,
		sendQueue:        make(chan []byte, tcpMaxSendQueue),
		sendFailureTimes: 0,
		connectTime:      time.Now().Unix(),
		recvBuf:          make([]byte, 0),
		status:           tcpNodeOnline,
		ctx:              ctx,
		lock:             new(sync.Mutex),
		onclose:          make([]NodeFunc, 0),
		wg:               new(sync.WaitGroup),
		onServerEvents:   make([]OnServerEventFunc, 0),
		recvBufLock:      new(sync.Mutex),
	}
	for _, f := range opts {
		f(node)
	}
	go node.asyncSendService()
	return node
}

func NodeClose(f NodeFunc) NodeOption {
	return func(n *TcpClientNode) {
		n.onclose = append(n.onclose, f)
	}
}

func (node *TcpClientNode) close() {
	node.lock.Lock()
	if node.status & tcpNodeOnline <= 0 {
		node.lock.Unlock()
		return
	}
	if node.status & tcpNodeOnline > 0{
		node.status ^= tcpNodeOnline
		(*node.conn).Close()
		close(node.sendQueue)
	}
	log.Warnf("node close")
	node.lock.Unlock()
	for _, f := range node.onclose {
		f(node)
	}
}

func (node *TcpClientNode) send(data []byte) (int, error) {
	(*node.conn).SetWriteDeadline(time.Now().Add(time.Second * 3))
	return (*node.conn).Write(data)
}

func (node *TcpClientNode) AsyncSend(data []byte) {
	node.lock.Lock()
	if node.status & tcpNodeOnline <= 0 {
		node.lock.Unlock()
		return
	}

	//start1 := time.Now()
	for {
		if len(node.sendQueue) < cap(node.sendQueue) {
			break
		}
		log.Warnf("cache full, try wait, %v, %v", len(node.sendQueue) , cap(node.sendQueue))
	}
	node.sendQueue <- data
	node.lock.Unlock()
	//log.Debugf("############AsyncSend use time========= %+v", time.Since(start1))
}

//func (node *TcpClientNode) SendKeep(data []byte) {
//	node.keepLock.Lock()
//	node.keep[]
//	node.keepLock.Unlock()
//}

func (node *TcpClientNode) setReadDeadline(t time.Time) {
	(*node.conn).SetReadDeadline(t)
}

func (node *TcpClientNode) asyncSendService() {
	node.wg.Add(1)
	defer node.wg.Done()
	for {
		if node.status & tcpNodeOnline <= 0 {
			log.Info("tcp node is closed, clientSendService exit.")
			return
		}
		select {
		case msg, ok := <-node.sendQueue:
			//start := time.Now()
			if !ok {
				log.Info("tcp node sendQueue is closed, sendQueue channel closed.")
				return
			}
			//if len(msg) == 1 && msg[0] == byte(0) {
			//	close(node.sendQueue)
			//	log.Warnf("close sendQueue")
			//	return
			//}
			(*node.conn).SetWriteDeadline(time.Now().Add(time.Second * 30))
			size, err := (*node.conn).Write(msg)
			//log.Debugf("send: %+v, to %+v", msg, (*node.conn).RemoteAddr().String())
			if err != nil {
				log.Errorf("tcp send to %s error: %v", (*node.conn).RemoteAddr().String(), err)
				node.close()
				return
			}
			if size != len(msg) {
				log.Errorf("%s send not complete: %v", (*node.conn).RemoteAddr().String(), msg)
			}
			//fmt.Fprintf(os.Stderr, "write use time %v\r\n", time.Since(start))
		case <-node.ctx.Done():
			log.Debugf("context is closed, wait for exit, left: %d", len(node.sendQueue))
			if len(node.sendQueue) <= 0 {
				log.Info("tcp service, clientSendService exit.")
				return
			}
		}
	}
}

func (node *TcpClientNode) onMessage(msg []byte) {
	//node.recvBufLock.Lock()
	node.recvBuf = append(node.recvBuf, msg...)
	//node.recvBufLock.Unlock()
	//log.Debugf("data: %+v", node.recvBuf)

	for {
		if node.recvBuf == nil || len(node.recvBuf) < 6 {
			//log.Errorf("node.recvBuf len error %v", len(node.recvBuf))
			return
		}

		if len(node.recvBuf) > MAX_PACKAGE_LEN {
			log.Errorf("max len error")
			//node.recvBufLock.Lock()

			node.recvBuf = make([]byte, 0)
			//node.recvBufLock.Unlock()

			return
		}
///////////////////////////////////////////////////


		//clen := int(binary.LittleEndian.Uint32(node.recvBuf[:4]))
		//log.Debugf("clen=%+v", clen)
		//if len(node.recvBuf) < clen + 4 {
		//	log.Warnf("content len error")
		//	return //0, nil, 0, DataLenError
		//}
		//log.Debugf("cmd=%+v", node.recvBuf[4:6])
		//cmd     := int(binary.LittleEndian.Uint16(node.recvBuf[4:6]))
		//log.Debugf("content=%+v === %v", node.recvBuf[6 : clen + 4], string(node.recvBuf[6 : clen + 4]))
		//content := node.recvBuf[6 : clen + 4]
		////data  = append(data[:0], data[clen+4:]...)
		//log.Debugf("return(%+v)(%+v)(%+v)", cmd, content, nil)
		//node.recvBufLock.Lock()

		cmd, content, err := Unpack(&node.recvBuf)
		//node.recvBufLock.Unlock()
		if err != nil {
			log.Errorf("node.recvBuf error %v", err)
			return
		}
		//end:= clen+4//, nil
		//if content == nil {
		//	return
		//}
		/////////////////////////////////////////
		if !hasCmd(cmd) {
			//node.recvBufLock.Lock()

			node.recvBuf = make([]byte, 0)
			//node.recvBufLock.Unlock()

			log.Errorf("cmd（%v）does not exists", cmd)
			return
		}
//log.Debugf("agent node:%+v",content)
		//start := time.Now()
		for _, f := range node.onServerEvents {
			f(node, cmd, content)
		}
		//fmt.Fprintf(os.Stderr, "node.onServerEvents use time %v\r\n", time.Since(start))
		//switch cmd {
		//case CMD_TICK:
		//	node.AsyncSend(packDataTickOk)
		//case CMD_CRONTAB_CHANGE:
		//	//var data SendData
		//	//err := json.Unmarshal(content, &data)
		//	//if err != nil {
		//	//	log.Errorf("%+v", err)
		//	//} else {
		//	//	event := binary.LittleEndian.Uint32(data.Data[:4])
		//	//	go node.eventFired(int(event), data.Data[4:])
		//	//	//log.Infof("receive event[%v] %+v", event, string(data.Data[4:]))
		//	//	node.AsyncSend(Pack(CMD_CRONTAB_CHANGE, []byte(data.Unique)))
		//	//}
		//case CMD_PULL_COMMAND:
		//	//start := time.Now()
		//	//node.o nPullCommand(node)
		//	//log.Debugf("###############PullCommand use time %+v", time.Since(start))
		//case CMD_RUN_COMMAND:
		//default:
		//	node.AsyncSend(Pack(CMD_ERROR, []byte(fmt.Sprintf("tcp service does not support cmd: %d", cmd))))
		//	node.recvBuf = make([]byte, 0)
		//	return
		//}
		//node.recvBuf = append( node.recvBuf[:0],  node.recvBuf[end:]...)
	}
}

//func (node *TcpClientNode) eventFired(event int, data []byte) {
//	for _, f := range node.onevents {
//		f(event, data)
//	}
//}

func (node *TcpClientNode) readMessage() {
	//node := newNode(tcp.ctx, conn, NodeClose(tcp.agents.remove), NodePro(tcp.agents.append))
	var readBuffer [tcpDefaultReadBufferSize]byte
	// 设定3秒超时，如果添加到分组成功，超时限制将被清除
	for {
		size, err := (*node.conn).Read(readBuffer[0:])
		if err != nil {
			if err != io.EOF {
				log.Warnf("tcp node %s disconnect with error: %v", (*node.conn).RemoteAddr().String(), err)
			} else {
				log.Debugf("tcp node %s disconnect with error: %v", (*node.conn).RemoteAddr().String(), err)
			}
			node.close()
			return
		}
		//log.Debugf("#####################server receive message: %+v", readBuffer[:size])
		node.onMessage(readBuffer[:size])
	}
}


