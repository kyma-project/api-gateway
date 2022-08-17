package main

import (
	"context"
	"errors"
	"os"

	"github.com/kyma-incubator/api-gateway/internal/webhook"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	logger := zap.New()
	logger.WithName("api-gateway-cert-cronjob")
	secretName, ok := os.LookupEnv("SECRET_NAME")
	if !ok {
		logger.Error(errors.New("SECRET_NAME environment variable wasn't set"), "setup")
		return
	}

	secretNamespace, ok := os.LookupEnv("SECRET_NAMESPACE")
	if !ok {
		logger.Error(errors.New("SECRET_NAMESPACE environment variable wasn't set"), "setup")
		return
	}

	serviceName, ok := os.LookupEnv("SERVICE_NAME")
	if !ok {
		logger.Error(errors.New("SERVICE_NAME environment variable wasn't set"), "setup")
		return
	}

	err := webhook.SetupCertificates(context.Background(), logger, secretName, secretNamespace, serviceName)
	if err != nil {
		panic(err)
	}
}
