package builders

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
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

			refName := "refName"
			refVersion := "v1alpha1"
			refKind := "APIRule"
			var refUID k8sTypes.UID = "123"

			initialVs := networkingv1alpha3.VirtualService{}
			initialVs.Name = "shoudBeOverwritten"
			initialVs.Spec.Hosts = []string{"a,", "b", "c"}

			vs := VirtualService().From(&initialVs).GenerateName(name).Namespace(namespace).
				Owner(OwnerReference().Name(refName).APIVersion(refVersion).Kind(refKind).UID(refUID).Controller(true)).
				Spec(
					VirtualServiceSpec().
						Host(host).
						Gateway(gateway)).
				Get()
			Expect(vs.Name).To(BeEmpty())
			Expect(vs.GenerateName).To(Equal(name))
			Expect(vs.Namespace).To(Equal(namespace))
			Expect(vs.OwnerReferences).To(HaveLen(1))
			Expect(vs.OwnerReferences[0].Name).To(Equal(refName))
			Expect(vs.OwnerReferences[0].APIVersion).To(Equal(refVersion))
			Expect(vs.OwnerReferences[0].Kind).To(Equal(refKind))
			Expect(vs.OwnerReferences[0].UID).To(BeEquivalentTo(refUID))
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

			result := VirtualServiceSpec().
				Host(host).
				Host(host2).
				Gateway(gateway).
				Gateway(gateway2).
				HTTP(HTTPRoute().
					Match(MatchRequest().URI().Regex(matchURIRegex)).
					Match(MatchRequest().URI().Regex(matchURIRegex2)).
					Route(RouteDestination().Host(destHost).Port(destPort))).
				HTTP(HTTPRoute().
					Match(MatchRequest().URI().Regex(matchURIRegex3)).
					Route(RouteDestination().Host(destHost2).Port(destPort2))).
				Get()

			Expect(result.Hosts).To(HaveLen(2))
			Expect(result.Hosts[0]).To(Equal(host))
			Expect(result.Hosts[1]).To(Equal(host2))

			Expect(result.Gateways).To(HaveLen(2))
			Expect(result.Gateways[0]).To(Equal(gateway))
			Expect(result.Gateways[1]).To(Equal(gateway2))

			//Two HTTPRoute elements
			Expect(result.HTTP).To(HaveLen(2))

			//Two HTTPMatchRequest elements
			Expect(result.HTTP[0].Match).To(HaveLen(2))
			Expect(result.HTTP[0].Match[0].URI.Regex).To(Equal(matchURIRegex))
			Expect(result.HTTP[0].Match[1].URI.Regex).To(Equal(matchURIRegex2))
			Expect(result.HTTP[0].Route).To(HaveLen(1))
			Expect(result.HTTP[0].Route[0].Destination.Host).To(Equal(destHost))
			Expect(result.HTTP[0].Route[0].Destination.Port.Number).To(Equal(destPort))
			Expect(result.HTTP[0].Route[0].Destination.Port.Name).To(BeEmpty())

			//One HTTPMatchRequest element
			Expect(result.HTTP[1].Match).To(HaveLen(1))
			Expect(result.HTTP[1].Match[0].URI.Regex).To(Equal(matchURIRegex3))
			Expect(result.HTTP[1].Route).To(HaveLen(1))
			Expect(result.HTTP[1].Route[0].Destination.Host).To(Equal(destHost2))
			Expect(result.HTTP[1].Route[0].Destination.Port.Number).To(Equal(destPort2))
			Expect(result.HTTP[1].Route[0].Destination.Port.Name).To(BeEmpty())
		})
	})
})
