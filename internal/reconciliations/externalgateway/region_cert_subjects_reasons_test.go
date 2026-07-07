package externalgateway

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

func TestResolveRegionCertSubjects_ReasonedErrors(t *testing.T) {
	cases := []struct {
		name          string
		configMapData map[string]string
		configMapName string
		region        string
		wantReason    string
		omitConfigMap bool
	}{
		{
			name:          "RegionsConfigMap not found",
			omitConfigMap: true,
			configMapName: "missing",
			region:        "eu10",
			wantReason:    externalv1alpha1.ReasonRegionsConfigMapNotFound,
		},
		{
			name: "RegionsConfigMap multi-key without regions.yaml",
			configMapData: map[string]string{
				"other-key-1": "x",
				"other-key-2": "y",
			},
			configMapName: "regions",
			region:        "eu10",
			wantReason:    externalv1alpha1.ReasonRegionsConfigMapKeyAmbiguous,
		},
		{
			name:          "RegionsConfigMap empty",
			configMapData: map[string]string{},
			configMapName: "regions",
			region:        "eu10",
			wantReason:    externalv1alpha1.ReasonRegionsConfigMapInvalid,
		},
		{
			name: "RegionsConfigMap invalid YAML",
			configMapData: map[string]string{
				"regions.yaml": "this is not: valid: yaml: :",
			},
			configMapName: "regions",
			region:        "eu10",
			wantReason:    externalv1alpha1.ReasonRegionsConfigMapInvalid,
		},
		{
			name: "RegionsConfigMap has zero regions",
			configMapData: map[string]string{
				"regions.yaml": "regions: []",
			},
			configMapName: "regions",
			region:        "eu10",
			wantReason:    externalv1alpha1.ReasonRegionsConfigMapInvalid,
		},
		{
			name: "region not in ConfigMap",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - name: eu10
    subjects:
      - "C=US, CN=eu10"
`,
			},
			configMapName: "regions",
			region:        "us99",
			wantReason:    externalv1alpha1.ReasonRegionNotFound,
		},
		{
			name: "region has zero subjects",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - name: eu10
    subjects: []
`,
			},
			configMapName: "regions",
			region:        "eu10",
			wantReason:    externalv1alpha1.ReasonRegionHasNoSubjects,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			_ = externalv1alpha1.AddToScheme(scheme)

			builder := fake.NewClientBuilder().WithScheme(scheme)
			if !tc.omitConfigMap {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tc.configMapName,
						Namespace: "ns",
					},
					Data: tc.configMapData,
				}
				builder = builder.WithObjects(cm)
			}
			c := builder.Build()

			eg := &externalv1alpha1.ExternalGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
				Spec: externalv1alpha1.ExternalGatewaySpec{
					Region:           tc.region,
					RegionsConfigMap: tc.configMapName,
				},
			}

			_, err := ResolveRegionCertSubjects(context.Background(), c, eg)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			reason, ok := ErrorReason(err)
			if !ok || reason != tc.wantReason {
				t.Fatalf("expected reason %s, got reason=%q ok=%v (message: %s)", tc.wantReason, reason, ok, err.Error())
			}
		})
	}
}
