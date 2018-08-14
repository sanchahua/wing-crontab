package xlsoa_config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	DefaultConfigFile string = "./xlsoa_setup_service/xlsoa_config.yaml"
)

type Config struct {
	AuthAddr    string `yaml:"auth_addr"`
	KeySyncAddr string `yaml:"keysync_addr"`
}

func New() *Config {
	return &Config{}
}

func (cfg *Config) Load(cfgFile string) error {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}
