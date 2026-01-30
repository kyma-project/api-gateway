package helpers

import (
	"context"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	JWT_HANDLER_ORY   = "ory"
	JWT_HANDLER_ISTIO = "istio"

	CM_NS   = "kyma-system"
	CM_NAME = "api-gateway-config"
	CM_KEY  = "api-gateway-config"
)

func ReadConfigMap(ctx context.Context, client client.Client) ([]byte, error) {
	cm := &corev1.ConfigMap{}
	err := client.Get(ctx, types.NamespacedName{Namespace: CM_NS, Name: CM_NAME}, cm)
	if err != nil {
		return nil, err
	}
	return []byte(cm.Data[CM_KEY]), nil
}

var ReadConfigMapHandle = ReadConfigMap

type Config struct {
	JWTHandler string `yaml:"jwtHandler"`
}

func (c *Config) Reset() {
	c.JWTHandler = ""
}

func (c *Config) ResetToDefault() {
	c.JWTHandler = JWT_HANDLER_ORY
}

func (c *Config) ReadFromConfigMap(ctx context.Context, client client.Client) error {
	cmData, err := ReadConfigMapHandle(ctx, client)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(cmData, c)
	if err != nil {
		return err
	}
	return nil
}
