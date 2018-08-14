package service

import (
	"fmt"
)

type configEnvironment struct {
	Transport struct {
		Addr string `yaml:"addr"`
	} `yaml:"transport"`
	Prometheus struct {
		Listen struct {
			Addr string `yaml:"addr"`
		} `yaml:"listen"`
		Path string `yaml:"path"`
	} `yaml:"prometheus"`
	Hosts []configHost `yaml:hosts`
}

type configHost struct {
	Service string `yaml:service`
	Addr    string `yaml:addr`
	Oauth   bool   `yaml:oauth`
}

func (c *configHost) String() string {
	return fmt.Sprintf("ConfigHost{ service: '%v', addr: '%v', oauth: %v}", c.Service, c.Addr, c.Oauth)
}

func (c *configEnvironment) String() string {
	return fmt.Sprintf("ConfigEnvironment{ Transport: { Addr: '%v'}, Prometheus: { Listen: { Addr: '%v'}, Path: '%v'}, Hosts: %v }",
		c.Transport.Addr, c.Prometheus.Listen.Addr, c.Prometheus.Path, c.Hosts)
}

type configServerContext struct {
	Addr  string `yaml:"addr"`
	Oauth struct {
		Secure struct {
			Switch string `yaml:"switch"`
			Level  string `yaml:"level"`
		} `yaml:"secure"`
	}
}

func (c *configServerContext) String() string {
	return fmt.Sprintf("ConfigServerContext{ Addr: '%v', Oauth: { Secure: { Switch: '%v', Level: '%v'} } }", c.Addr, c.Oauth.Secure.Switch, c.Oauth.Secure.Level)
}
