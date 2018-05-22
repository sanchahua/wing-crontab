package agent

import (
	"library/file"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"app"
)

const (
	CMD_ERROR = iota          // 错误响应
	CMD_TICK           // 心跳包
	CMD_AGENT
	CMD_STOP
	CMD_RELOAD
	CMD_SHOW_MEMBERS
	CMD_CRONTAB_CHANGE
	CMD_RUN_COMMAND
	CMD_PULL_COMMAND
	CMD_DEL_CACHE
)

const (
	tcpMaxSendQueue               = 10000
	tcpDefaultReadBufferSize      = 1024
)

const (
	serviceEnable = 1 << iota
	agentStatusOnline
	agentStatusConnect
)
const (
	tcpNodeOnline = 1 << iota
)
const MAX_PACKAGE_LEN = 10240


type NodeFunc func(n *TcpClientNode)
type NodeOption func(n *TcpClientNode)
type TcpClients []*TcpClientNode


type OnPosFunc func(r []byte)
var (
	packDataTickOk     = Pack(CMD_TICK, []byte("keepalive res ok"))
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

