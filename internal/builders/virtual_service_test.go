package builders

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sTypes "k8s.io/apimachinery/pkg/types"
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

			vs := VirtualService().Name(name).Namespace(namespace).
				Owner(OwnerReference().Name(refName).APIVersion(refVersion).Kind(refKind).UID(refUID).Controller(true)).
				Spec(
					VirtualServiceSpec().
						Host(host).
						Gateway(gateway)).
				Get()
			fmt.Printf("%#v", vs)
			Expect(vs.Name).To(Equal(name))
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

			result := VirtualServiceSpec().
				Host(host).
				Gateway(gateway).
				HTTP(
					MatchRequest().URI().Regex(matchURIRegex),
					RouteDestination().Host(destHost).Port(destPort)).
				Get()

			Expect(result.Hosts).To(HaveLen(1))
			Expect(result.Hosts[0]).To(Equal(host))

			Expect(result.Gateways).To(HaveLen(1))
			Expect(result.Gateways[0]).To(Equal(gateway))

			Expect(result.HTTP).To(HaveLen(1))
			Expect(result.HTTP[0].Match).To(HaveLen(1))
			Expect(result.HTTP[0].Match[0].URI.Regex).To(Equal(matchURIRegex))
			Expect(result.HTTP[0].Route).To(HaveLen(1))
			Expect(result.HTTP[0].Route[0].Destination.Host).To(Equal(destHost))
			Expect(result.HTTP[0].Route[0].Destination.Port.Number).To(Equal(destPort))
			Expect(result.HTTP[0].Route[0].Destination.Port.Name).To(BeEmpty())
		})
	})
})
