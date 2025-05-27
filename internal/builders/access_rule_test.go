package builders

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	internalTypes "github.com/kyma-project/api-gateway/internal/types/ory"
)

var _ = Describe("Builder for", func() {

	host := "oauthkeeper.cluster.local"
	hostPath := "/.*"
	destHost := "somehost.somenamespace.svc.cluster.local"
	var destPort uint32 = 4321
	methods := []v1beta1.HttpMethod{http.MethodGet, http.MethodPost, http.MethodPut}
	testScopes := []string{"read", "write"}

	Describe("AccessRule", func() {
		It("should build the object", func() {
			name := "testName"
			namespace := "testNs"

			testUpstreamURL := fmt.Sprintf("http://%s:%d", destHost, destPort)
			testMatchURL := fmt.Sprintf("<http|https>://%s<%s>", host, hostPath)

			requiredScopes := &internalTypes.OauthIntrospectionConfig{
				RequiredScope: testScopes,
			}
			requiredScopesJSON, _ := json.Marshal(requiredScopes)

			rawConfig := &runtime.RawExtension{
				Raw: requiredScopesJSON,
			}

			ar := AccessRule().GenerateName(name).Namespace(namespace).
				Spec(AccessRuleSpec().
					Upstream(Upstream().
						URL(testUpstreamURL)).
					Match(Match().
						URL(testMatchURL).
						Methods(methods)).
					Authorizer(Authorizer().
						Handler(Handler().
							Name("allow"))).
					Authenticators(Authenticators().
						Handler(Handler().
							Name("oauth2_introspection").
							Config(rawConfig)).
						Handler(Handler().
							Name("jwt"))).
					Mutators(Mutators().
						Handler(Handler().
							Name("hydrator")))).
				Get()
			Expect(ar.Name).To(BeEmpty())
			Expect(ar.GenerateName).To(Equal(name))
			Expect(ar.Namespace).To(Equal(namespace))
			Expect(ar.Spec.Upstream.URL).To(Equal(testUpstreamURL))
			Expect(ar.Spec.Match.URL).To(Equal(testMatchURL))
			Expect(ar.Spec.Match.Methods).To(BeEquivalentTo([]string{http.MethodGet, http.MethodPost, http.MethodPut}))
			Expect(ar.Spec.Authorizer.Handler.Name).To(Equal("allow"))
			Expect(len(ar.Spec.Authenticators)).To(Equal(2))
			Expect(ar.Spec.Authenticators[0].Name).To(Equal("oauth2_introspection"))
			Expect(ar.Spec.Authenticators[0].Config).To(Equal(rawConfig))
			Expect(ar.Spec.Authenticators[1].Name).To(Equal("jwt"))
			Expect(len(ar.Spec.Mutators)).To(Equal(1))
			Expect(ar.Spec.Mutators[0].Name).To(Equal("hydrator"))
		})
	})
})
