package v2alpha1

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Validate hosts", func() {
	It("Should fail if there are no hosts defined", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

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
					ptr.To(v2alpha1.Host("")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host must be a valid FQDN or short host name"))
	})

	It("Should succeed if host is a valid FQDN", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("host.com")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed if host is a short host name", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To("gateway-name"),
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("short-name-host")),
				},
			},
		}

		gwList := networkingv1beta1.GatewayList{
			Items: []*networkingv1beta1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateway-name",
					},
					Spec: v1beta1.Gateway{
						Servers: []*v1beta1.Server{
							{
								Hosts: []string{"*.example.com"},
							},
						},
					},
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, gwList, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail if host is a short host name and referenced Gateway is missing", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To("gateway-name"),
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("short-name-host")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Unable to find Gateway gateway-name"))
	})

	It("Should fail if host is a short host name and referenced Gateway has various hosts definitions", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Gateway: ptr.To("gateway-name"),
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("short-name-host")),
				},
			},
		}

		gwList := networkingv1beta1.GatewayList{
			Items: []*networkingv1beta1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateway-name",
					},
					Spec: v1beta1.Gateway{
						Servers: []*v1beta1.Server{
							{
								Hosts: []string{"*.example.com"},
							},
							{
								Hosts: []string{"*.example2.com"},
							},
						},
					},
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, gwList, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Lowercase RFC 1123 label is only supported as the APIRule host when selected Gateway has a single host definition matching *.<fqdn> format"))
	})

	It("Should fail if any host that is occupied by any Virtual Service exposed by another resource", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-name",
				Namespace: "some-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("host.example.com")),
					ptr.To(v2alpha1.Host("occupied.example.com")),
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
		problems := validateHosts(".spec", virtualServiceList, networkingv1beta1.GatewayList{}, apiRule)

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
					ptr.To(v2alpha1.Host("host.example.com")),
					ptr.To(v2alpha1.Host("occupied.example.com")),
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
		problems := validateHosts(".spec", virtualServiceList, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})
})

func getMapWithOwnerLabel(apiRule *v2alpha1.APIRule) map[string]string {
	labelKey, labelValue := getExpectedOwnerLabel(apiRule)
	return map[string]string{
		labelKey: labelValue,
	}
}
