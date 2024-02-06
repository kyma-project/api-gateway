package network

import (
	"context"
	_ "embed"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/dynamic"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

//go:embed coredns-custom.yml
var coreDnsCustomManifest []byte

func CreateKymaLocalDnsRewrite(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	log.Printf("Applying custom CoreDNS extension for local.kyma.dev")

	// We use the ParseYaml function instead of yaml.Unmarshal, because the latter doesn't handle indentation properly.
	resources, err := manifestprocessor.ParseYaml(coreDnsCustomManifest)
	if err != nil {
		return err
	}

	_, err = resourceMgr.CreateResources(k8sClient, resources...)
	if err != nil {
		return err
	}

	return restartCoreDnsPods()
}

func DeleteKymaLocalDnsRewrite(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	log.Printf("Cleaning up custom CoreDNS extension")

	resources, err := manifestprocessor.ParseYaml(coreDnsCustomManifest)
	if err != nil {
		return err
	}

	err = resourceMgr.DeleteResources(k8sClient, resources...)
	if err != nil {
		return err
	}

	return restartCoreDnsPods()
}

func restartCoreDnsPods() error {
	c := k8sclient.GetK8sClient()
	dep := &appsv1.Deployment{}
	err := c.Get(context.Background(), client.ObjectKey{Name: "coredns", Namespace: "kube-system"}, dep)
	if err != nil {
		return err
	}

	patch := client.StrategicMergeFrom(dep.DeepCopy())

	if dep.Spec.Template.ObjectMeta.Annotations == nil {
		dep.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	dep.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	return c.Patch(context.Background(), dep, patch)
}
