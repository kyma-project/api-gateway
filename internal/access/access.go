package access

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/signature"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func ShouldAllowAccessToV1Beta1(ctx context.Context, k8sClient client.Client) (bool, error) {
	var cm corev1.ConfigMap
	err := k8sClient.Get(ctx, client.ObjectKey{Name: "shoot-info", Namespace: "kube-system"}, &cm)
	if err != nil {
		return false, err
	}
	clusterDomain, ok := cm.Data["domain"]
	if !ok {
		return false, nil
	}

	var accessCM corev1.ConfigMap
	err = k8sClient.Get(ctx, client.ObjectKey{Name: "apirule-access", Namespace: "kyma-system"}, &accessCM)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	accessSignature, ok := accessCM.BinaryData["access.sig"]
	if !ok {
		return false, nil
	}

	msg, valid, err := signature.DecryptAndVerifySignature(accessSignature)
	if err != nil || !valid {
		return false, err
	}

	return msg == clusterDomain || strings.TrimSpace(msg) == clusterDomain, nil
}
