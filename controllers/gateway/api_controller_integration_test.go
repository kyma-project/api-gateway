package gateway_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests needs to be executed serially because of the shared state of the JWT Handler in the API Controller.
var _ = Describe("APIRule Controller", Serial, func() {
	const (
		testNameBase               = "test"
		testIDLength               = 5
		testServiceNameBase        = "httpbin"
		testServicePort     uint32 = 443
	)

	Context("when creating APIRule in version v2alpha1", Ordered, func() {
		Context("respect x-validation rules only", Ordered, func() {
			It("should be able to create an APIRule with noAuth=true", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should be able to create an APIRule with jwt", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should be able to create an APIRule with jwt and noAuth=false", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(false)
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should be able to create an APIRule with jwt and mutators", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should fail to create an APIRule without noAuth and jwt", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("One of the following fields must be set: noAuth, jwt, extAuth"))
			})

			It("should fail to create an APIRule with noAuth=false", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(false)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("One of the following fields must be set: noAuth, jwt, extAuth"))
			})

			It("should fail to create an APIRule with jwt and noAuth=true", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				rule.Jwt = &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{},
					Authorizations:  []*gatewayv2alpha1.JwtAuthorization{},
				}
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("One of the following fields must be set: noAuth, jwt, extAuth"))
			})

			It("should fail to create an APIRule with more than one host", func() {
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("httpbin-istio-jwt-happy-base.kyma.local")
				secondServiceHost := gatewayv2alpha1.Host("other-istio-jwt-happy-base.kyma.local")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost, &secondServiceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.hosts: Too many: 2: must have at most 1 items"))
			})
		})

		Context("gateway name should be valid", Ordered, func() {
			It("should create an APIRule with a valid gateway", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1Gateway(apiRuleName, testNamespace, serviceName, testNamespace, testGatewayURL, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			invalidHelper := func(gatewayName string) {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1Gateway(apiRuleName, testNamespace, serviceName, testNamespace, gatewayName, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.gateway: Invalid value: \"string\": Gateway must be in the namespace/name format"))
			}

			It("should not create an APIRule with an empty gateway", func() {
				invalidHelper("")
			})

			It("should not create an APIRule with too long gateway namespace name", func() {
				invalidHelper("insane-very-long-namespace-name-exceeding-sixty-three-characters/validname")
			})

			It("should not create an APIRule with too long gateway name", func() {
				invalidHelper("validnamespace/insane-very-long-namespace-name-exceeding-sixty-three-characters")
			})

			It("should not create an APIRule with just the namespace", func() {
				invalidHelper("validnamespace/")
			})

			It("should not create an APIRule with just the gateway name", func() {
				invalidHelper("/validgateway")
			})

			It("should not create an APIRule with double slashed gateway name", func() {
				invalidHelper("namespace//gateway")
			})
		})

		Context("hosts should be a valid FQDN or a short host name", Ordered, func() {
			It("should create an APIRule with a valid FQDN host", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("test.some-example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should create an APIRule with short host name that has length of 1 character", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("a")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should create an APIRule with host name that has 1 char labels and 2 chars top-level domain", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHost := gatewayv2alpha1.Host("a.b.ca")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should create an APIRule with host name that has length of 255 characters", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				sixtyThreeA := strings.Repeat("a", 63)
				host255 := fmt.Sprintf("%s.%s.%s.%s.com", sixtyThreeA, sixtyThreeA, sixtyThreeA, strings.Repeat("b", 59))
				serviceHost := gatewayv2alpha1.Host(host255)
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(host255).To(HaveLen(255))
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())
			})

			It("should not create an APIRule with host name longer than 255 characters", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				sixtyThreeA := strings.Repeat("a", 63)
				host256 := fmt.Sprintf("%s.%s.%s.%s.com", sixtyThreeA, sixtyThreeA, sixtyThreeA, strings.Repeat("b", 60))
				serviceHost := gatewayv2alpha1.Host(host256)
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(host256).To(HaveLen(256))
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.hosts[0]: Too long: may not be longer than 255"))
			})

			invalidHelper := func(host gatewayv2alpha1.Host) {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := testServiceNameBase
				serviceHosts := []*gatewayv2alpha1.Host{ptr.To(gatewayv2alpha1.Host(host))}

				rule := testRulev2alpha1("/img", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()

				// when
				Expect(c.Create(context.Background(), svc)).Should(Succeed())
				err := c.Create(context.Background(), apiRule)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.hosts[0]: Invalid value: \"string\": Host must be a lowercase RFC 1123 label (must consist of lowercase alphanumeric characters or '-', and must start and end with an lowercase alphanumeric character), a fully qualified domain name, or a wildcard domain name"))
			}

			It("should not create an APIRule with an empty host", func() {
				invalidHelper("")
			})

			It("should not create an APIRule when host name has uppercase letters", func() {
				invalidHelper("eXample.com")
				invalidHelper("example.cOm")
			})

			It("should not create an APIRule with host label longer than 63 characters", func() {
				invalidHelper(gatewayv2alpha1.Host(strings.Repeat("a", 64) + ".com"))
				invalidHelper(gatewayv2alpha1.Host("example." + strings.Repeat("a", 64)))
			})

			It("should not create an APIRule when any domain label is empty", func() {
				invalidHelper(".com")
				invalidHelper("example..com")
				invalidHelper("example.")
			})

			It("should not create an APIRule when top level domain is too short", func() {
				invalidHelper("example.c")
			})

			It("should not create an APIRule when host contains wrong characters", func() {
				invalidHelper("*example.com")
				invalidHelper("exam*ple.com")
				invalidHelper("example*.com")
				invalidHelper("example.*com")
				invalidHelper("example.co*m")
				invalidHelper("example.com*")
			})

			It("should not create an APIRule when host starts or ends with a hyphen", func() {
				invalidHelper("-example.com")
				invalidHelper("example-.com")
				invalidHelper("example.-com")
				invalidHelper("example.com-")
			})
		})

		Context("rule path validation respected", func() {
			It("should fail when path consists of a path and *", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := generateTestName(testServiceNameBase, testIDLength)
				serviceHost := gatewayv2alpha1.Host("example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img*", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				Expect(c.Create(context.Background(), svc)).Should(Succeed())

				// when
				err := c.Create(context.Background(), apiRule)

				// then
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("spec.rules[0].path: Invalid value: \"/img*\": spec.rules[0].path"))

			})

			It("should apply APIRule when path contains only /*", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := generateTestName(testServiceNameBase, testIDLength)
				serviceHost := gatewayv2alpha1.Host("example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/*", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				Expect(c.Create(context.Background(), svc)).Should(Succeed())

				// when then
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})

			It("should apply APIRule when path contains no *", func() {
				// given
				apiRuleName := generateTestName(testNameBase, testIDLength)
				serviceName := generateTestName(testServiceNameBase, testIDLength)
				serviceHost := gatewayv2alpha1.Host("example.com")
				serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

				rule := testRulev2alpha1("/img-new/1", []gatewayv2alpha1.HttpMethod{http.MethodGet})
				rule.NoAuth = ptr.To(true)
				apiRule := testApiRulev2alpha1(apiRuleName, testNamespace, serviceName, testNamespace, serviceHosts, testServicePort, []gatewayv2alpha1.Rule{rule})
				svc := testService(serviceName, testNamespace, testServicePort)
				defer func() {
					deleteResource(apiRule)
					deleteResource(svc)
				}()
				Expect(c.Create(context.Background(), svc)).Should(Succeed())

				// when then
				Expect(c.Create(context.Background(), apiRule)).Should(Succeed())

			})
		})
	})
})

func testRulev2alpha1(path string, methods []gatewayv2alpha1.HttpMethod) gatewayv2alpha1.Rule {
	return gatewayv2alpha1.Rule{
		Path:    path,
		Methods: methods,
	}
}

func testApiRulev2alpha1(name, namespace, serviceName, serviceNamespace string, serviceHosts []*gatewayv2alpha1.Host, servicePort uint32, rules []gatewayv2alpha1.Rule) *gatewayv2alpha1.APIRule {
	var gateway = testGatewayURL

	return &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.APIRuleSpec{
			Hosts:   serviceHosts,
			Gateway: &gateway,
			Service: &gatewayv2alpha1.Service{
				Name:      &serviceName,
				Namespace: &serviceNamespace,
				Port:      &servicePort,
			},
			Rules: rules,
		},
	}
}

func testApiRulev2alpha1Gateway(name, namespace, serviceName, serviceNamespace, gateway string, servicePort uint32, rules []gatewayv2alpha1.Rule) *gatewayv2alpha1.APIRule {
	serviceHost := gatewayv2alpha1.Host("example.com")
	serviceHosts := []*gatewayv2alpha1.Host{&serviceHost}

	return &gatewayv2alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.APIRuleSpec{
			Hosts:   serviceHosts,
			Gateway: &gateway,
			Service: &gatewayv2alpha1.Service{
				Name:      &serviceName,
				Namespace: &serviceNamespace,
				Port:      &servicePort,
			},
			Rules: rules,
		},
	}
}

func testService(name, namespace string, servicePort uint32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: int32(servicePort),
				},
			},
		},
	}
}

func generateTestName(name string, length int) string {
	rand.NewSource(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return name + "-" + string(b)
}

func matchingLabelsFunc(apiRuleName, namespace string) client.ListOption {
	labels := make(map[string]string)
	labels[processing.OwnerLabelName] = apiRuleName
	labels[processing.OwnerLabelNamespace] = namespace
	return client.MatchingLabels(labels)
}

func updateJwtHandlerTo(jwtHandler string) {
	cm := &corev1.ConfigMap{}
	Expect(c.Get(context.Background(), client.ObjectKey{Name: helpers.CM_NAME, Namespace: helpers.CM_NS}, cm)).Should(Succeed())

	if !strings.Contains(cm.Data[helpers.CM_KEY], jwtHandler) {
		By(fmt.Sprintf("Updating JWT handler config map to %s", jwtHandler))
		cm.Data = map[string]string{
			helpers.CM_KEY: fmt.Sprintf("jwtHandler: %s", jwtHandler),
		}
		Expect(c.Update(context.Background(), cm)).To(Succeed())

		By("Waiting until config map is updated")
		Eventually(func(g Gomega) {
			g.Expect(c.Get(context.Background(), client.ObjectKey{Name: cm.Name, Namespace: cm.Namespace}, cm)).Should(Succeed())
			g.Expect(cm.Data).To(HaveKeyWithValue(helpers.CM_KEY, fmt.Sprintf("jwtHandler: %s", jwtHandler)))
		}, eventuallyTimeout).Should(Succeed())
	}
}

func deleteResource(object client.Object) {
	By(fmt.Sprintf("Deleting resource %s as part of teardown", object.GetName()))
	err := c.Delete(context.Background(), object)

	if err != nil {
		Expect(errors.IsNotFound(err)).To(BeTrue())
	}

	Eventually(func(g Gomega) {
		err := c.Get(context.Background(), client.ObjectKeyFromObject(object), object)
		g.Expect(errors.IsNotFound(err)).To(BeTrue())
	}, eventuallyTimeout).Should(Succeed())
}
