package default_domain

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"os"
	"testing"

	apinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestDefaultDomain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Default Domain Suite")
}

var _ = ReportAfterSuite("custom reporter", func(report types.Report) {
	logger := zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))

	if key, ok := os.LookupEnv("ARTIFACTS"); ok {
		reportsFilename := fmt.Sprintf("%s/%s", key, "junit-processing.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	} else {
		if err := os.MkdirAll("../../reports", 0755); err != nil {
			logger.Error(err, "could not create directory")
		}

		reportsFilename := fmt.Sprintf("%s/%s", "../../reports", "junit-processing.xml")
		logger.Info("Generating reports at", "location", reportsFilename)
		err := reporters.GenerateJUnitReport(report, reportsFilename)

		if err != nil {
			logger.Error(err, "Junit Report Generation Error")
		}
	}
})

var _ = Describe("GetDomainFromKymaGateway", func() {
	It("should get domain from default kyma gateway if it exists", func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: kymaGatewayName, Namespace: kymaGatewayNamespace},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTPS"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
				},
			},
		}

		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromKymaGateway(context.Background(), client)

		// then
		Expect(err).ShouldNot(HaveOccurred())
		Expect(host).To(Equal("local.kyma.dev"))
	})

	It("should return error if gateway does not have an HTTPS server", func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: kymaGatewayName, Namespace: kymaGatewayNamespace},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromKymaGateway(context.Background(), client)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("gateway must have a single https server definition, num=0"))
		Expect(host).To(Equal(""))
	})

	It("should return error if gateway does not have an HTTPS server when gateway has multiple servers", func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: kymaGatewayName, Namespace: kymaGatewayNamespace},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
					},
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromKymaGateway(context.Background(), client)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("gateway must have a single https server definition, num=0"))
		Expect(host).To(Equal(""))
	})

	It(`should return error if gateway has a HTTPS server but host do not start with "*." prefix`, func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: kymaGatewayName, Namespace: kymaGatewayNamespace},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTPS"},
						Hosts: []string{
							"local.kyma.dev",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromKymaGateway(context.Background(), client)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal(`gateway https server host "local.kyma.dev" does not start with the prefix "*."`))
		Expect(host).To(Equal(""))
	})

	It(`should return error if gateway has a HTTPS server but host do not define domain after "*." prefix`, func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: kymaGatewayName, Namespace: kymaGatewayNamespace},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTPS"},
						Hosts: []string{
							"*.",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromKymaGateway(context.Background(), client)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal(`gateway https server host "*." does not define domain after the prefix "*."`))
		Expect(host).To(Equal(""))
	})

	It("should return empty domain and not found error if gateway does not exist", func() {
		// given
		client := getFakeClient()

		// when
		host, err := GetDomainFromKymaGateway(context.Background(), client)

		// then
		Expect(err).Should(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())
		Expect(host).To(Equal(""))
	})
})

var _ = Describe("GetDomainFromGateway", func() {
	It("should get domain from gateway if it exists", func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-namespace"},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
				},
			},
		}

		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromGateway(context.Background(), client, "gateway-name", "gateway-namespace")

		// then
		Expect(err).ShouldNot(HaveOccurred())
		Expect(host).To(Equal("local.kyma.dev"))
	})

	It("should return error if gateway defines more than a single host", func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-namespace"},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
							"*.remote.kyma.dev",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromGateway(context.Background(), client, "gateway-name", "gateway-namespace")

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("gateway must specify server(s) with the same single host"))
		Expect(host).To(Equal(""))
	})

	It("should return error if gateway defines servers with different hosts", func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-namespace"},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTP"},
						Hosts: []string{
							"*.local.kyma.dev",
						},
					},
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTPS"},
						Hosts: []string{
							"*.remote.kyma.dev",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromGateway(context.Background(), client, "gateway-name", "gateway-namespace")

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal("gateway must specify server(s) with the same single host"))
		Expect(host).To(Equal(""))
	})

	It(`should return error if gateway has a HTTPS server but host do not start with "*." prefix`, func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-namespace"},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTPS"},
						Hosts: []string{
							"local.kyma.dev",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromGateway(context.Background(), client, "gateway-name", "gateway-namespace")

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal(`gateway server host "local.kyma.dev" does not start with the prefix "*."`))
		Expect(host).To(Equal(""))
	})

	It(`should return error if gateway has a HTTPS server but host do not define domain after "*." prefix`, func() {
		// given
		gateway := networkingv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-namespace"},
			Spec: apinetworkingv1beta1.Gateway{
				Servers: []*apinetworkingv1beta1.Server{
					{
						Port: &apinetworkingv1beta1.Port{Protocol: "HTTPS"},
						Hosts: []string{
							"*.",
						},
					},
				},
			},
		}
		client := getFakeClient(&gateway)

		// when
		host, err := GetDomainFromGateway(context.Background(), client, "gateway-name", "gateway-namespace")

		// then
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(Equal(`gateway server host "*." does not define domain after the prefix "*."`))
		Expect(host).To(Equal(""))
	})

	It("should return empty domain and not found error if gateway does not exist", func() {
		// given
		client := getFakeClient()

		// when
		host, err := GetDomainFromGateway(context.Background(), client, "gateway-name", "gateway-namespace")

		// then
		Expect(err).Should(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())
		Expect(host).To(Equal(""))
	})
})

func getFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	err := networkingv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}
