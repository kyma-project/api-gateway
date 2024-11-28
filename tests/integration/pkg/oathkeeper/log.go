package oathkeeper

import (
	"bufio"
	"bytes"
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var logger = log.New(os.Stderr, "oathkeeper-log ", log.LstdFlags)

func LogInfo() {
	conf := config.GetConfigOrDie()
	k8sClient := kubernetes.NewForConfigOrDie(conf)
	ctx := context.Background()

	pods, err := k8sClient.CoreV1().Pods("kyma-system").List(ctx, v12.ListOptions{
		LabelSelector: "app.kubernetes.io/name=oathkeeper",
	})
	if err != nil {
		logger.Printf("Fetching Oathkeeper pods for logging: %s", err.Error())
	}

	for _, pod := range pods.Items {

		logger.Printf("Pod %s status: %s", pod.Name, pod.Status.Phase)
		for _, condition := range pod.Status.Conditions {
			logger.Printf("Pod %s condition: '%s', status: '%s', reason: '%s' message: '%s'", pod.Name,
				condition.Type, condition.Status, condition.Reason, condition.Message)
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			logger.Printf("Pod %s container %s status: %v", pod.Name, containerStatus.Name, containerStatus.State)
		}

		writeLogs(ctx, k8sClient, pod, "oathkeeper")
		writeLogs(ctx, k8sClient, pod, "oathkeeper-maester")
		writeLogs(ctx, k8sClient, pod, "init")
	}
}

func writeLogs(ctx context.Context, c *kubernetes.Clientset, pod corev1.Pod, container string) {

	req := c.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: container,
	})
	logs, err := req.Stream(ctx)
	if err != nil {
		logger.Printf("Fetching %s container logs: %s", container, err.Error())
	}
	defer func() {
		e := logs.Close()
		if e != nil {
			logger.Printf("error %s closing logs stream: %s", container, err.Error())
		}
	}()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		logger.Printf("Copying %s container logs: %s", container, err.Error())
	}

	scanner := bufio.NewScanner(buf)
	logsLogger := log.New(os.Stderr, container+"-container ", log.LstdFlags)
	for scanner.Scan() {
		logsLogger.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("Reading %s container logs: %s", container, err.Error())
	}
}
