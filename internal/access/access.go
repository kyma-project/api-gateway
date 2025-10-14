package access

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/signature"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
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
	ctrl.Log.Info("access signature", "signature", accessSignature, "ok", ok)
	if !ok {
		return false, nil
	}

	msg, valid, err := signature.DecryptAndVerifySignature(accessSignature)
	ctrl.Log.Info("decrypted message", "msg", msg, "valid", valid, "err", err)
	if err != nil || !valid {
		if err != nil {
			ctrl.Log.Error(err, "failed to decrypt message")
		}
		return false, err
	}
	msg = strings.TrimSpace(msg)
	clusterDomain = strings.TrimSpace(clusterDomain)
	ctrl.Log.Info("trim", "clusterDomain", clusterDomain, "msg", msg)

	match, err := regexp.MatchString(`^\*\.[^*]*$`, msg)
	if err != nil {
		return false, err
	}
	if match {
		ctrl.Log.Info("wildcard regex match", "msg", msg)
		return strings.HasSuffix(clusterDomain, strings.TrimPrefix(msg, "*.")), nil
	}
	ctrl.Log.Info("wildcard regex not match", "msg", msg)
	return strings.TrimSpace(msg) == clusterDomain, nil
}
