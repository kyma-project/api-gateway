package helpers

import (
	"os"

	"gopkg.in/yaml.v2"
)

const JWT_HANDLER_ORY = "ory"
const JWT_HANDLER_ISTIO = "istio"

const CONFIG_FILE = "/api-gateway-config/api-gateway-config.yaml"

var ReadFileHandle = os.ReadFile

type Config struct {
	JWTHandler string `yaml:"jwtHandler"`
}

func DefaultConfig() *Config {
	return &Config{JWTHandler: JWT_HANDLER_ORY}
}

func LoadConfig() (*Config, error) {
	configData, err := ReadFileHandle(CONFIG_FILE)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}

	return config, err
}
