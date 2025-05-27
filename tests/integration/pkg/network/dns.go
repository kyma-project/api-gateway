package network

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/avast/retry-go/v4"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
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

	_, err = resourceMgr.CreateOrUpdateResources(k8sClient, resources...)
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

	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = make(map[string]string)
	}
	dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	return c.Patch(context.Background(), dep, patch)
}

// function waits until the given domain points to the given IP
// the domain must be a wildcard one because the function probes the <randomHost>.<domain>
func WaitUntilDNSReady(domain string, expectedIP net.IP, retryOpts []retry.Option) error {
	return retry.Do(func() error {
		ready, err := isDNSReady(domain, expectedIP)
		if err != nil {
			return fmt.Errorf("error while checking if domain %s is ready: %w", domain, err)
		}
		if !ready {
			return fmt.Errorf("domain %s is not ready yet", domain)
		}
		return nil
	}, retryOpts...)
}

// function checks whether the given domain points to the given IP
// the domain must be a wildcard one because the function probes the <randomHost>.<domain>
func isDNSReady(domain string, expectedIP net.IP) (bool, error) {
	randomHost := helpers.GenerateRandomString(3)
	ips, err := net.LookupIP(fmt.Sprintf("%s.%s", randomHost, domain))
	if err != nil {
		return false, nil
	}
	if len(ips) != 0 {
		for _, ip := range ips {
			if ip.Equal(expectedIP) {
				return true, nil
			}
		}
	}
	return false, err
}
