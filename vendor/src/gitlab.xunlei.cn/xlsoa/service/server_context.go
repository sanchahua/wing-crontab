package service

import (
	//grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	_ "github.com/pkg/errors"
	xlsoa_config "gitlab.xunlei.cn/xlsoa/config"
	"gitlab.xunlei.cn/xlsoa/service/log"
	"google.golang.org/grpc"
)

type ServerContext struct {
	env              *Environment
	addr             string
	withOauthSecure  bool
	oAuthSecureLevel OauthSecureLevel

	oauth *oAuthGuardian
}

func NewServerContext(env *Environment) *ServerContext {
	c := &ServerContext{
		env:              env,
		withOauthSecure:  true,
		oAuthSecureLevel: OauthSecureLevelDegradeWhenException,
	}

	c.initArguments()
	log.Printf("[xlsoa] [server context] With addr '%v'\n", c.addr)
	log.Printf("[xlsoa] [server context] With oauthSecure '%v'\n", c.withOauthSecure)
	log.Printf("[xlsoa] [server context] With oauthSecureLevel '%v'\n", c.oAuthSecureLevel)
	c.initOauthSecure()

	return c
}

func (c *ServerContext) initOauthSecure() {
	var err error

	if !c.withOauthSecure {
		return
	}

	if c.oauth, err = NewOauthGuardian(c); err != nil {
		log.Printf("[xlsoa] [server context] [Warn] NewOauthGuardian error: %v. No oauth secure will be startup.\n", err)
		return
	}
}

func (c *ServerContext) Close() {
	if c.oauth != nil {
		c.oauth.Close()
		c.oauth = nil
	}
}

func (c *ServerContext) initArguments() {
	// From env
	c.addr = checkEnv("MODULES_XLSOA_SERVER_CONTEXT_ADDR")
	sw := checkEnv("MODULES_XLSOA_SERVER_CONTEXT_OAUTH_SECURE_SWITCH")
	if sw != "" {
		if sw == "on" {
			c.withOauthSecure = true
		} else if sw == "off" {
			c.withOauthSecure = false
		}
	}

	// Try from configure file
	cfg := c.getServerContextConfig()
	if cfg == nil {
		return
	}

	if cfg.Addr != "" {
		c.addr = cfg.Addr
	}

	if cfg.Oauth.Secure.Switch != "" {
		if cfg.Oauth.Secure.Switch == "on" {
			c.withOauthSecure = true
		} else if cfg.Oauth.Secure.Switch == "off" {
			c.withOauthSecure = false
		} else {
			log.Printf("[xlsoa] [server context] [Error] Invalid config value 'oauth.secure.switch':'%v', abandon it.\n", cfg.Oauth.Secure.Switch)
		}
	}

	if cfg.Oauth.Secure.Level != "" {
		if level, ok := oAuthSecureLevelStringToLevel[cfg.Oauth.Secure.Level]; ok {
			c.oAuthSecureLevel = level
		} else {
			log.Printf("[xlsoa] [server context] [Error] Invalid config value 'oauth.secure.level':'%v', abandon it.\n", cfg.Oauth.Secure.Level)
		}
	}
}

func (c *ServerContext) getServerContextConfig() *configServerContext {
	var err error

	loader := c.env.GetConfigLoader()
	if loader == nil {
		log.Println("[xlsoa] [server context] No default config loaded")
		return nil
	}

	cfg := &configServerContext{}
	var v *xlsoa_config.Value
	if v, err = loader.Get("modules.xlsoa.server.context"); err != nil || v == nil {
		log.Printf("[xlsoa] [server context] No 'modules.xlsoa.server.context' configured.")
		return nil
	} else if err = v.Populate(cfg); err != nil {
		log.Printf("[xlsoa] [server context] [Error] Populate 'modules.xlsoa.server.context' fail")
		return nil
	}
	log.Printf("[xlsoa] [server context] Default config loaded '%v'\n", cfg)
	return cfg
}

func (c *ServerContext) GrpcUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	if c.oauth != nil {
		return c.oauth.UnaryServerInterceptor()
	}
	return nil
}

func (c *ServerContext) GrpcStreamServerInterceptor() grpc.StreamServerInterceptor {
	if c.oauth != nil {
		return c.oauth.StreamServerInterceptor()
	}
	return nil
}

func (c *ServerContext) GetAddr() string {
	if c.addr == "" {
		panic("ServerContext: addr empty!")
	}
	return c.addr
}

func (c *ServerContext) GetEnv() *Environment {
	return c.env
}

func (c *ServerContext) GetOauthSecureLevel() OauthSecureLevel {
	return c.oAuthSecureLevel
}
