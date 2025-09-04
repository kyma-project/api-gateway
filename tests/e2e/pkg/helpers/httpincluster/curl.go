package httpincluster

import (
	"bytes"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	podName       = "curl"
	containerName = "curl"
)

type Options struct {
	Method string
}

func WithMethod(method string) Option {
	return func(o *Options) {
		o.Method = method
	}
}

type Option func(o *Options)

func RunRequestFromInsideCluster(t *testing.T, namespace string, url string, options ...Option) (string, string, error) {
	t.Helper()
	opts := &Options{
		Method: http.MethodGet,
	}
	for _, opt := range options {
		opt(opts)
	}

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Could not create resources client: err=%s", err)
		return "", "", err
	}

	curlPodName := envconf.RandomName(podName, 16)
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: curlPodName, Namespace: namespace},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Args:  []string{"sleep", "infinity"},
					Image: "curlimages/curl:8.14.1",
					Name:  containerName,
				},
			},
		},
	}

	t.Logf("Creating curl pod: %s", curlPodName)
	err = r.Create(t.Context(), &pod)
	if err != nil {
		return "", "", err
	}
	setup.DeclareCleanup(t, func() {
		t.Log("Deleting pod: ", pod.Name)
		err := r.Delete(setup.GetCleanupContext(), &pod)
		if err != nil {
			t.Logf("Failed to delete pod %s: %v", pod.Name, err)
		}
	})
	err = wait.For(conditions.New(r).PodRunning(&pod))
	if err != nil {
		return "", "", err
	}

	cmd := []string{"curl", "-ik", "-sSL", "-m", "10", "-X", opts.Method, "--fail-with-body", url}

	var stdout, stderr bytes.Buffer
	err = r.ExecInPod(t.Context(), pod.GetNamespace(), pod.GetName(), containerName, cmd, &stdout, &stderr)
	stdOutStr := strings.TrimSpace(stdout.String())
	stdErrStr := strings.TrimSpace(stderr.String())
	t.Logf("[%s] stdout: %v", curlPodName, stdOutStr)
	t.Logf("[%s] stderr: %v", curlPodName, stdErrStr)

	return stdOutStr, stdErrStr, err
}
