module github.com/kyma-incubator/api-gateway

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.2
	github.com/ory/oathkeeper-maester v0.1.0
	github.com/pkg/errors v0.9.0
	istio.io/api v0.0.0-20200812202721-24be265d41c3
	istio.io/client-go v0.0.0-20200916161914-94f0e83444ca
	k8s.io/api v0.18.15
	k8s.io/apimachinery v0.18.15
	k8s.io/client-go v0.18.15
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/onsi/ginkgo v1.14.0 => github.com/onsi/ginkgo v1.12.1
