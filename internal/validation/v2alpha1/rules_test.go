package v2alpha1

import (
	"context"
	"net/http"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Validate rules", func() {
	sampleServiceName := "some-service"
	host := v2alpha1.Host(sampleServiceName + ".test.dev")

	It("should fail for empty rules", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules:   nil,
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

	It("should return an error when no service is defined for rule with no service on spec level", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Hosts: []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						Path:   "/abc",
						NoAuth: ptr.To(true),
					},
				},
			},
		}

		service := getService(sampleServiceName, "default")
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].service"))
		Expect(problems[0].Message).To(Equal("The rule must define a service, because no service is defined on spec level"))
		Expect(problems[1].AttributePath).To(Equal(".spec.rules[0]"))
		Expect(problems[1].Message).To(Equal("Failed to execute sidecar injection validation, err: service name is required but missing"))
	})

	It("should report multiple problem at once", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "https://issuer.example.com",
									JwksUri: "https://issuer.test/.well-known/jwks.json",
								},
							},
						},
					},
					{
						Path:   "/abc",
						NoAuth: ptr.To(true),
					},
					{
						Path: "/test",
						Jwt:  &v2alpha1.JwtConfig{},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(2))
	})

	It("should fail for the same path and method", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodGet},
					},
					{
						Path:    "/abc/def",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodGet},
					},
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodGet, http.MethodPost},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("Path /abc with method GET conflicts with at least one of the other defined paths"))
	})

	DescribeTable("should fail for invalid path", func(path string, shouldFail bool, expectedMessage string) {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						NoAuth: ptr.To(true),
						Path:   path,
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		if shouldFail {
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].path"))
			Expect(problems[0].Message).To(Equal(expectedMessage))
		} else {
			Expect(problems).To(HaveLen(0))
		}

	},
		Entry("should not fail for valid path",
			"/test",
			false,
			""),
		Entry("should not fail for valid path with {**}",
			"/test/{**}",
			false,
			""),
		Entry(
			"should fail when operator {**} exists after {**}",
			"/{**}/{**}",
			true,
			"Only one {**} operator is allowed in the path."),
		Entry(
			"should fail when operator {**} exists after {**}",
			"/{**}/{**}/{**}",
			true,
			"Only one {**} operator is allowed in the path."),
		Entry(
			"should fail when operator {*} exists after {**}",
			"/{**}/test/{*}",
			true,
			"The {**} operator must be the last operator in the path."),
		Entry("should not fail when operator {**} exists after {*}",
			"/test/{*}/{**}",
			false,
			""),
	)

	It("should succeed for the same path but different methods", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				UID: "67890",
			},
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodGet},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should fail with service without labels selector when NoAuth is used", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].injection"))
		Expect(problems[0].Message).To(Equal("Target service label selectors are not defined"))
	})

	It("should fail with service on path level when NoAuth is used", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: []*v2alpha1.Host{&host},
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].injection"))
		Expect(problems[0].Message).To(Equal("Target service label selectors are not defined"))
	})

	It("should succeed with service without namespace", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should succeed with service on path level without namespace", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Rules: []v2alpha1.Rule{
					{
						Path:    "/abc",
						Service: getApiRuleService(sampleServiceName, uint32(8080)),
						NoAuth:  ptr.To(true),
						Methods: []v2alpha1.HttpMethod{http.MethodPost},
					},
				},
				Hosts: []*v2alpha1.Host{&host},
			},
		}

		service := getService(sampleServiceName)
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should invoke sidecar injection validation", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						Path: "/abc",
						Jwt: &v2alpha1.JwtConfig{
							Authentications: []*v2alpha1.JwtAuthentication{
								{
									Issuer:  "https://issuer.example.com",
									JwksUri: "https://issuer.test/.well-known/jwks.json",
								},
							},
						},
					},
				},
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].injection"))
		Expect(problems[0].Message).To(Equal("Target service label selectors are not defined"))
	})

	It("should invoke sidecar injection validation on noAuth expsed workload", func() {
		//given
		apiRule := &v2alpha1.APIRule{
			Spec: v2alpha1.APIRuleSpec{
				Service: getApiRuleService(sampleServiceName, uint32(8080)),
				Hosts:   []*v2alpha1.Host{&host},
				Rules: []v2alpha1.Rule{
					{
						Path:   "/abc",
						NoAuth: ptr.To(true),
					},
				},
			},
		}

		service := getService(sampleServiceName)
		service.Spec.Selector = map[string]string{}
		fakeClient := createFakeClient(service)

		//when
		problems := validateRules(context.Background(), fakeClient, ".spec", apiRule)

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].injection"))
		Expect(problems[0].Message).To(Equal("Target service label selectors are not defined"))
	})
})
