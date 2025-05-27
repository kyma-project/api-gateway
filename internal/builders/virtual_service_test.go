package builders

import (
	"fmt"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	v1beta12 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
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
						AddHost(host).
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
				AddHost(host).
				AddHost(host2).
				Gateway(gateway).
				Gateway(gateway2).
				HTTP(HTTPRoute().
					Match(MatchRequest().Uri().Regex(matchURIRegex)).
					Match(MatchRequest().Uri().Regex(matchURIRegex2)).
					Headers(NewHttpRouteHeadersBuilder().SetHostHeader(host).RemoveUpstreamCORSPolicyHeaders().Get()).
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
			Expect(result.Http[0].Headers.Response.Remove).To(ConsistOf([]string{
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
				AllowOrigins:     apirulev1beta1.StringMatch{{"exact": "localhost"}},
				AllowMethods:     []string{"GET", "POST"},
				AllowCredentials: ptr.To(true),
				AllowHeaders:     []string{"Allowed-Header"},
				ExposeHeaders:    []string{"Exposed-Header"},
				MaxAge:           ptr.To(metav1.Duration{Duration: time.Second}),
			}

			result := VirtualServiceSpec().
				AddHost(host).
				Gateway(gateway).
				HTTP(HTTPRoute().
					CorsPolicy(CorsPolicy().FromApiRuleCorsPolicy(corsPolicy)).
					Match(MatchRequest().Uri().Regex(matchURIRegex)).
					Headers(NewHttpRouteHeadersBuilder().SetHostHeader(host).RemoveUpstreamCORSPolicyHeaders().Get()).
					Route(RouteDestination().Host(destHost).Port(destPort))).Get()

			Expect(result.Hosts).To(HaveLen(1))
			Expect(result.Hosts[0]).To(Equal(host))

			Expect(result.Gateways).To(HaveLen(1))
			Expect(result.Gateways[0]).To(Equal(gateway))

			Expect(result.Http).To(HaveLen(1))

			Expect(result.Http[0].Match).To(HaveLen(1))
			Expect(result.Http[0].Match[0].Uri.GetRegex()).To(Equal(matchURIRegex))
			Expect(result.Http[0].Headers.Request.Set).To(Equal(map[string]string{"x-forwarded-host": host}))

			Expect(result.Http[0].Headers.Response.Remove).To(ConsistOf([]string{
				ExposeHeadersName,
				AllowHeadersName,
				AllowCredentialsName,
				AllowMethodsName,
				AllowOriginName,
				MaxAgeName,
			}))

			Expect(result.Http[0].CorsPolicy.AllowOrigins).To(HaveLen(1))
			Expect(result.Http[0].CorsPolicy.AllowOrigins).To(ConsistOf(&v1beta12.StringMatch{MatchType: &v1beta12.StringMatch_Exact{Exact: "localhost"}}))
			Expect(result.Http[0].CorsPolicy.AllowCredentials).To(Not(BeNil()))
			Expect(result.Http[0].CorsPolicy.AllowCredentials.Value).To(BeTrue())
			Expect(result.Http[0].CorsPolicy.AllowMethods).To(ConsistOf("GET", "POST"))
			Expect(result.Http[0].CorsPolicy.AllowHeaders).To(ConsistOf("Allowed-Header"))
			Expect(result.Http[0].CorsPolicy.ExposeHeaders).To(ConsistOf("Exposed-Header"))

			Expect(result.Http[0].Route).To(HaveLen(1))
			Expect(result.Http[0].Route[0].Destination.Host).To(Equal(destHost))
			Expect(result.Http[0].Route[0].Destination.Port.Number).To(Equal(destPort))
			Expect(result.Http[0].Route[0].Weight).To(Equal(int32(100)))
		})

		Describe("MethodRegEx", func() {

			expectHttpMethodRegex := func(regex *regexp.Regexp, httpMethod string) {
				Expect(regex.MatchString(httpMethod)).To(BeTrue())
				Expect(regex.MatchString(fmt.Sprintf("%sA", httpMethod))).To(BeFalse())
				Expect(regex.MatchString(fmt.Sprintf("A%s", httpMethod))).To(BeFalse())
				Expect(regex.MatchString(fmt.Sprintf("%s ", httpMethod))).To(BeFalse())
				Expect(regex.MatchString(fmt.Sprintf(" %s", httpMethod))).To(BeFalse())
			}

			It("should create regex only matching exact given method", func() {
				mr := MatchRequest().MethodRegEx("GET").Get()

				regex := regexp.MustCompile(mr.Method.GetRegex())

				expectHttpMethodRegex(regex, "GET")
				Expect(regex.MatchString("GUT")).To(BeFalse())
				Expect(regex.MatchString("PUT")).To(BeFalse())
			})

			It("should create regex matching exact multiple methods", func() {
				mr := MatchRequest().MethodRegEx("GET", "PUT", "POST", "PATCH").Get()

				regex := regexp.MustCompile(mr.Method.GetRegex())

				expectHttpMethodRegex(regex, "GET")
				Expect(regex.MatchString("GUT")).To(BeFalse())
				expectHttpMethodRegex(regex, "PUT")
				expectHttpMethodRegex(regex, "POST")
				expectHttpMethodRegex(regex, "PATCH")
				Expect(regex.MatchString("DELETE")).To(BeFalse())
			})
		})
	})
})
