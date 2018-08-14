package config

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
)

type Config struct {
	name string
	addr string

	dataCenterName string
	nodeName       string
	instanceName   string
}

type OptionFunc func(c *Config)

func New(name string, opts ...OptionFunc) *Config {
	c := &Config{
		name: name,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Config) Load() (Loader, error) {

	var err error

	// Init environment
	if c.addr == "" {
		c.addr = os.Getenv(ENV_CONFIG_CENTER_ADDR)
		log.Printf("Config: Checking %v from ENV: '%v'\n", ENV_CONFIG_CENTER_ADDR, c.addr)
	}

	if c.dataCenterName == "" {
		c.dataCenterName = os.Getenv(ENV_DATACENTER_NAME)
		log.Printf("Config: Checking %v from ENV: '%v'\n", ENV_DATACENTER_NAME, c.dataCenterName)
	}
	if c.nodeName == "" {
		c.nodeName = os.Getenv(ENV_NODE_NAME)
		log.Printf("Config: Checking %v from ENV: '%v'\n", ENV_NODE_NAME, c.nodeName)
	}

	if c.addr == "" {
		return nil, errors.New("No loader installed")
	}

	// Create cache loader with config center loader.
	prefix := fmt.Sprintf("%v/%v/", CONFIG_KEY_PREFIX, c.name)
	opts := []configCenterLoaderOptionFunc{}
	if c.dataCenterName != "" {
		opts = append(opts, ConfigCenterLoaderWithProperty("dc", c.dataCenterName))
	}
	if c.nodeName != "" {
		opts = append(opts, ConfigCenterLoaderWithProperty("node", c.nodeName))
	}
	if c.instanceName != "" {
		opts = append(opts, ConfigCenterLoaderWithProperty("instance", c.instanceName))
	}
	log.Printf("Datacenter: %v, Node: %v, Instance: %v\n", c.dataCenterName, c.nodeName, c.instanceName)
	ccLoader := NewConfigCenterLoader(c.addr, prefix, opts...)
	if err = ccLoader.Init(); err != nil {
		return nil, errors.Wrap(err, "Init ConfigCenterLoader error")
	}

	loader := NewCacheLoader(
		CONFIG_CACHE_DIR,
		c.name+".yaml",
		ccLoader,
	)
	if err = loader.Init(); err != nil {
		return nil, errors.Wrap(err, "CacheLoader Init fail")
	}

	return loader, nil
}

func WithAddr(addr string) OptionFunc {
	return func(c *Config) {
		c.addr = addr
	}
}

func WithDcName(name string) OptionFunc {
	return func(c *Config) {
		c.dataCenterName = name
	}
}
func WithNodeName(name string) OptionFunc {
	return func(c *Config) {
		c.nodeName = name
	}
}
func WithInstanceName(name string) OptionFunc {
	return func(c *Config) {
		c.instanceName = name
	}
}
