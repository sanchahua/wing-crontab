package agent

import (
	"sync"
	"net"
	"library/file"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"app"
	"context"
)

const (
	CMD_SET_PRO = iota // 注册客户端操作，加入到指定分组
	CMD_AUTH           // 认证（暂未使用）
	CMD_ERROR          // 错误响应
	CMD_TICK           // 心跳包
	CMD_EVENT          // 事件
	CMD_AGENT
	CMD_STOP
	CMD_RELOAD
	CMD_SHOW_MEMBERS
	CMD_POS
	CMD_CRONTAB_CHANGE
)

const (
	tcpMaxSendQueue               = 10000
	tcpDefaultReadBufferSize      = 1024
)

const (
	FlagSetPro = iota
	FlagPing
	FlagControl
	FlagAgent
)

const (
	serviceEnable = 1 << iota
	agentStatusOnline
	agentStatusConnect
)
const (
	tcpNodeOnline = 1 << iota
)

type tcpClientNode struct {
	conn *net.Conn   // 客户端连接进来的资源句柄
	sendQueue chan []byte // 发送channel
	sendFailureTimes int64       // 发送失败次数
	recvBuf []byte      // 读缓冲区
	connectTime int64       // 连接成功的时间戳
	status int
	wg *sync.WaitGroup
	lock *sync.Mutex          // 互斥锁，修改资源时锁定
	onclose []NodeFunc
	agents tcpClients
	onpro []NodeFunc
	ctx context.Context
}

type NodeFunc func(n *tcpClientNode)
type NodeOption func(n *tcpClientNode)

type tcpClients []*tcpClientNode


type OnPosFunc func(r []byte)
type AgentServerOption func(s *TcpService)
var (
	packDataTickOk     = Pack(CMD_TICK, []byte("ok"))
	packDataSetPro     = Pack(CMD_SET_PRO, []byte("ok"))
)

type AgentConfig struct {
	Enable bool `toml:"enable"`
	Type string `toml:"type"`
	Lock string `toml:"lock"`
	AgentListen string `toml:"agent_listen"`
	ConsulAddress string `toml:"consul_address"`
}

func getConfig() (*AgentConfig, error) {
	var config AgentConfig
	configFile := app.ConfigPath + "/agent.toml"
	if !file.Exists(configFile) {
		log.Errorf("config file not found: %s", configFile)
		return nil, app.ErrorFileNotFound
	}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Println(err)
		return nil, app.ErrorFileParse
	}
	return &config, nil
}

