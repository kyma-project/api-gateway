package gateway

import (
	"context"
	"github.com/go-logr/logr"
	"os"
	"os/exec"
	"sync"
)

type APIRuleReconcilerStarter struct {
	isStarted bool
	options   StarterOptions
	*sync.Mutex
	cmd *exec.Cmd
}

type StarterOptions struct {
	setupLog logr.Logger
}

func NewAPIRuleReconcilerStarter(
	setupLog logr.Logger,
) *APIRuleReconcilerStarter {
	return &APIRuleReconcilerStarter{
		isStarted: false,
		options: StarterOptions{
			setupLog: setupLog,
		},
		Mutex: &sync.Mutex{},
	}
}

func (r *APIRuleReconcilerStarter) SetupAndStartManager() error {
	r.Lock()
	defer r.Unlock()

	if r.isStarted {
		r.options.setupLog.Info("APIRule reconciler is already started, skipping setup")
		return nil
	}

	r.options.setupLog.Info("Starting APIRule reconciler manager")
	cmd := exec.CommandContext(context.Background(), "./apirule-manager")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		r.options.setupLog.Error(err, "Failed to start APIRule reconciler manager")
	}

	r.cmd = cmd
	r.isStarted = true
	r.options.setupLog.Info("Succesfully started APIRule reconciler")
	return nil
}

func (r *APIRuleReconcilerStarter) StopManager() error {
	r.Lock()
	defer r.Unlock()

	r.options.setupLog.Info("Stopping APIRule reconciler manager")
	if !r.isStarted {
		r.options.setupLog.Info("APIRule reconciler is not started, nothing to stop")
		return nil
	}
	if r.cmd != nil {
		if err := r.cmd.Process.Kill(); err != nil {
			r.options.setupLog.Error(err, "Failed to stop APIRule reconciler manager")
			return err
		}
		r.options.setupLog.Info("APIRule reconciler manager stopped successfully")
		r.isStarted = false
	}

	return nil
}
