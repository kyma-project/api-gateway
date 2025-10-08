package setup

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"log"
	"os"
	"path"
	"sigs.k8s.io/yaml"
	"testing"
	"time"
)

const (
	podLogsDir     = "pods"
	podLogFileName = "%s@%s.log"
)

var logsTimeStamp = time.Now().In(time.FixedZone("CET", 2*60*60)).Format("02_01_2006-15_04_05CET")
var basePath = path.Join(".", "logs")

func DumpClusterResources(t *testing.T) {
	t.Helper()
	if githubWorkspace, ok := os.LookupEnv("GITHUB_WORKSPACE"); ok {
		basePath = path.Join(githubWorkspace, "logs")
	}
	dumpPath := path.Join(basePath, logsTimeStamp, t.Name(), "resources")
	_, err := os.Stat(dumpPath)
	if !os.IsNotExist(err) {
		return
	}

	dir := os.MkdirAll(dumpPath, 0o755)
	storeLogsFromAllPods(t)

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Could not create resources client: err=%s", err)
		return
	}
	discoClient, err := discovery.NewDiscoveryClientForConfig(r.GetConfig())
	if err != nil {
		log.Fatalf("Error creating discovery client: %v", err)
	}
	serverPreferredResources, err := discovery.ServerPreferredResources(discoClient)
	if err != nil {
		t.Logf("Could not get server preferred resources: err=%s", err)
		return
	}

	t.Logf("Found %d server preferred resources", len(serverPreferredResources))
	t.Logf("Dumping all cluster resources to %s", dumpPath)
	for _, serverPreferredResource := range serverPreferredResources {
		for _, apiResource := range serverPreferredResource.APIResources {
			resourceName := apiResource.Name
			var unstructuredList = &unstructured.UnstructuredList{}
			groupVersion, err := schema.ParseGroupVersion(serverPreferredResource.GroupVersion)
			if err != nil {
				t.Logf("Could not parse group version: err=%s", err)
				continue
			}
			unstructuredList.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   groupVersion.Group,
				Version: groupVersion.Version,
				Kind:    apiResource.Kind,
			})
			err = r.List(context.Background(), unstructuredList)
			if err != nil {
				t.Logf("Could not list resources for CRD %s: err=%s", resourceName, err)
				continue
			}
			fileName := path.Join(dumpPath, resourceName)
			if dir != nil {
				t.Logf("Could not create log directory: err=%s", dir)
				continue
			}
			fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				t.Logf("Could not open log file: err=%s", err)
				continue
			}
			resourceYaml, err := yaml.Marshal(unstructuredList)
			if err != nil {
				t.Logf("Could not marshal resource to JSON: err=%s", err)
				continue
			}
			_, err = fileHandle.WriteString(string(resourceYaml))
			if err != nil {
				t.Logf("Could not write to log file: err=%s", err)
				continue
			}

			err = fileHandle.Close()
			if err != nil {
				t.Logf("Could not close log file: err=%s", err)
			}
		}
	}
}

func storeLogsFromAllPods(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(path.Join(basePath, logsTimeStamp, t.Name(), podLogsDir)); !os.IsNotExist(err) {
		return
	}
	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Could not create resources client: err=%s", err)
		return
	}
	podList := &corev1.PodList{}
	err = r.List(context.Background(), podList)
	if err != nil {
		t.Logf("Could not list pods: err=%s", err)
		return
	}
	p := path.Join(basePath, logsTimeStamp, t.Name(), podLogsDir)
	err = os.MkdirAll(p, 0o755)
	if err != nil {
		t.Logf("Could not create log directory: err=%s", err)
		return
	}
	for _, pod := range podList.Items {
		err := storeLogsFromPodToFile(t, pod.Namespace, pod.Name)
		if err != nil {
			t.Logf("Could not get logs from pod %s in namespace %s: err=%s", pod.Name, pod.Namespace, err)
		}
	}
}

func storeLogsFromPodToFile(t *testing.T, namespace, podName string) error {
	t.Helper()
	fileName := path.Join(basePath, logsTimeStamp, t.Name(), podLogsDir, fmt.Sprintf(podLogFileName, podName, namespace))

	clientSet, err := client.GetClientSet(t)
	if err != nil {
		t.Logf("Could not get client set: err=%s", err)
		return err
	}
	logsStream, err := clientSet.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Timestamps: true,
	}).Stream(context.Background())
	if err != nil {
		t.Logf("Could not get logs stream: err=%s", err)
		return err
	}
	defer func() {
		err := logsStream.Close()
		if err != nil {
			t.Logf("Could not close logs stream: err=%s", err)
		}
	}()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(logsStream)
	if err != nil {
		t.Logf("Could not read logs stream: err=%s", err)
		return err
	}

	fileHandle, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Logf("Could not open log file: err=%s", err)
		return err
	}
	defer func() {
		err := fileHandle.Close()
		if err != nil {
			t.Logf("Could not close log file: err=%s", err)
		}
	}()
	_, err = fileHandle.WriteString(buf.String())
	if err != nil {
		t.Logf("Could not write to log file: err=%s", err)
		return err
	}
	return nil
}
