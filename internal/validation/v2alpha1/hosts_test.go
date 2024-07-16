package v2alpha1

import (
	"fmt"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validate hosts", func() {

	It("Should fail if there are no hosts defined", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts"))
		Expect(problems[0].Message).To(Equal("No hosts defined"))
	})

	It("Should fail if host is empty", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr(""),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host is not fully qualified domain name"))
	})

	It("Should pass if host is FQDN", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host.example.com"),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail if host is not FQDN", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host-without-domain"),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host is not fully qualified domain name"))
	})

	It("Should allow lenghty host name with numbers and dashes", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host-with-numbers-123-and-dashes.long-domain-with-numbers-123-and-dashes.com"),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail if host name contains wrong characters", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host-with-wrong-character-like-*.example.com"),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host is not fully qualified domain name"))
	})

	It("Should fail if any segment in host name starts or ends with dash", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("-host.example.com"),
					getHostPtr("host-.example.com"),
					getHostPtr("host.example-.com"),
					getHostPtr("host.example.-com"),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(4))
		for i := 0; i < 4; i++ {
			Expect(problems[i].AttributePath).To(Equal(fmt.Sprintf(".spec.hosts[%d]", i)))
			Expect(problems[i].Message).To(Equal("Host is not fully qualified domain name"))
		}
	})

	It("Should fail if any host is not FQDN", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("host.example.com"),
					getHostPtr("host-without-domain"),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[1]"))
		Expect(problems[0].Message).To(Equal("Host is not fully qualified domain name"))
	})

	It("Should fail for a host that is occupied by a Virtual Service exposed by another resource", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					getHostPtr("occupied.example.com"),
				},
			},
		}
		virtualService := &networkingv1beta1.VirtualService{
			Spec: v1beta1.VirtualService{
				Hosts: []string{"occupied.example.com"},
			},
		}
		virtualServiceList := networkingv1beta1.VirtualServiceList{
			Items: []*networkingv1beta1.VirtualService{virtualService},
		}

		//when
		problems := validateHosts(".spec", virtualServiceList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host is occupied by another Virtual Service"))
	})

	It("Should fail for any host that is occupied by any Virtual Service exposed by another resource", func() {
		//given
		apiRule := &v2alpha1.APIRule{
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

	It("Should not fail if host is occupied by the Virtual Service exposed by another resource", func() {
		//given
		apiRule := &v2alpha1.APIRule{
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
