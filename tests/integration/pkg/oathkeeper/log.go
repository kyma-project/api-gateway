package oathkeeper

import (
	"bytes"
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func LogInfo() {
	conf := config.GetConfigOrDie()
	k8sClient := kubernetes.NewForConfigOrDie(conf)
	ctx := context.Background()

	pods, err := k8sClient.CoreV1().Pods("kyma-system").List(ctx, v12.ListOptions{
		LabelSelector: "app.kubernetes.io/name=oathkeeper",
	})
	if err != nil {
		log.Printf("Fetching Oathkeeper pods for logging: %s", err.Error())
	}

	for _, pod := range pods.Items {

		log.Printf("Pod %s status: %s", pod.Name, pod.Status.Phase)
		for _, condition := range pod.Status.Conditions {
			log.Printf("Pod %s condition: '%s', status: '%s', reasson: '%s' message: '%s'", pod.Name,
				condition.Type, condition.Status, condition.Reason, condition.Message)
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			log.Printf("Pod %s container %s status: %v", pod.Name, containerStatus.Name, containerStatus.State)
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
		log.Printf("Fetching %s container logs: %s", container, err.Error())
	}
	defer func() {
		e := logs.Close()
		if e != nil {
			log.Printf("error %s closing logs stream: %s", container, err.Error())
		}
	}()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		log.Printf("Copying %s container logs: %s", container, err.Error())
	}
	str := buf.String()
	log.Printf("Logs for container %s:", container)
	log.Println(str)
}
