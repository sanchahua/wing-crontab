package service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.xunlei.cn/xlsoa/common/jwt"
	xlsoa_config "gitlab.xunlei.cn/xlsoa/config"
	"gitlab.xunlei.cn/xlsoa/service/log"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Environment struct {
	jwt             *jwt.Config
	transportAddr   string // Typically, envoy egress
	configFilePaths []string
	credsFilePaths  []string

	hosts map[string]configHost

	// config
	configLoader xlsoa_config.Loader

	// Promethus
	promHttpAddr    string
	promMetricsPath string
}

func NewEnvironment() *Environment {

	env := &Environment{
		configFilePaths: []string{"./soa.yml", "./soa.yaml", "./conf/soa.yml", "./conf/soa.yaml", "./xlsoa/conf/internal.yaml"},
		credsFilePaths:  []string{"./creds.json", "./conf/creds.json"},
		hosts:           make(map[string]configHost),
	}

	env.initDefaultConfig()
	env.initArguments()

	// Log debug
	log.Printf("[xlsoa] [environment] With transportAddr '%v'\n", env.transportAddr)
	log.Printf("[xlsoa] [environment] With prometheusHttp '%v'\n", env.promHttpAddr)
	log.Printf("[xlsoa] [environment] With prometheusMetricsPath '%v'\n", env.promMetricsPath)

	env.initJwtConfig()
	env.initPromHttp()

	return env

}

// Load default xlsoa config loader
func (env *Environment) initDefaultConfig() {
	paths := env.configFilePaths
	if f := checkEnv("XLSOA_KUBESERVICE_CONFIG_FILE"); f != "" {
		//It's a special handling of xlsoa service deployed in
		//k8s cluster. The xlsoa config file is not convenient
		//to be placed in a relative path along with the program.
		paths = []string{f}
	}

	for _, p := range paths {
		log.Printf("[xlsoa] [environment] Try loading config '%v'\n", p)

		data, err := ioutil.ReadFile(p)
		if err != nil {
			log.Printf("[xlsoa] [environment] Read file error:'%v'\n", err)
			continue
		}

		loader := xlsoa_config.NewYamlLoader(data)
		if err := loader.Init(); err != nil {
			log.Printf("[xlsoa] [environment] New config loader error:'%v'\n", err)
			continue
		}

		log.Printf("[xlsoa] [environment] Load default config success from '%v'\n", p)
		env.configLoader = loader
		break
	}

	if env.configLoader == nil {
		log.Printf("[xlsoa] [environment] [Warn] No config loaded from  '%v'\n", paths)
	}
}

// Init arguments from different sources.
// User specified will be the most prior.
func (env *Environment) initArguments() {
	// From env
	env.transportAddr = checkEnv("MODULES_XLSOA_ENVIRONMENT_TRANSPORT_ADDR")
	env.promHttpAddr = checkEnv("MODULES_XLSOA_ENVIRONMENT_PROMETHEUS_LISTEN_ADDR")
	env.promMetricsPath = checkEnv("MODULES_XLSOA_ENVIRONMENT_PROMETHEUS_PATH")

	// Try from configure file
	var cfg *configEnvironment
	if cfg = env.getEnvironmentConfig(); cfg == nil {
		return
	}

	if cfg.Transport.Addr != "" {
		env.transportAddr = cfg.Transport.Addr
	}
	if cfg.Prometheus.Listen.Addr != "" {
		env.promHttpAddr = cfg.Prometheus.Listen.Addr
	}
	if cfg.Prometheus.Path != "" {
		env.promMetricsPath = cfg.Prometheus.Path
	}

	for _, h := range cfg.Hosts {
		env.hosts[h.Service] = h
	}
}

func (env *Environment) initJwtConfig() {
	paths := env.credsFilePaths

	for _, p := range paths {
		c := jwt.NewConfig()
		err := c.LoadFromFile(p)
		if err != nil {
			log.Printf("[xlsoa] [environment] Try '%v' fail\n", p)
			continue
		}

		log.Printf("[xlsoa] [environment] Load '%v' success\n", p)
		env.jwt = c
		break
	}

	if env.jwt == nil {
		log.Printf("[xlsoa] [environment] [Warn] No 'creds.json' is loaded from '%v'\n", paths)
	}
}

func (env *Environment) initPromHttp() {

	if env.promHttpAddr == "" {
		return
	}
	log.Printf("[xlsoa] [environment] Promethus http address '%v'\n", env.promHttpAddr)

	path := "/metrics"
	if env.promMetricsPath != "" {
		path = env.promMetricsPath
	}

	mu := http.NewServeMux()
	mu.Handle(path, promhttp.Handler())
	go http.ListenAndServe(env.promHttpAddr, mu)
}

// Get and populate configEnvironment
func (env *Environment) getEnvironmentConfig() *configEnvironment {
	var err error
	if env.configLoader == nil {
		log.Println("[xlsoa] [environment] No default config loaded")
		return nil
	}

	cfg := &configEnvironment{}
	var v *xlsoa_config.Value
	if v, err = env.configLoader.Get("modules.xlsoa.environment"); err != nil || v == nil {
		log.Printf("[xlsoa] [environment] [Error] Get config 'modules.xlsoa.environment' fail")
		return nil
	} else if err = v.Populate(cfg); err != nil {
		log.Printf("[xlsoa] [environment] [Error] Populate 'modules.xlsoa.environment' fail")
		return nil
	}
	log.Printf("[xlsoa] [environment] Default config loaded '%v'\n", cfg)
	return cfg
}

func (env *Environment) GetConfigLoader() xlsoa_config.Loader {
	return env.configLoader
}

func (env *Environment) GetServiceAddr(name string) string {
	h, ok := env.hosts[name]
	if ok && h.Addr != "" {
		return h.Addr
	}

	return env.transportAddr
}

func (env *Environment) CheckServiceOauthSecure(name string) bool {
	h, ok := env.hosts[name]
	if ok {
		return h.Oauth
	}

	return true
}

func (env *Environment) GetJwtSign(scopes []string) string {
	if env.jwt == nil {
		return ""
	}
	return env.jwt.Sign(scopes)
}

func (env *Environment) GrpcDialer() func(string, time.Duration) (net.Conn, error) {
	return func(target string, duration time.Duration) (net.Conn, error) {
		addr := env.GetServiceAddr(target)
		if addr == "" {
			return nil, errors.New(fmt.Sprintf("No service address found for '%v'", target))
		}
		log.Printf("[xlsoa] [environment] Dial %v:%v\n", target, addr)
		return net.Dial("tcp", addr)
	}
}
