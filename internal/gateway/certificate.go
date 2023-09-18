package gateway

import (
	"context"
	_ "embed"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed certificate.yaml
var certificateManifest []byte

func reconcileCertificate(ctx context.Context, k8sClient client.Client, name, domain, certSecretName string) error {

	ctrl.Log.Info("Reconciling Certificate", "Name", name, "Domain", domain, "SecretName", certSecretName)
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Domain"] = domain
	templateValues["SecretName"] = certSecretName

	return reconcileResource(ctx, k8sClient, certificateManifest, templateValues)
}
