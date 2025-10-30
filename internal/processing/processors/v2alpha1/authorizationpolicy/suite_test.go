package authorizationpolicy_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	rulev1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/kyma-project/api-gateway/tests"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "authorizationpolicy v2alpha1 Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report ginkgotypes.Report) {
	tests.GenerateGinkgoJunitReport("authorizationpolicy-v2alpha1-suite", report)
})

var apiRuleName = "test-apirule"
var apiRuleNamespace = "example-namespace"
var serviceName = "example-service"
var testLogger = ctrl.Log.WithName("istio-test")
var testExpectedScopeKeys = []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}

var getAuthorizationPolicy = func(name string, namespace string, serviceName string, hosts, methods []string) *securityv1beta1.AuthorizationPolicy {
	ap := securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				processing.LegacyOwnerLabel: fmt.Sprintf("%s.%s", apiRuleName, apiRuleNamespace),
			},
		},
		Spec: v1beta1.AuthorizationPolicy{
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: map[string]string{
					"app": serviceName,
				},
			},
			Rules: []*v1beta1.Rule{
				{
					From: []*v1beta1.Rule_From{
						{
							Source: &v1beta1.Source{
								RequestPrincipals: []string{"*"},
							},
						},
					},
					To: []*v1beta1.Rule_To{
						{
							Operation: &v1beta1.Operation{
								Hosts:   hosts,
								Methods: methods,
								Paths:   []string{"/"},
							},
						},
					},
				},
			},
		},
	}

	apHash, err := hashbasedstate.GetAuthorizationPolicyHash(&ap)
	Expect(err).ShouldNot(HaveOccurred())
	ap.Labels["gateway.kyma-project.io/hash"] = apHash
	ap.Labels["gateway.kyma-project.io/index"] = "0"

	return &ap
}

var getActionMatcher = func(action string, namespace string, serviceName string, principalsName string, principals types.GomegaMatcher, methods types.GomegaMatcher, paths types.GomegaMatcher, notPaths types.GomegaMatcher) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Action": WithTransform(ActionToString, Equal(action)),
		"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
			"ObjectMeta": MatchFields(IgnoreExtras, Fields{
				"Namespace": Equal(namespace),
			}),
			"Spec": MatchFields(IgnoreExtras, Fields{
				"Selector": PointTo(MatchFields(IgnoreExtras, Fields{
					"MatchLabels": ContainElement(serviceName),
				})),
				"Rules": ContainElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"From": ContainElement(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Source": PointTo(MatchFields(IgnoreExtras, Fields{
									principalsName: principals,
								})),
							})),
						),
						"To": ContainElements(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Operation": PointTo(MatchFields(IgnoreExtras, Fields{
									"Methods":  methods,
									"Paths":    paths,
									"NotPaths": notPaths,
								})),
							})),
						),
					})),
				),
			}),
		})),
	}))
}

var getAudienceMatcher = func(action string, hashLabelValue string, indexLabelValue string, audiences []string) types.GomegaMatcher {
	var audiencesMatchers []types.GomegaMatcher

	for _, audience := range audiences {
		m := PointTo(MatchFields(IgnoreExtras, Fields{
			"Key":    Equal("request.auth.claims[aud]"),
			"Values": ContainElement(audience),
		}))
		audiencesMatchers = append(audiencesMatchers, m)
	}

	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Action": WithTransform(ActionToString, Equal(action)),
		"Obj": PointTo(MatchFields(IgnoreExtras, Fields{
			"ObjectMeta": MatchFields(IgnoreExtras, Fields{
				"Labels": And(
					HaveKeyWithValue("gateway.kyma-project.io/index", indexLabelValue),
					HaveKeyWithValue("gateway.kyma-project.io/hash", hashLabelValue),
				),
			}),
			"Spec": MatchFields(IgnoreExtras, Fields{
				"Rules": ContainElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"When": ContainElements(audiencesMatchers),
					})),
				),
			}),
		})),
	}))
}

var ActionToString = func(a processing.Action) string { return a.String() }

func getFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = rulev1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = securityv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}
