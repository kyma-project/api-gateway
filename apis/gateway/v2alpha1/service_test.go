package v2alpha1_test

import (
	"context"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Service", func() {

	DescribeTable("FindServiceNamespace",
		func(apiRule *v2alpha1.APIRule, rule v2alpha1.Rule, expectedValue string, expectError bool) {
			namespace, err := v2alpha1.FindServiceNamespace(apiRule, rule)
			if expectError {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(expectedValue))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(namespace).To(Equal(expectedValue))
			}
		},
		Entry("should return rule service namespace when it's set",
			&v2alpha1.APIRule{},
			v2alpha1.Rule{
				Service: &v2alpha1.Service{
					Namespace: ptr.To("test-namespace"),
				},
			},
			"test-namespace",
			false,
		),
		Entry("should return apiRule service namespace when rule service namespace is not set",
			&v2alpha1.APIRule{
				Spec: v2alpha1.APIRuleSpec{
					Service: &v2alpha1.Service{
						Namespace: ptr.To("apiRule-service-namespace"),
					},
				},
			},
			v2alpha1.Rule{},
			"apiRule-service-namespace",
			false,
		),
		Entry("should return apiRule namespace when neither rule nor apiRule service namespace is set",
			&v2alpha1.APIRule{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "apiRule-namespace",
				},
			},
			v2alpha1.Rule{},
			"apiRule-namespace",
			false,
		),
		Entry("should return error when neither rule nor apiRule service namespace is set and apiRule is nil",
			nil,
			v2alpha1.Rule{},
			"apiRule is nil",
			true,
		),
	)

	Context("GetSelectorFromService", func() {

		It("should return error when service is not set", func() {
			_, err := v2alpha1.GetSelectorFromService(context.Background(), createFakeClient(), &v2alpha1.APIRule{}, v2alpha1.Rule{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service name is required but missing"))
		})

		It("should return error when service name is not set", func() {
			_, err := v2alpha1.GetSelectorFromService(context.Background(), createFakeClient(), &v2alpha1.APIRule{}, v2alpha1.Rule{
				Service: &v2alpha1.Service{},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("service name is required but missing"))
		})

		It("should return error when service is not found", func() {
			_, err := v2alpha1.GetSelectorFromService(context.Background(), createFakeClient(), &v2alpha1.APIRule{}, v2alpha1.Rule{
				Service: &v2alpha1.Service{
					Name: ptr.To("test-service"),
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("services \"test-service\" not found"))
		})

		It("should return empty selector when service has no selector", func() {
			s := newServiceBuilder().
				withName("test-service").
				withNamespace("test-namespace").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient, &v2alpha1.APIRule{}, v2alpha1.Rule{
				Service: &v2alpha1.Service{
					Name:      ptr.To("test-service"),
					Namespace: ptr.To("test-namespace"),
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector).To(Equal(v2alpha1.PodSelector{}))
		})

		It("should return selector from service", func() {
			s := newServiceBuilder().
				withName("test-service").
				withNamespace("test-namespace").
				addSelector("app", "test-service").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient, &v2alpha1.APIRule{}, v2alpha1.Rule{
				Service: &v2alpha1.Service{
					Name:      ptr.To("test-service"),
					Namespace: ptr.To("test-namespace"),
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector.Selector.MatchLabels).To(Equal(map[string]string{"app": "test-service"}))
			Expect(selector.Namespace).To(Equal("test-namespace"))
		})

		It("should use service from Rule when APIRule Spec has Service set", func() {
			s := newServiceBuilder().
				withName("test-rule-service").
				withNamespace("test-rule-namespace").
				addSelector("app", "test-service").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient,
				&v2alpha1.APIRule{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "apirule-namespace",
					},
					Spec: v2alpha1.APIRuleSpec{
						Service: &v2alpha1.Service{
							Name:      ptr.To("test-spec-service"),
							Namespace: ptr.To("test-spec-namespace"),
						},
					},
				}, v2alpha1.Rule{
					Service: &v2alpha1.Service{
						Name:      ptr.To("test-rule-service"),
						Namespace: ptr.To("test-rule-namespace"),
					},
				})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector.Selector.MatchLabels).To(Equal(map[string]string{"app": "test-service"}))
			Expect(selector.Namespace).To(Equal("test-rule-namespace"))
		})

		It("should use service from APIRule Spec when rule has no service set", func() {
			s := newServiceBuilder().
				withName("test-service").
				withNamespace("test-namespace").
				addSelector("app", "test-service").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient,
				&v2alpha1.APIRule{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "apirule-namespace",
					},
					Spec: v2alpha1.APIRuleSpec{
						Service: &v2alpha1.Service{
							Name:      ptr.To("test-service"),
							Namespace: ptr.To("test-namespace"),
						},
					},
				}, v2alpha1.Rule{})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector.Selector.MatchLabels).To(Equal(map[string]string{"app": "test-service"}))
			Expect(selector.Namespace).To(Equal("test-namespace"))
		})

		It("should use default as service namespace when namespace is empty", func() {
			s := newServiceBuilder().
				withName("test-service").
				withNamespace("default").
				addSelector("app", "test-service").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient,
				&v2alpha1.APIRule{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "apirule-namespace",
					},
					Spec: v2alpha1.APIRuleSpec{
						Service: &v2alpha1.Service{
							Name:      ptr.To("test-service"),
							Namespace: ptr.To(""),
						},
					},
				}, v2alpha1.Rule{})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector.Selector.MatchLabels).To(Equal(map[string]string{"app": "test-service"}))
			Expect(selector.Namespace).To(Equal("default"))
		})

		It("should use service namespace from APIRule Spec when Rule Service has name, but no namespace set", func() {
			s := newServiceBuilder().
				withName("test-rule-service").
				withNamespace("test-spec-namespace").
				addSelector("app", "test-service").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient,
				&v2alpha1.APIRule{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "apirule-namespace",
					},
					Spec: v2alpha1.APIRuleSpec{
						Service: &v2alpha1.Service{
							Name:      ptr.To("test-spec-service"),
							Namespace: ptr.To("test-spec-namespace"),
						},
					},
				}, v2alpha1.Rule{
					Service: &v2alpha1.Service{
						Name: ptr.To("test-rule-service"),
					},
				})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector.Selector.MatchLabels).To(Equal(map[string]string{"app": "test-service"}))
			Expect(selector.Namespace).To(Equal("test-spec-namespace"))
		})

		It("should use service namespace from APIRule when Rule Service and Spec Service have no namespace set", func() {
			s := newServiceBuilder().
				withName("test-rule-service").
				withNamespace("apirule-namespace").
				addSelector("app", "test-service").
				build()

			fakeClient := createFakeClient(s)
			selector, err := v2alpha1.GetSelectorFromService(context.Background(), fakeClient,
				&v2alpha1.APIRule{
					ObjectMeta: v1.ObjectMeta{
						Namespace: "apirule-namespace",
					},
					Spec: v2alpha1.APIRuleSpec{
						Service: &v2alpha1.Service{
							Name: ptr.To("test-spec-service"),
						},
					},
				}, v2alpha1.Rule{
					Service: &v2alpha1.Service{
						Name: ptr.To("test-rule-service"),
					},
				})
			Expect(err).NotTo(HaveOccurred())
			Expect(selector.Selector.MatchLabels).To(Equal(map[string]string{"app": "test-service"}))
			Expect(selector.Namespace).To(Equal("apirule-namespace"))
		})
	})
})
