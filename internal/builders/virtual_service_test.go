package builders

import (
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

var _ = Describe("Builder for", func() {

	host := "oauthkeeper.cluster.local"
	gateway := "some-gateway"

	matchURIRegex := "*/"
	destHost := "somehost.somenamespace.svc.cluster.local"
	var destPort uint32 = 4321

	Describe("VirtualService", func() {
		It("should build the object", func() {
			name := "testName"
			namespace := "testNs"

			initialVs := networkingv1beta1.VirtualService{}
			initialVs.Name = "shoudBeOverwritten"
			initialVs.Spec.Hosts = []string{"a,", "b", "c"}

			vs := VirtualService().From(&initialVs).GenerateName(name).Namespace(namespace).
				Spec(
					VirtualServiceSpec().
						Host(host).
						Gateway(gateway)).
				Get()
			Expect(vs.Name).To(BeEmpty())
			Expect(vs.GenerateName).To(Equal(name))
			Expect(vs.Namespace).To(Equal(namespace))
			Expect(vs.Spec.Hosts).To(HaveLen(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(host))
			Expect(vs.Spec.Gateways).To(HaveLen(1))
			Expect(vs.Spec.Gateways[0]).To(Equal(gateway))
		})
	})

	Describe("VirtualService Spec", func() {
		It("should build the spec", func() {

			host2 := "oauthkeeper2.cluster.local"
			gateway2 := "some-gateway-2"
			matchURIRegex2 := "/abcd"
			matchURIRegex3 := "/xyz/foobar"
			destHost2 := "somehost2.somenamespace.svc.cluster.local"
			var destPort2 uint32 = 4343
			timeout := time.Second * 60

			result := VirtualServiceSpec().
				Host(host).
				Host(host2).
				Gateway(gateway).
				Gateway(gateway2).
				HTTP(HTTPRoute().
					Match(MatchRequest().Uri().Regex(matchURIRegex)).
					Match(MatchRequest().Uri().Regex(matchURIRegex2)).
					Headers(NewHttpRouteHeadersBuilder().SetHostHeader(host).SetCORSPolicyHeaders(apirulev1beta1.CorsPolicy{}).Get()).
					Route(RouteDestination().Host(destHost).Port(destPort)).
					Timeout(timeout)).
				HTTP(HTTPRoute().
					Match(MatchRequest().Uri().Regex(matchURIRegex3)).
					Route(RouteDestination().Host(destHost2).Port(destPort2)).
					Timeout(time.Second * 180)).
				Get()

			Expect(result.Hosts).To(HaveLen(2))
			Expect(result.Hosts[0]).To(Equal(host))
			Expect(result.Hosts[1]).To(Equal(host2))

			Expect(result.Gateways).To(HaveLen(2))
			Expect(result.Gateways[0]).To(Equal(gateway))
			Expect(result.Gateways[1]).To(Equal(gateway2))

			//Two HTTPRoute elements
			Expect(result.Http).To(HaveLen(2))

			//Two HTTPMatchRequest elements
			Expect(result.Http[0].Match).To(HaveLen(2))
			Expect(result.Http[0].Match[0].Uri.GetRegex()).To(Equal(matchURIRegex))
			Expect(result.Http[0].Match[1].Uri.GetRegex()).To(Equal(matchURIRegex2))
			Expect(result.Http[0].Headers.Request.Set).To(Equal(map[string]string{"x-forwarded-host": host}))

			//Expect host headers set to remove
			Expect(result.Http[0].Headers.Response.Remove).To(ContainElements([]string{
				ExposeHeadersName,
				AllowHeadersName,
				AllowCredentialsName,
				AllowMethodsName,
				AllowOriginName,
				MaxAgeName,
			}))
			Expect(result.Http[0].Route).To(HaveLen(1))
			Expect(result.Http[0].Route[0].Destination.Host).To(Equal(destHost))
			Expect(result.Http[0].Route[0].Destination.Port.Number).To(Equal(destPort))
			//Expect(result.Http[0].Route[0].Destination.Port.Name).To(BeEmpty())
			Expect(result.Http[0].Route[0].Weight).To(Equal(int32(100)))

			Expect(result.Http[0].Timeout).To(Equal(durationpb.New(timeout)))
			Expect(result.Http[1].Timeout).To(Equal(durationpb.New(time.Second * 180)))

			//One HTTPMatchRequest element
			Expect(result.Http[1].Match).To(HaveLen(1))
			Expect(result.Http[1].Match[0].Uri.GetRegex()).To(Equal(matchURIRegex3))
			Expect(result.Http[1].Route).To(HaveLen(1))
			Expect(result.Http[1].Route[0].Destination.Host).To(Equal(destHost2))
			Expect(result.Http[1].Route[0].Destination.Port.Number).To(Equal(destPort2))
			//Expect(result.Http[1].Route[0].Destination.Port.Name).To(BeEmpty())
			Expect(result.Http[1].Route[0].Weight).To(Equal(int32(100)))
		})

		It("should build the CORS headers", func() {
			corsPolicy := apirulev1beta1.CorsPolicy{
				AllowOrigins:     apirulev1beta1.StringMatch{"exact": "localhost"},
				AllowMethods:     []string{"GET", "POST"},
				AllowCredentials: ptr.To(true),
				AllowHeaders:     []string{"test"},
				ExposeHeaders:    []string{"test"},
				MaxAge:           ptr.To(metav1.Duration{Duration: time.Second}),
			}

			result := VirtualServiceSpec().
				Host(host).
				Gateway(gateway).
				HTTP(HTTPRoute().
					Match(MatchRequest().Uri().Regex(matchURIRegex)).
					Headers(NewHttpRouteHeadersBuilder().SetHostHeader(host).SetCORSPolicyHeaders(corsPolicy).Get()).
					Route(RouteDestination().Host(destHost).Port(destPort))).Get()

			Expect(result.Hosts).To(HaveLen(1))
			Expect(result.Hosts[0]).To(Equal(host))

			Expect(result.Gateways).To(HaveLen(1))
			Expect(result.Gateways[0]).To(Equal(gateway))

			Expect(result.Http).To(HaveLen(1))

			Expect(result.Http[0].Match).To(HaveLen(1))
			Expect(result.Http[0].Match[0].Uri.GetRegex()).To(Equal(matchURIRegex))
			Expect(result.Http[0].Headers.Request.Set).To(Equal(map[string]string{"x-forwarded-host": host}))

			Expect(result.Http[0].Headers.Response.Remove).To(Not(ContainElements([]string{
				ExposeHeadersName,
				AllowHeadersName,
				AllowCredentialsName,
				AllowMethodsName,
				AllowOriginName,
				MaxAgeName,
			})))

			Expect(result.Http[0].Headers.Response.Set).To(HaveKeyWithValue(ExposeHeadersName, "test"))
			Expect(result.Http[0].Headers.Response.Set).To(HaveKeyWithValue(AllowHeadersName, "test"))
			Expect(result.Http[0].Headers.Response.Set).To(HaveKeyWithValue(AllowCredentialsName, "true"))
			Expect(result.Http[0].Headers.Response.Set).To(HaveKeyWithValue(AllowMethodsName, "GET,POST"))
			Expect(result.Http[0].Headers.Response.Set).To(HaveKeyWithValue(MaxAgeName, "1"))

			Expect(result.Http[0].Route).To(HaveLen(1))
			Expect(result.Http[0].Route[0].Destination.Host).To(Equal(destHost))
			Expect(result.Http[0].Route[0].Destination.Port.Number).To(Equal(destPort))
			Expect(result.Http[0].Route[0].Weight).To(Equal(int32(100)))
		})
	})
})
