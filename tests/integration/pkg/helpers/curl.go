package helpers

import (
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	defaultWaitBackoff = wait.Backoff{
		Cap:      10 * time.Minute,
		Duration: time.Second,
		Steps:    30,
		Factor:   1.5,
	}
)

func RunCurlInPod(namespace string, command []string) ([]byte, error) {
	ctx := context.Background()
	cfg, err := client.GetK8SConfig()
	if err != nil {
		return nil, err
	}
	
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	run := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "curl-",
			Namespace:    namespace,
			Labels:       map[string]string{"test-workload": "true"},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "curl",
					Image:   "curlimages/curl",
					Command: command,
				},
			},
		},
	}

	run, err = c.CoreV1().Pods(namespace).Create(ctx, run, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	var exitCode int32
	err = wait.ExponentialBackoffWithContext(ctx, defaultWaitBackoff, func(ctx context.Context) (bool, error) {
		p, err := c.CoreV1().Pods(namespace).Get(ctx, run.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(p.Status.ContainerStatuses) == 0 {
			return false, nil
		}
		for _, s := range p.Status.ContainerStatuses {
			if s.Name == "curl" && s.State.Terminated != nil {
				exitCode = s.State.Terminated.ExitCode
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	out, err := c.CoreV1().Pods(namespace).GetLogs(run.Name, &corev1.PodLogOptions{Container: "curl"}).DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	if exitCode != 0 {
		return out, fmt.Errorf("non-zero exit code: %v", exitCode)
	}

	return out, nil
}
