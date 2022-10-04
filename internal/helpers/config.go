package helpers

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	JWTHandler string `yaml:"jwtHandler"`
}

const CONFIG_FILE = "/api-gateway-config.yaml"

func LoadConfig() (*Config, error) {
	config := &Config{}
	configData, err := ioutil.ReadFile(CONFIG_FILE)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}

	return config, err
}
