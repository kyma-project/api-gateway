package v2alpha1

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

var _ = Describe("Sidecar injection validation", func() {
	It("should use service namespace and name from rule for pod selection when it's defined", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080), ptr.To("spec-level-ns")),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService("rule-level", uint32(8080), ptr.To("rule-level-ns")),
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		ruleService := getService("rule-level", "rule-level-ns")
		specService := getService("spec-level", "spec-level-ns")
		ruleLevelNs := getNamespace("rule-level-ns")
		specLevelNs := getNamespace("spec-level-ns")

		fakeClient := createFakeClient(ruleService, specService, ruleLevelNs, specLevelNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "rule-level-ns",
				Labels: map[string]string{
					"app": "rule-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod rule-level-ns/test does not have an injected istio sidecar"))
	})

	It("should use service namespace from spec and service name from rule for pod selection when no service namespace is defined on rule-level", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080), ptr.To("spec-level-ns")),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService("rule-level", uint32(8080)),
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		ruleService := getService("rule-level", "spec-level-ns")
		specService := getService("spec-level", "spec-level-ns")
		specLevelNs := getNamespace("spec-level-ns")
		fakeClient := createFakeClient(ruleService, specService, specLevelNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "spec-level-ns",
				Labels: map[string]string{
					"app": "rule-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod spec-level-ns/test does not have an injected istio sidecar"))
	})

	It("should use service namespace and name from spec for pod selection when no service is defined on rule-level", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080), ptr.To("spec-level-ns")),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		specService := getService("spec-level", "spec-level-ns")
		specLevelNs := getNamespace("spec-level-ns")
		fakeClient := createFakeClient(specService, specLevelNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "spec-level-ns",
				Labels: map[string]string{
					"app": "spec-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod spec-level-ns/test does not have an injected istio sidecar"))
	})

	It("should use service namespace from APIRule and service name from rule for pod selection when no service namespace is defined on spec or rule", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080)),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService("rule-level", uint32(8080)),
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		ruleService := getService("rule-level", "api-rule-ns")
		specService := getService("spec-level", "api-rule-ns")
		apiRuleNs := getNamespace("api-rule-ns")

		fakeClient := createFakeClient(ruleService, specService, apiRuleNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "api-rule-ns",
				Labels: map[string]string{
					"app": "rule-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod api-rule-ns/test does not have an injected istio sidecar"))
	})

	It("should use service namespace from APIRule and service name from spec for pod selection when no service namespace is defined on spec or rule", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080)),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		specService := getService("spec-level", "api-rule-ns")
		apiRuleNs := getNamespace("api-rule-ns")

		fakeClient := createFakeClient(specService, apiRuleNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "api-rule-ns",
				Labels: map[string]string{
					"app": "spec-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod api-rule-ns/test does not have an injected istio sidecar"))
	})

	It("should use 'default' as service namespace for pod selection and service name from rule when service namespace contains an empty string", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080), ptr.To("")),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService("rule-level", uint32(8080)),
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		ruleService := getService("rule-level", "default")
		specService := getService("spec-level", "default")
		ruleLevelNs := getNamespace("rule-level-ns")
		specLevelNs := getNamespace("spec-level-ns")
		defaultNs := getNamespace("default")
		fakeClient := createFakeClient(ruleService, specService, ruleLevelNs, specLevelNs, defaultNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels: map[string]string{
					"app": "rule-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod default/test does not have an injected istio sidecar"))
	})

	It("should use 'default' as service namespace for pod selection and service name from spec when service namespace contains an empty string", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080), ptr.To("")),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		specService := getService("spec-level", "default")
		specLevelNs := getNamespace("spec-level-ns")
		defaultNs := getNamespace("default")
		fakeClient := createFakeClient(specService, specLevelNs, defaultNs)

		err := fakeClient.Create(context.Background(), &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels: map[string]string{
					"app": "spec-level",
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Pod default/test does not have an injected istio sidecar"))
	})

	It("should return error when spec.service.name is not defined", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: &v2alpha1.Service{},
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		fakeClient := createFakeClient()

		//when
		_, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])

		//then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("service name is required but missing"))
	})

	It("should fail when the service has no selector", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: v1.ObjectMeta{
				Name:      "api-rule",
				Namespace: "api-rule-ns",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService("spec-level", uint32(8080), ptr.To("spec-level-ns")),
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: getHosts("test.dev"),
			},
		}

		specService := corev1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      "spec-level",
				Namespace: "spec-level-ns",
			},
		}

		specLevelNs := getNamespace("spec-level-ns")

		fakeClient := createFakeClient(&specService, specLevelNs)

		//when
		problems, err := validateSidecarInjection(context.Background(), fakeClient, "some.attribute", apiRule, apiRule.Spec.Rules[0])
		Expect(err).NotTo(HaveOccurred())

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Target service label selectors are not defined"))
	})
})
