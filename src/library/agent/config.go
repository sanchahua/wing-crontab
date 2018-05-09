package agent

import (
	"library/file"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"app"
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


type NodeFunc func(n *tcpClientNode)
type NodeOption func(n *tcpClientNode)
type tcpClients []*tcpClientNode


type OnPosFunc func(r []byte)
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

