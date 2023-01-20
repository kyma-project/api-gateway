package jwt

import (
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config JWT configuration structure
type Config struct {
	EnvConfig    envConfig
	ClientConfig clientConfig
}

type clientConfig struct {
	ClientTimeout time.Duration
}

type envConfig struct {
	ClientTimeout time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` //Don't forget the unit!
}

func NewJwtConfig() (Config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return Config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := Config{EnvConfig: env}
	config.ClientConfig = clientConfig{ClientTimeout: env.ClientTimeout}

	return config, nil
}
