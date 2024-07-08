package v2alpha1_test

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
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
	jwtConfigJSON := fmt.Sprintf(`{"authentications": [{"issuer": "%s", "jwksUri": "%s"}]}`, jwtIssuer, jwksUri)
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
	jwtV1beta1Rule := GetRuleFor(path, []gatewayv1beta1.HttpMethod{http.MethodGet}, []*gatewayv1beta1.Mutator{}, jwtV1beta1)

	jwtV2alpha1 := gatewayv2alpha1.Rule{
		Path:    path,
		Methods: []gatewayv2alpha1.HttpMethod{http.MethodGet},
		Jwt: &gatewayv2alpha1.JwtConfig{
			Authentications: []*gatewayv2alpha1.JwtAuthentication{
				{
					Issuer:  jwtIssuer,
					JwksUri: jwksUri,
				},
			},
		},
	}

	noAuthV1beta1 := []*gatewayv1beta1.Authenticator{
		{
			Handler: &gatewayv1beta1.Handler{
				Name: gatewayv1beta1.AccessStrategyNoAuth,
			},
		},
	}
	noAuthV1beta1Rule := GetRuleFor(path, []gatewayv1beta1.HttpMethod{http.MethodGet}, []*gatewayv1beta1.Mutator{}, noAuthV1beta1)

	noAuthV2alpha1Rule := gatewayv2alpha1.Rule{
		Path:    path,
		Methods: []gatewayv2alpha1.HttpMethod{http.MethodGet},
		NoAuth:  ptr.To(true),
	}

	It("with v1beta1 no_auth and v2alpha1 noAuth should provide only Istio VS", func() {
		// given

		rulesV1beta1 := []gatewayv1beta1.Rule{noAuthV1beta1Rule}
		v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

		rulesV2alpha1 := []gatewayv2alpha1.Rule{noAuthV2alpha1Rule}
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

	It("with v1beta1 jwt and v2alpha1 jwt should provide VirtualService, AuthorizationPolicy and RequestAuthentication", func() {
		// given

		rulesV1beta1 := []gatewayv1beta1.Rule{jwtV1beta1Rule}
		v1beta1ApiRule := GetAPIRuleFor(rulesV1beta1)

		rulesV2alpha1 := []gatewayv2alpha1.Rule{jwtV2alpha1}
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
})

func getV2alpha1APIRuleFor(name, namespace string, rules []gatewayv2alpha1.Rule) *gatewayv2alpha1.APIRule {

	serviceHost := gatewayv2alpha1.Host("myService.test.com")

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
