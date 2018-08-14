package jwt

import (
	"io/ioutil"
	"encoding/json"
	"strings"
	xlsoa_jws "gitlab.xunlei.cn/xlsoa/common/jws"
	"golang.org/x/oauth2/jws"
)

var (
	defaultHeader    = &jws.Header{Algorithm: "RS256", Typ: "JWT"}
)

//service的描述的配置文件结构,该描述文件由service_manager server产生
type Config struct {
	//标识该服务的ID
	ServiceId           string      `json:"service_id"`

	//标识该服务的名字，用于服务注册和发现
	ServiceName         string      `json:"service_name"`

	//访问其它服务时所使用的身份标识ID
	ClientId            string      `json:"client_id"`

	//访问其它服务时进行jwt签名时所使用的的秘钥
	ClientSecret        string      `json:"client_secret"`

	//认证服务器的服务名
	AuthServiceName     string      `json:"auth_service_name"`

	//key同步服务器的服务名
	KeySyncServiceName  string      `json:"key_sync_service_name"`

	//Gateway服务器的服务名
	GatewayServiceName  string       `json:"gateway_service_name"`
}

func NewConfig() *Config{
	return &Config{}
}

func (config *Config)LoadFromFile(file string) error{
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}
	return nil
}

//对Jwt描述文件使用config.ClientSecret作为秘钥进行hmac256签名
//其中ClientId填入头部的header.KeyID，其他信息不需要填写
func (config *Config)Sign(scopes []string) string{
	header := defaultHeader

	//在申请授权时填写的header里，keyID填写ClientId字段
	//在授权成功返回的header里，keyID填写AuthKeyId字段, 该字段用于查找对应的Key来校验token。(这些key在key同步流程中定时由认证服务器同步到本地)
	header.KeyID = config.ClientId

	claimSet := &jws.ClaimSet{
		Scope:  strings.Join(scopes, " "),
	}
	result, _ := xlsoa_jws.Encode(header, claimSet, config.ClientSecret)
	return result
}
