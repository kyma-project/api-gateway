package environment

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync/atomic"
)

type Config struct {
	RunsOnStage bool
	Loaded      *atomic.Bool
}

type Loader struct {
	K8sClient client.Client
	Config    *Config
	Log       logr.Logger
}

func (l *Loader) Start(ctx context.Context) error {
	if l.Config.Loaded == nil {
		l.Config.Loaded = &atomic.Bool{}
	}
	if l.Config.Loaded.Load() {
		return nil
	}

	// Load the configuration from the Kubernetes client
	err := l.loadConfig(ctx)
	if err != nil {
		return err
	}

	l.Config.Loaded.Store(true)
	l.Log.Info("Environment configuration loaded", "RunsOnStage", l.Config.RunsOnStage)
	return nil
}

// loadConfig loads the configuration from the Kubernetes client.
// In particular, it reads the "shoot-info" ConfigMap from "kube-system" and sets the RunsOnStage field according
// to whether the "projectName" data field is "kyma-stage" or not.
func (l *Loader) loadConfig(ctx context.Context) error {
	// Read the ConfigMap from the Kubernetes client
	cm := &corev1.ConfigMap{}
	err := l.K8sClient.Get(ctx, types.NamespacedName{Namespace: "kube-system", Name: "shoot-info"}, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Log.Info("ConfigMap 'shoot-info' not found in 'kube-system' namespace, assuming not running on stage")
			l.Config.RunsOnStage = false
			return nil
		}
		l.Log.Error(err, "Failed to get ConfigMap 'shoot-info' from 'kube-system' namespace")
		return err
	}

	// Check if the projectName is "kyma-stage"
	if projectName, exists := cm.Data["projectName"]; exists && projectName == "kyma-stage" {
		l.Config.RunsOnStage = true
	} else {
		l.Config.RunsOnStage = false
	}

	return nil
}
