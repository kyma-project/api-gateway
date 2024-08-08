package v2alpha1

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validate hosts", func() {
	It("Should fail if any host that is occupied by any Virtual Service exposed by another resource", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host.example.com"),
					getHostPtr("occupied.example.com"),
				},
			},
		}
		virtualService1 := &networkingv1beta1.VirtualService{
			Spec: v1beta1.VirtualService{
				Hosts: []string{
					"not-occupied1.example.com",
					"not-occupied2.example.com"},
			},
		}
		virtualService2 := &networkingv1beta1.VirtualService{
			Spec: v1beta1.VirtualService{
				Hosts: []string{
					"not-occupied3.example.com",
					"occupied.example.com"},
			},
		}

		virtualServiceList := networkingv1beta1.VirtualServiceList{
			Items: []*networkingv1beta1.VirtualService{
				virtualService1,
				virtualService2,
			},
		}

		//when
		problems := validateHosts(".spec", virtualServiceList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[1]"))
		Expect(problems[0].Message).To(Equal("Host is occupied by another Virtual Service"))
	})

	It("Should not fail if a host is occupied by the Virtual Service related to the same API Rule", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host.example.com"),
					getHostPtr("occupied.example.com"),
				},
			},
		}
		virtualService1 := &networkingv1beta1.VirtualService{
			Spec: v1beta1.VirtualService{
				Hosts: []string{
					"not-occupied1.example.com",
					"not-occupied2.example.com"},
			},
		}
		virtualService2 := &networkingv1beta1.VirtualService{
			Spec: v1beta1.VirtualService{
				Hosts: []string{
					"not-occupied3.example.com",
					"occupied.example.com"},
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: getMapWithOwnerLabel(apiRule),
			},
		}
		virtualServiceList := networkingv1beta1.VirtualServiceList{
			Items: []*networkingv1beta1.VirtualService{
				virtualService1,
				virtualService2,
			},
		}

		//when
		problems := validateHosts(".spec", virtualServiceList, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})
})

func getHostPtr(hostName string) *v2alpha1.Host {
	host := v2alpha1.Host(hostName)
	return &host
}

func getMapWithOwnerLabel(apiRule *v2alpha1.APIRule) map[string]string {
	labelKey, labelValue := getExpectedOwnerLabel(apiRule)
	return map[string]string{
		labelKey: labelValue,
	}
}
