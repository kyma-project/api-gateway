package resources

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var sc *runtime.Scheme

var _ = Describe("Resources", func() {
	sc = runtime.NewScheme()
	Expect(v1beta1.AddToScheme(sc)).To(Succeed())
	Expect(networkingv1alpha3.AddToScheme(sc)).To(Succeed())
	Expect(networkingv1beta1.AddToScheme(sc)).To(Succeed())

	DescribeTable(
		"FindUserCreatedResourcesDescribe",
		func(ctx context.Context, logger logr.Logger, c client.Client, configuration resourceFinderConfiguration, conditionResult bool, want []Resource, wantErr bool) {
			i := &ResourcesFinder{
				ctx:           ctx,
				logger:        logger,
				client:        c,
				configuration: configuration,
			}
			got, err := i.FindUserCreatedResources(func(ctx context.Context, c client.Client, res Resource) bool { return conditionResult })
			Expect(err != nil).To(Equal(wantErr))
			Expect(got).To(BeEquivalentTo(want))
		},
		Entry("Should get nothing if there are only managed API-Gateway resources present", context.Background(),
			logr.Discard(),
			fake.NewClientBuilder().WithScheme(sc).WithObjects(&v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "managed-namespace",
					Name:      "controlled-apirule",
				},
			}).Build(),
			resourceFinderConfiguration{Resources: []ResourceConfiguration{
				{
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "gateway.kyma-project.io",
						Version: "v1beta1",
						Kind:    "APIRule",
					},
					ControlledList: []ResourceMeta{
						{
							Name:      "controlled-apirule",
							Namespace: "managed-namespace",
						},
					},
				},
			},
			},
			true,
			nil,
			false,
		),
		Entry("Should get resource if there is a customer resource present", context.Background(),
			logr.Discard(),
			fake.NewClientBuilder().WithScheme(sc).WithObjects(&v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "managed-namespace",
					Name:      "controlled-apirule",
				},
			}, &v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "unmanaged-namespace",
					Name:      "user-apirule",
				},
			}).Build(),
			resourceFinderConfiguration{Resources: []ResourceConfiguration{
				{
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "gateway.kyma-project.io",
						Version: "v1beta1",
						Kind:    "APIRule",
					},
					ControlledList: []ResourceMeta{
						{
							Namespace: "managed-namespace",
							Name:      "controlled-apirule",
						},
					},
				},
			},
			},
			true,
			[]Resource{
				{
					ResourceMeta: ResourceMeta{
						Namespace: "unmanaged-namespace",
						Name:      "user-apirule",
					}, GVK: schema.GroupVersionKind{
						Group:   "gateway.kyma-project.io",
						Version: "v1beta1",
						Kind:    "APIRule",
					},
				},
			},
			false,
		),
		Entry("Should get resource if there is a customer resource present in a specific version only", context.Background(),
			logr.Discard(),
			fake.NewClientBuilder().WithScheme(sc).WithObjects(&networkingv1beta1.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "unmanaged-namespace",
					Name:      "user-vs",
				},
			}).Build(),
			resourceFinderConfiguration{Resources: []ResourceConfiguration{
				{
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "networking.istio.io",
						Version: "v1alpha3",
						Kind:    "VirtualService",
					},
				},
				{
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "networking.istio.io",
						Version: "v1beta1",
						Kind:    "VirtualService",
					},
				},
			},
			},
			true,
			[]Resource{
				{
					ResourceMeta: ResourceMeta{
						Namespace: "unmanaged-namespace",
						Name:      "user-vs",
					}, GVK: schema.GroupVersionKind{
						Group:   "networking.istio.io",
						Version: "v1beta1",
						Kind:    "VirtualService",
					},
				},
			},
			false,
		),
		Entry("Should not get resource if there is a customer resource present but condition check returns false", context.Background(),
			logr.Discard(),
			fake.NewClientBuilder().WithScheme(sc).WithObjects(&v1beta1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "unmanaged-namespace",
					Name:      "user-apirule",
				},
			}).Build(),
			resourceFinderConfiguration{Resources: []ResourceConfiguration{
				{
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "gateway.kyma-project.io",
						Version: "v1beta1",
						Kind:    "APIRule",
					},
				},
			},
			},
			false,
			nil,
			false,
		),
	)
})

var _ = Describe("NewResourcesFinderFromConfigYaml", func() {
	It("Should read configuration from yaml", func() {
		config, err := NewResourcesFinderFromConfigYaml(context.Background(), nil, logr.Logger{}, "test_assets/test_resources_list.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(config).To(BeEquivalentTo(&ResourcesFinder{
			ctx:    context.Background(),
			logger: logr.Logger{},
			client: nil,
			configuration: resourceFinderConfiguration{
				Resources: []ResourceConfiguration{
					{
						GroupVersionKind: schema.GroupVersionKind{
							Group:   "gateway.kyma-project.io",
							Version: "v1beta1",
							Kind:    "APIRule",
						},
						ControlledList: []ResourceMeta{
							{
								Name:      "api-rule-\\d\\.\\d\\d",
								Namespace: "managed-namespace",
							},
						},
					},
				},
			},
		},
		))
	})

	It("Should fail if the configuration contains invalid regex", func() {
		_, err := NewResourcesFinderFromConfigYaml(context.Background(), nil, logr.Logger{}, "test_assets/test_wrong_resources_list.yaml")
		Expect(err).To(HaveOccurred())
	})
})

var _ = DescribeTable("contains", func(a []ResourceMeta, b Resource, should bool) {
	Expect(contains(a, b.ResourceMeta)).To(Equal(should))
},
	Entry("Should return true if the array contains the resource", []ResourceMeta{{Name: "test", Namespace: "test-ns"}},
		Resource{ResourceMeta: ResourceMeta{
			Name:      "test",
			Namespace: "test-ns",
		}}, true),
	Entry("Should return false if the array doesn't contain the resource", []ResourceMeta{{Name: "test", Namespace: "test-ns"}},
		Resource{ResourceMeta: ResourceMeta{
			Name:      "test",
			Namespace: "test",
		}}, false))

func Test_contains(t *testing.T) {
	type args struct {
		s []ResourceMeta
		e Resource
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				s: []ResourceMeta{
					{Name: "test", Namespace: "test-ns"},
				},
				e: Resource{ResourceMeta: ResourceMeta{
					Name:      "test",
					Namespace: "test-ns",
				}},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := contains(tt.args.s, tt.args.e.ResourceMeta)
			if got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
			if err != nil != tt.wantErr {
				t.Errorf("error happened = %v, wanted %v", err != nil, tt.wantErr)
			}
		})
	}
}
