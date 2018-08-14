package consul

import (
	"errors"
	"github.com/hashicorp/consul/api"
)

/*
*操作consul 的client
 */

type Client struct {
	client     *api.Client
	consulAddr string
	agent      *api.Agent  //操作agent的接口，用于注册服务、ttl
	health     *api.Health //操作Health的接口，用于发现健康的服务和watch
}

func NewClient(consulAddr string) *Client {
	return &Client{
		consulAddr: consulAddr,
	}
}

func (cli *Client) Init() error {
	var err error
	cli.client, err = api.NewClient(
		&api.Config{
			Address: cli.consulAddr,
		},
	)
	if err != nil {
		return err
	}
	cli.agent = cli.client.Agent()
	cli.health = cli.client.Health()
	return nil
}

/**
 * 设置service为passing状态
 *
 * @param serviceid
 */
func (cli *Client) CheckPass(serviceid string) {
	cli.agent.PassTTL("service:"+serviceid, "") //consul内部的serviceid：前缀"service:"+注册时提交的serviceid
}

/**
 * 设置service为不可用状态。
 *
 * @param serviceid
 */
func (cli *Client) CheckFail(serviceid string) {
	cli.agent.FailTTL("service:"+serviceid, "")
}

/**
 * 注册一个consul service
 *
 * @param service
 */
func (cli *Client) RegisterService(service *api.AgentServiceRegistration) error {
	return cli.agent.ServiceRegister(service)
}

/**
 * 根据serviceid注销service
 *
 * @param serviceid
 */
func (cli *Client) DeregisterService(serviceid string) {
	cli.agent.ServiceDeregister(serviceid)
}

func (cli *Client) LookupHealthService(serviceName string, queryParams *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	return cli.health.Service(serviceName, "", true, queryParams)
}

func (cli *Client) GetAgentInfo() (map[string]map[string]interface{}, error) {
	return cli.agent.Self()
}

//从consul服务器获取DataCenter的名字，用作zoneId
func (cli *Client) GetZoneIdFromConsul() (string, error) {
	consulAgentInfo, err := cli.GetAgentInfo()
	if err != nil {
		return "", err
	}

	var config map[string]interface{}
	var ok bool
	if config, ok = consulAgentInfo["Config"]; !ok {
		return "", errors.New("can't find zoneid")
	}

	var dataCenter interface{}
	if dataCenter, ok = config["Datacenter"]; !ok {
		return "", errors.New("can't find zoneid")
	}

	var zoneId string
	if zoneId, ok = dataCenter.(string); !ok {
		return "", errors.New("can't find zoneid")
	}

	return zoneId, nil
}

//Get node id from consul
func (cli *Client) GetNodeIdFromConsul() (string, error) {
	consulAgentInfo, err := cli.GetAgentInfo()
	if err != nil {
		return "", err
	}

	var config map[string]interface{}
	var ok bool
	if config, ok = consulAgentInfo["Config"]; !ok {
		return "", errors.New("can't find zoneid")
	}

	var nodeName interface{}
	if nodeName, ok = config["NodeName"]; !ok {
		return "", errors.New("can't find node name")
	}

	var nodeId string
	if nodeId, ok = nodeName.(string); !ok {
		return "", errors.New("can't find node name")
	}

	return nodeId, nil
}
