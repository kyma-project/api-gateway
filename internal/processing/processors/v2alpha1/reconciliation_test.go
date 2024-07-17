package v2alpha1_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"istio.io/api/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"net/http"

	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	jwtIssuer = "https://example.com/"
	jwksUri   = "https://example.com/.well-known/jwks.json"
	path      = "/test"
)

var _ = Describe("Reconciliation", func() {

	Context("with one rule", func() {

		It("with v1beta1 no_auth and v2alpha1 noAuth should provide only Istio VS", func() {
			// given
			rulesV1beta1 := []gatewayv1beta1.Rule{getNoAuthV1beta1Rule(path)}
			v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

			rulesV2alpha1 := []gatewayv2alpha1.Rule{getNoAuthV2alpha1Rule(path)}
			v2alpha1ApiRule := getV2alpha1APIRuleFor("test-apirule", "some-namespace", rulesV2alpha1)

			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := v2alpha1.NewReconciliation(v2alpha1ApiRule, v1beta1ApiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(1))

			Expect(createdObjects[0]).To(BeAssignableToTypeOf(&networkingv1beta1.VirtualService{}))
		})

		It("with v1beta1 JWT and v2alpha1 JWT should provide VirtualService, AuthorizationPolicy and RequestAuthentication", func() {
			// given

			rulesV1beta1 := []gatewayv1beta1.Rule{getJwtV1beta1Rule(path, jwtIssuer, jwksUri)}
			v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

			rulesV2alpha1 := []gatewayv2alpha1.Rule{getJwtV2alpha1Rule(path, jwtIssuer, jwksUri)}
			v2alpha1ApiRule := getV2alpha1APIRuleFor("test-apirule", "some-namespace", rulesV2alpha1)

			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := v2alpha1.NewReconciliation(v2alpha1ApiRule, v1beta1ApiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(3))

			vsCreated, apCreated, raCreated := false, false, false

			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Http)).To(Equal(1))
					Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
					Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal("example-service.some-namespace.svc.cluster.local"))
				} else if raOk {
					raCreated = true
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(1))
					Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
					Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
				} else if apOk {
					apCreated = true
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
			Expect(vsCreated).To(BeTrue())
			Expect(apCreated).To(BeTrue())
			Expect(raCreated).To(BeTrue())
		})

		It("with two JWT Authentications in v1beta1 and v2alpha1 should provide VirtualService, AuthorizationPolicy and two RequestAuthentication", func() {
			// given
			// v1beta1 Rule
			jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, jwtIssuer, jwksUri)
			differentJwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, "https://different.com/", "https://different.com/.well-known/jwks.json")
			authenticatorsV1beta1 := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyJwt,
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyJwt,
						Config: &runtime.RawExtension{
							Raw: []byte(differentJwtConfigJSON),
						},
					},
				},
			}
			rulesV1beta1 := GetRuleFor(path, []gatewayv1beta1.HttpMethod{http.MethodGet}, []*gatewayv1beta1.Mutator{}, authenticatorsV1beta1)
			v1beta1ApiRule := GetAPIRuleFor([]gatewayv1beta1.Rule{rulesV1beta1})

			// v2alpha1 Rule
			rulesV2alpha1 := gatewayv2alpha1.Rule{
				Path:    path,
				Methods: []gatewayv2alpha1.HttpMethod{http.MethodGet},
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  jwtIssuer,
							JwksUri: jwksUri,
						},
						{
							Issuer:  "https://different.com/",
							JwksUri: "https://different.com/.well-known/jwks.json",
						},
					},
				},
			}
			v2alpha1ApiRule := getV2alpha1APIRuleFor("test-apirule", "some-namespace", []gatewayv2alpha1.Rule{rulesV2alpha1})

			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := v2alpha1.NewReconciliation(v2alpha1ApiRule, v1beta1ApiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(3))

			vsCreated, raCreated, apCreated := false, false, false

			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Http)).To(Equal(1))
					Expect(len(vs.Spec.Http[0].Route)).To(Equal(1))
					Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal("example-service.some-namespace.svc.cluster.local"))
				} else if raOk {
					raCreated = true
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(2))

					for _, jwtRule := range ra.Spec.JwtRules {
						switch jwtRule.Issuer {
						case jwtIssuer:
							Expect(jwtRule.JwksUri).To(Equal(jwksUri))
						case "https://different.com/":
							Expect(jwtRule.JwksUri).To(Equal("https://different.com/.well-known/jwks.json"))
						default:
							Fail("Unexpected issuer")
						}
					}

				} else if apOk {
					apCreated = true
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
			Expect(vsCreated).To(BeTrue())
			Expect(apCreated).To(BeTrue())
			Expect(raCreated).To(BeTrue())
		})
	})

	Context("with two rules", func() {
		It("with two JWT rules in v1beta1 and v2alpha1 should provide VirtualService, AuthorizationPolicy and two RequestAuthentication", func() {
			// given

			rulesV1beta1 := []gatewayv1beta1.Rule{getJwtV1beta1Rule(path, jwtIssuer, jwksUri), getJwtV1beta1Rule("/different-path", "https://different.com/", "https://different.com/.well-known/jwks.json")}
			v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

			rulesV2alpha1 := []gatewayv2alpha1.Rule{getJwtV2alpha1Rule(path, jwtIssuer, jwksUri), getJwtV2alpha1Rule("/different-path", "https://different.com/", "https://different.com/.well-known/jwks.json")}
			v2alpha1ApiRule := getV2alpha1APIRuleFor("test-apirule", "some-namespace", rulesV2alpha1)

			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := v2alpha1.NewReconciliation(v2alpha1ApiRule, v1beta1ApiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(5))

			vsCreated, apCreated := false, false
			numberOfCreatedRequestAuthentications := 0

			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					for _, http := range vs.Spec.Http {
						Expect(len(http.Route)).To(Equal(1))
						Expect(http.Route[0].Destination.Host).To(Equal("example-service.some-namespace.svc.cluster.local"))

						switch http.Match[0].Uri.MatchType.(*v1beta1.StringMatch_Regex).Regex {
						case "/test":
							break
						case "/different-path":
							break
						default:
							Fail("Unexpected match type")
						}
					}
				} else if raOk {
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(1))

					switch ra.Spec.JwtRules[0].Issuer {
					case jwtIssuer:
						Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
					case "https://different.com/":
						Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal("https://different.com/.well-known/jwks.json"))
					default:
						Fail("Unexpected issuer")

					}
					numberOfCreatedRequestAuthentications++

				} else if apOk {
					apCreated = true
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
			Expect(vsCreated).To(BeTrue())
			Expect(apCreated).To(BeTrue())
			Expect(numberOfCreatedRequestAuthentications).To(Equal(2))
		})

		It("with no_auth rule and JWT rule in v1beta1 and noAuth rule and JWT rule v2alpha1 should provide VirtualService, AuthorizationPolicy and RequestAuthentication", func() {
			// given

			rulesV1beta1 := []gatewayv1beta1.Rule{getJwtV1beta1Rule(path, jwtIssuer, jwksUri), getNoAuthV1beta1Rule("/different-path")}
			v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

			rulesV2alpha1 := []gatewayv2alpha1.Rule{getJwtV2alpha1Rule(path, jwtIssuer, jwksUri), getNoAuthV2alpha1Rule("/different-path")}
			v2alpha1ApiRule := getV2alpha1APIRuleFor("test-apirule", "some-namespace", rulesV2alpha1)

			service := GetService(ServiceName)
			fakeClient := GetFakeClient(service)

			// when
			var createdObjects []client.Object
			reconciliation := v2alpha1.NewReconciliation(v2alpha1ApiRule, v1beta1ApiRule, GetTestConfig(), &testLogger)
			for _, processor := range reconciliation.GetProcessors() {
				results, err := processor.EvaluateReconciliation(context.Background(), fakeClient)
				Expect(err).To(BeNil())
				for _, result := range results {
					createdObjects = append(createdObjects, result.Obj)
				}
			}

			// then
			Expect(createdObjects).To(HaveLen(4))

			vsCreated, raCreated, apCreated := false, false, false

			for _, createdObj := range createdObjects {
				vs, vsOk := createdObj.(*networkingv1beta1.VirtualService)
				ra, raOk := createdObj.(*securityv1beta1.RequestAuthentication)
				ap, apOk := createdObj.(*securityv1beta1.AuthorizationPolicy)

				if vsOk {
					vsCreated = true
					Expect(vs).NotTo(BeNil())
					Expect(len(vs.Spec.Http)).To(Equal(2))
					for _, http := range vs.Spec.Http {
						Expect(len(http.Route)).To(Equal(1))
						Expect(http.Route[0].Destination.Host).To(Equal("example-service.some-namespace.svc.cluster.local"))

						switch http.Match[0].Uri.MatchType.(*v1beta1.StringMatch_Regex).Regex {
						case "/test":
							break
						case "/different-path":
							break
						default:
							Fail("Unexpected match type")
						}
					}
				} else if raOk {
					raCreated = true
					Expect(ra).NotTo(BeNil())
					Expect(len(ra.Spec.JwtRules)).To(Equal(1))
					Expect(ra.Spec.JwtRules[0].Issuer).To(Equal(jwtIssuer))
					Expect(ra.Spec.JwtRules[0].JwksUri).To(Equal(jwksUri))
				} else if apOk {
					apCreated = true
					Expect(ap).NotTo(BeNil())
					Expect(len(ap.Spec.Rules)).To(Equal(1))
				}
			}
			Expect(vsCreated).To(BeTrue())
			Expect(apCreated).To(BeTrue())
			Expect(raCreated).To(BeTrue())
		})
	})

	It("v2alpha1 reconciliation calls v2alpha1 api rule validation", func() {
		// given
		rulesV1beta1 := []gatewayv1beta1.Rule{getNoAuthV1beta1Rule(path)}
		v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

		rulesV2alpha1 := []gatewayv2alpha1.Rule{getNoAuthV2alpha1Rule(path)}
		v2alpha1ApiRule := getV2alpha1APIRuleFor("test-apirule", "some-namespace", rulesV2alpha1)

		brokenHost := gatewayv2alpha1.Host("host-without-domain")
		v2alpha1ApiRule.Spec.Hosts[0] = &brokenHost

		service := GetService(ServiceName)
		fakeClient := GetFakeClient(service)

		// when
		reconciliation := v2alpha1.NewReconciliation(v2alpha1ApiRule, v1beta1ApiRule, GetTestConfig(), &testLogger)
		problems, _ := reconciliation.Validate(context.Background(), fakeClient)

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.hosts[0]"))
		Expect(problems[0].Message).To(Equal("Host is not fully qualified domain name"))
	})
})

func getV2alpha1APIRuleFor(name, namespace string, rules []gatewayv2alpha1.Rule) *gatewayv2alpha1.APIRule {

	serviceHost := gatewayv2alpha1.Host("myservice.test.com")

	return &gatewayv2alpha1.APIRule{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.APIRuleSpec{
			Gateway: ptr.To("some-gateway"),
			Service: &gatewayv2alpha1.Service{
				Name: ptr.To("example-service"),
				Port: ptr.To(uint32(8080)),
			},
			Hosts: []*gatewayv2alpha1.Host{&serviceHost},
			Rules: rules,
		},
	}

}

func getJwtV1beta1Rule(path, issuer, jwksUri string) gatewayv1beta1.Rule {
	jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, issuer, jwksUri)
	jwtV1beta1 := []*gatewayv1beta1.Authenticator{
		{
			Handler: &gatewayv1beta1.Handler{
				Name: gatewayv1beta1.AccessStrategyJwt,
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		},
	}
	return GetRuleFor(path, []gatewayv1beta1.HttpMethod{http.MethodGet}, []*gatewayv1beta1.Mutator{}, jwtV1beta1)
}

func getNoAuthV1beta1Rule(path string) gatewayv1beta1.Rule {

	noAuthV1beta1 := []*gatewayv1beta1.Authenticator{
		{
			Handler: &gatewayv1beta1.Handler{
				Name: gatewayv1beta1.AccessStrategyNoAuth,
			},
		},
	}
	return GetRuleFor(path, []gatewayv1beta1.HttpMethod{http.MethodGet}, []*gatewayv1beta1.Mutator{}, noAuthV1beta1)
}

func getJwtV2alpha1Rule(path, issuer, jwksUri string) gatewayv2alpha1.Rule {
	return gatewayv2alpha1.Rule{
		Path:    path,
		Methods: []gatewayv2alpha1.HttpMethod{http.MethodGet},
		Jwt: &gatewayv2alpha1.JwtConfig{
			Authentications: []*gatewayv2alpha1.JwtAuthentication{
				{
					Issuer:  issuer,
					JwksUri: jwksUri,
				},
			},
		},
	}
}

func getNoAuthV2alpha1Rule(path string) gatewayv2alpha1.Rule {
	return gatewayv2alpha1.Rule{
		Path:    path,
		Methods: []gatewayv2alpha1.HttpMethod{http.MethodGet},
		NoAuth:  ptr.To(true),
	}
}
