package v2alpha1

import (
	"fmt"
	"strings"

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
		Expect(problems[0].Message).To(Equal("Host must be a valid FQDN or short name"))
	})

	It("Should succeed if host is FQDN", func() {
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

	It("Should succeed if host is FQDN label only (short-name)", func() {
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

	It("Should fail if host name has uppercase letters", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("host.exaMple.com")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host must be a valid FQDN or short name"))
	})

	It("Should allow lenghty host name with numbers and dashes", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("host-with-numbers-1234567890-and-dashes----------up-to-63-chars.domain-with-numbers-1234567890-and-dashes--------up-to-63-chars.com")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail if the host FQDN is longer than 253 characters", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("long.%s.com" + strings.Repeat("a", 245))),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host must be a valid FQDN or short name"))
	})

	It("Should fail if any domain label is too long", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host(fmt.Sprintf("%s.com", strings.Repeat("a", 64)))),
					ptr.To(v2alpha1.Host(fmt.Sprintf("host.%s.com", strings.Repeat("a", 64)))),
					ptr.To(v2alpha1.Host(fmt.Sprintf("host.example.%s", strings.Repeat("a", 64)))),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(3))
		for i := 0; i < 3; i++ {
			Expect(problems[i].AttributePath).To(Equal(fmt.Sprintf(".spec.hosts[%d]", i)))
			Expect(problems[i].Message).To(Equal("Host must be a valid FQDN or short name"))
		}
	})

	It("Should fail if any domain label is empty", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host(".domain.com")),
					ptr.To(v2alpha1.Host("host..com")),
					ptr.To(v2alpha1.Host("host.domain.")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(3))
		for i := 0; i < 3; i++ {
			Expect(problems[i].AttributePath).To(Equal(fmt.Sprintf(".spec.hosts[%d]", i)))
			Expect(problems[i].Message).To(Equal("Host must be a valid FQDN or short name"))
		}
	})

	It("Should fail if top level domain is too short", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("too.short.x")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host must be a valid FQDN or short name"))
	})

	It("Should fail if any domain label contain wrong characters", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("wrong-*-inside.example.com")),
					ptr.To(v2alpha1.Host("host.wrong-*-inside.com")),
					ptr.To(v2alpha1.Host("host.example.wrong-*-inside")),
					ptr.To(v2alpha1.Host("*-wrong.example.com")),
					ptr.To(v2alpha1.Host("host.*-wrong.com")),
					ptr.To(v2alpha1.Host("host.example.*-wrong")),
					ptr.To(v2alpha1.Host("wrong-*.example.com")),
					ptr.To(v2alpha1.Host("host.wrong-*.com")),
					ptr.To(v2alpha1.Host("host.example.wrong-*")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(9))
		for i := 0; i < 9; i++ {
			Expect(problems[i].AttributePath).To(Equal(fmt.Sprintf(".spec.hosts[%d]", i)))
			Expect(problems[i].Message).To(Equal("Host must be a valid FQDN or short name"))
		}
	})

	It("Should fail if any segment in host name starts or ends with dash", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{
					ptr.To(v2alpha1.Host("-host.example.com")),
					ptr.To(v2alpha1.Host("host-.example.com")),
					ptr.To(v2alpha1.Host("host.example-.com")),
					ptr.To(v2alpha1.Host("host.-example.com")),
					ptr.To(v2alpha1.Host("host.example.-com")),
					ptr.To(v2alpha1.Host("host.example.com-")),
				},
			},
		}

		//when
		problems := validateHosts(".spec", networkingv1beta1.VirtualServiceList{}, networkingv1beta1.GatewayList{}, apiRule)

		//then
		Expect(problems).To(HaveLen(6))
		for i := 0; i < 6; i++ {
			Expect(problems[i].AttributePath).To(Equal(fmt.Sprintf(".spec.hosts[%d]", i)))
			Expect(problems[i].Message).To(Equal("Host must be a valid FQDN or short name"))
		}
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
