package helpers

import (
	_ "embed"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"strings"
)

const devDomain = "local.kyma.dev"

var expectedCrds = [...]string{
	"dnsproviders.dns.gardener.cloud",
	"dnsentries.dns.gardener.cloud",
	"certificates.cert.gardener.cloud",
	"issuers.cert.gardener.cloud",
}

var configmapGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
var crdGVR = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}

func CheckGardenerCRD(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	log.Printf("Checking Gardener setup")

	for _, crdName := range expectedCrds {
		_, err := resourceMgr.GetResourceWithoutNS(k8sClient, crdGVR, crdName, retry.Attempts(5))
		if err != nil {
			return fmt.Errorf("can't find could not find crd: %s because of error: %s, check your Gardener cluster extensions", crdName, err.Error())
		}
	}

	log.Printf("Gardener setup checked")
	return nil
}

func GetGardenerDomain(resourceMgr *resource.Manager, k8sClient dynamic.Interface) (string, error) {
	cm, err := resourceMgr.GetResource(k8sClient, configmapGVR, "kube-system", "shoot-info", retry.Attempts(5))
	if err != nil {
		return "", fmt.Errorf("can't get shoot-info configmap, error: %w", err)
	}

	domain, _, err := unstructured.NestedString(cm.Object, "data", "domain")
	if err != nil {
		return "", fmt.Errorf("can't get domain from shoot-info configmap, error: %w", err)
	}

	return domain, nil
}

func IsGardenerDetected(resourceMgr *resource.Manager, k8sClient dynamic.Interface) (bool, error) {
	_, err := resourceMgr.GetResource(k8sClient, configmapGVR, "kube-system", "shoot-info", retry.Attempts(2))
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	domain, err := GetGardenerDomain(resourceMgr, k8sClient)
	if err != nil {
		return false, err
	}

	if strings.Contains(domain, devDomain) {
		return false, nil
	}

	return true, nil
}
