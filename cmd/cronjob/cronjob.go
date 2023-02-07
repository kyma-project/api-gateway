package main

import (
	"context"
	"errors"
	"os"

	"github.com/kyma-project/api-gateway/internal/webhook"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	logger := zap.New()
	logger.WithName("cert-generation-cronjob")
	crdName, ok := os.LookupEnv("CRD_NAME")
	if !ok {
		logger.Error(errors.New("CRD_NAME environment variable wasn't set"), "setup")
		return
	}

	secretName, ok := os.LookupEnv("SECRET_NAME")
	if !ok {
		logger.Error(errors.New("SECRET_NAME environment variable wasn't set"), "setup")
		return
	}

	err := webhook.SetupCertificates(context.Background(), logger, crdName, secretName)
	if err != nil {
		panic(err)
	}
}
